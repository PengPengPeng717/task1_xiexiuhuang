package chunker

import (
	"fmt"
	"strings"

	"mining-pipeline/internal/model"
)

const defaultChunkSize = 800

func ChunkDocument(d model.Document, size int) []string {
	if size <= 0 {
		size = defaultChunkSize
	}
	text := strings.TrimSpace(d.ContentText)
	if d.SourceType == model.SourcePrice || len(text) <= size {
		if text == "" {
			text = d.Title
		}
		return []string{text}
	}
	var parts []string
	for len(text) > 0 {
		if len(text) <= size {
			parts = append(parts, text)
			break
		}
		cut := size
		if idx := strings.LastIndex(text[:cut], "\n\n"); idx > size/2 {
			cut = idx
		}
		parts = append(parts, strings.TrimSpace(text[:cut]))
		text = strings.TrimSpace(text[cut:])
	}
	return parts
}

func BuildChunks(d model.Document, parts []string, embed func(string) []float32) []model.Chunk {
	var chunks []model.Chunk
	for i, p := range parts {
		if strings.TrimSpace(p) == "" {
			continue
		}
		chunks = append(chunks, model.Chunk{
			ChunkID:     fmt.Sprintf("%s-%d", d.DocumentID, i),
			DocumentID:  d.DocumentID,
			ChunkIndex:  i,
			Text:        p,
			PublishedAt: d.PublishedAt,
			SourceType:  d.SourceType,
			SourceURL:   d.CanonicalURL,
			Embedding:   embed(p),
		})
	}
	return chunks
}
