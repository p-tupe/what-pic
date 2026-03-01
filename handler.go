package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"strings"
	"time"
)

const analyzeTimeout = 5 * time.Minute

func analyzeHandler(p Provider, maxMB int64, rl *rateLimiter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if err := r.ParseMultipartForm(maxMB << 20); err != nil {
			if strings.Contains(err.Error(), "too large") {
				http.Error(w, fmt.Sprintf("Upload too large — maximum is %d MB.", maxMB), http.StatusRequestEntityTooLarge)
			} else {
				http.Error(w, "Could not parse upload: "+err.Error(), http.StatusBadRequest)
			}
			return
		}

		files := r.MultipartForm.File["files[]"]
		if len(files) == 0 {
			http.Error(w, "No files received. Make sure you selected at least one image.", http.StatusBadRequest)
			return
		}

		if rl != nil {
			ok, remaining := rl.allow(rl.realIP(r), len(files))
			if !ok {
				w.Header().Set("Retry-After", "3600")
				http.Error(w,
					fmt.Sprintf("Rate limit reached. You have %d image(s) remaining in the current hour.", remaining),
					http.StatusTooManyRequests)
				return
			}
		}

		prompt := r.FormValue("prompt")
		if prompt == "" {
			prompt = "Describe this image."
		}
		if len(prompt) > 4000 {
			http.Error(w, "Prompt too long — maximum is 4000 characters.", http.StatusBadRequest)
			return
		}
		format := r.FormValue("format")
		if format == "" {
			format = "text"
		}

		ctx, cancel := context.WithTimeout(r.Context(), analyzeTimeout)
		defer cancel()

		var results []Result
		for _, fh := range files {
			f, err := fh.Open()
			if err != nil {
				http.Error(w, "Could not open "+fh.Filename+": "+err.Error(), http.StatusInternalServerError)
				return
			}
			data, err := io.ReadAll(f)
			f.Close()
			if err != nil {
				http.Error(w, "Could not read "+fh.Filename+": "+err.Error(), http.StatusInternalServerError)
				return
			}

			mimeType := fh.Header.Get("Content-Type")
			if mimeType == "" {
				mimeType = http.DetectContentType(data)
			}
			mt, _, _ := mime.ParseMediaType(mimeType)
			if mt != "" {
				mimeType = mt
			}
			if !strings.HasPrefix(mimeType, "image/") {
				results = append(results, Result{
					File:   fh.Filename,
					Result: fmt.Sprintf("skipped: not an image (detected %q)", mimeType),
				})
				continue
			}

			exifData := extractEXIF(data)
			fullPrompt := prompt + exifSummary(exifData)

			out, err := p.Analyze(ctx, data, mimeType, fullPrompt)
			if err != nil {
				// surface provider errors clearly rather than burying them
				results = append(results, Result{
					File:   fh.Filename,
					Result: "error: " + err.Error(),
				})
				continue
			}

			results = append(results, Result{File: fh.Filename, Result: out})
		}

		body, contentType, err := formatResults(results, format)
		if err != nil {
			http.Error(w, "Could not format results: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Disposition",
			fmt.Sprintf(`attachment; filename="results.%s"`, fileExtension(format)))
		if _, err := w.Write(body); err != nil {
			log.Printf("write error: %v", err)
		}
	}
}
