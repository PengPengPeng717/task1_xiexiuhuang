package retriever_test

import (
	"testing"
	"time"

	"mining-pipeline/internal/embedder"
	"mining-pipeline/internal/model"
	"mining-pipeline/internal/retriever"
)

func TestSearchFindsDocument(t *testing.T) {
	emb := embedder.NewFromEnv()
	text := "澳洲锂出口政策调整"
	vec, _ := emb.Embed(text)
	chunks := []model.Chunk{{
		ChunkID: "c1", DocumentID: "doc-1", Text: text,
		PublishedAt: time.Now().UTC(), SourceType: model.SourcePolicy,
		SourceURL: "https://example.com/p1", Embedding: vec,
	}}
	r := retriever.New(chunks, emb)
	hits, _, cov := r.Search("近7天澳洲锂政策", 5)
	if cov != "ok" || len(hits) == 0 {
		t.Fatalf("coverage=%s hits=%d", cov, len(hits))
	}
}
