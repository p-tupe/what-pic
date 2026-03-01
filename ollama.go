package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Provider interface {
	Analyze(ctx context.Context, img []byte, mimeType, prompt string) (string, error)
}

type Ollama struct {
	Host  string
	Model string
}

func (o *Ollama) Analyze(ctx context.Context, img []byte, _ string, prompt string) (string, error) {
	body, _ := json.Marshal(map[string]any{
		"model":  o.Model,
		"prompt": prompt,
		"images": []string{base64.StdEncoding.EncodeToString(img)},
		"stream": false,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.Host+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			return "", fmt.Errorf("cannot connect to Ollama at %s — is it running?", o.Host)
		}
		return "", fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// error body may still carry a JSON message
		var e struct {
			Error string `json:"error"`
		}
		if json.NewDecoder(resp.Body).Decode(&e) == nil && e.Error != "" {
			return "", fmt.Errorf("ollama: %s", e.Error)
		}
		return "", fmt.Errorf("ollama: HTTP %d", resp.StatusCode)
	}

	var result struct {
		Response string `json:"response"`
		Error    string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("ollama: failed to decode response: %w", err)
	}
	if result.Error != "" {
		return "", fmt.Errorf("ollama: %s", result.Error)
	}
	if result.Response == "" {
		return "", fmt.Errorf("ollama: empty response — is %q a vision-capable model?", o.Model)
	}
	return result.Response, nil
}
