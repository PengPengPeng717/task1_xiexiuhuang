package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"mining-pipeline/internal/embedder"
	"mining-pipeline/internal/model"
	"mining-pipeline/internal/store/jsonl"
	"mining-pipeline/serve/handler"
)

func TestQueryHandler(t *testing.T) {
	dir := t.TempDir()
	text := "近7天澳洲锂出口政策调整监测"
	vec := embedder.PlaceholderEmbed(text)
	chunks := []model.Chunk{{
		ChunkID: "c1", DocumentID: "p1", Text: text,
		PublishedAt: time.Now().UTC(), SourceType: model.SourcePolicy,
		SourceURL: "https://example.com/p1", Embedding: vec,
	}}
	if err := jsonl.WriteChunks(dir+"/chunks.jsonl", chunks); err != nil {
		t.Fatal(err)
	}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := &handler.QueryHandler{DataDir: dir, Embed: embedder.NewFromEnv()}
	h.Register(r)
	req := httptest.NewRequest(http.MethodGet, "/query?question="+text, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	var resp struct {
		Coverage string `json:"coverage"`
		Hits     []any  `json:"hits"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Coverage != "ok" || len(resp.Hits) == 0 {
		t.Fatalf("coverage=%s hits=%v", resp.Coverage, resp.Hits)
	}
}
