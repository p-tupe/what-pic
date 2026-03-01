package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type Result struct {
	File   string `json:"file" yaml:"file"`
	Result string `json:"result" yaml:"result"`
}

func formatResults(results []Result, format string) ([]byte, string, error) {
	switch format {
	case "json":
		b, err := json.MarshalIndent(results, "", "  ")
		return b, "application/json", err

	case "yaml":
		b, err := yaml.Marshal(results)
		return b, "application/yaml", err

	case "csv":
		var buf bytes.Buffer
		w := csv.NewWriter(&buf)
		_ = w.Write([]string{"file", "result"})
		for _, r := range results {
			_ = w.Write([]string{r.File, r.Result})
		}
		w.Flush()
		return buf.Bytes(), "text/csv", w.Error()

	default: // text
		var sb strings.Builder
		for _, r := range results {
			fmt.Fprintf(&sb, "=== %s ===\n%s\n\n", r.File, r.Result)
		}
		return []byte(sb.String()), "text/plain", nil
	}
}

func fileExtension(format string) string {
	switch format {
	case "json":
		return "json"
	case "yaml":
		return "yaml"
	case "csv":
		return "csv"
	default:
		return "txt"
	}
}
