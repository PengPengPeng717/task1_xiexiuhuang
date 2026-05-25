package model

import "time"

type Document struct {
	DocumentID    string     `json:"document_id"`
	SourceType    SourceType `json:"source_type"`
	CanonicalURL  string     `json:"canonical_url"`
	Title         string     `json:"title"`
	PublishedAt   time.Time  `json:"published_at"`
	ContentText   string     `json:"content_text"`
	ContentSHA256 string     `json:"content_sha256"`
	Region        string     `json:"region,omitempty"`
	Commodity     string     `json:"commodity,omitempty"`
	Exchange      string     `json:"exchange,omitempty"`
	TradeDate     string     `json:"trade_date,omitempty"`
}

type Chunk struct {
	ChunkID     string     `json:"chunk_id"`
	DocumentID  string     `json:"document_id"`
	ChunkIndex  int        `json:"chunk_index"`
	Text        string     `json:"text"`
	PublishedAt time.Time  `json:"published_at"`
	SourceType  SourceType `json:"source_type"`
	SourceURL   string     `json:"source_url"`
	Embedding   []float32  `json:"embedding"`
}

type Manifest struct {
	NewsCount     int       `json:"news_count"`
	PolicyCount   int       `json:"policy_count"`
	PriceCount    int       `json:"price_count"`
	TotalCount    int       `json:"total_count"`
	LastIngestAt  time.Time `json:"last_ingest_at"`
	FailedSources []string  `json:"failed_sources"`
	EmbeddingMode string    `json:"embedding_mode"`
}
