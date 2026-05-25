package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"mining-pipeline/internal/embedder"
	"mining-pipeline/internal/retriever"
	"mining-pipeline/internal/store/jsonl"
)

const disclaimer = "以下为知识库原文摘录，不构成投资或政策分析建议。"

type QueryHandler struct {
	DataDir string
	Embed   *embedder.Client
}

type hitDTO struct {
	Text        string  `json:"text"`
	SourceURL   string  `json:"source_url"`
	PublishedAt string  `json:"published_at"`
	SourceType  string  `json:"source_type"`
	Score       float32 `json:"score"`
}

type queryResponse struct {
	Disclaimer string   `json:"disclaimer"`
	Hits       []hitDTO `json:"hits"`
	Coverage   string   `json:"coverage"`
}

func (h *QueryHandler) Register(r *gin.Engine) {
	r.GET("/query", h.handle)
	r.POST("/query", h.handle)
}

func (h *QueryHandler) handle(c *gin.Context) {
	q := c.Query("question")
	if q == "" {
		var body struct {
			Question string `json:"question"`
		}
		_ = c.ShouldBindJSON(&body)
		q = body.Question
	}
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "question is required"})
		return
	}
	chunks, err := jsonl.LoadChunks(h.DataDir + "/chunks.jsonl")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	rtr := retriever.New(chunks, h.Embed)
	hits, _, coverage := rtr.Search(q, 5)
	resp := queryResponse{Disclaimer: disclaimer, Hits: []hitDTO{}, Coverage: coverage}
	for _, hit := range hits {
		resp.Hits = append(resp.Hits, hitDTO{
			Text:        hit.Chunk.Text,
			SourceURL:   hit.Chunk.SourceURL,
			PublishedAt: hit.Chunk.PublishedAt.Format(time.RFC3339),
			SourceType:  string(hit.Chunk.SourceType),
			Score:       hit.Score,
		})
	}
	c.JSON(http.StatusOK, resp)
}
