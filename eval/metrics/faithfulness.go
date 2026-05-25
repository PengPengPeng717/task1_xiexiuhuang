package metrics

import (
	"strings"

	"mining-pipeline/internal/retriever"
)

func Faithfulness(hits []retriever.Hit, expectedPoints []string) bool {
	if len(expectedPoints) == 0 {
		return len(hits) > 0
	}
	var combined strings.Builder
	for _, h := range hits {
		combined.WriteString(h.Chunk.Text)
		combined.WriteString(h.Chunk.SourceURL)
	}
	text := combined.String()
	for _, p := range expectedPoints {
		if p != "" && strings.Contains(text, p) {
			return true
		}
	}
	return false
}
