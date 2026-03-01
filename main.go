package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	port := getenv("PORT", "8080")
	maxMBStr := getenv("MAX_UPLOAD_MB", "50")
	maxMB, _ := strconv.ParseInt(maxMBStr, 10, 64)
	if maxMB <= 0 {
		maxMB = 50
	}

	p := &Ollama{
		Host:  getenv("OLLAMA_HOST", "http://localhost:11434"),
		Model: getenv("OLLAMA_MODEL", "llava"),
	}

	var rl *rateLimiter
	if getenv("RATE_LIMIT_IMAGES", "") != "" {
		trustProxy := getenv("TRUST_PROXY", "") != ""
		rl = newRateLimiter(trustProxy)
		log.Printf("rate limiting enabled: %d images/hour/IP (trust_proxy=%v)", rateLimit, trustProxy)
	}

	mux := http.NewServeMux()
	mux.Handle("GET /", http.FileServer(http.Dir("static")))
	mux.HandleFunc("POST /analyze", analyzeHandler(p, maxMB, rl))

	addr := fmt.Sprintf(":%s", port)
	log.Printf("what-pic listening on %s (ollama=%s model=%s)", addr, p.Host, p.Model)
	log.Fatal(http.ListenAndServe(addr, mux))
}
