package embedder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"os"
	"time"
)

type Client struct {
	URL    string
	Model  string
	APIKey string
	HTTP   *http.Client
}

func NewFromEnv() *Client {
	return &Client{
		URL:    os.Getenv("EMBEDDING_URL"),
		Model:  envOr("EMBEDDING_MODEL", "text-embedding-3-small"),
		APIKey: os.Getenv("EMBEDDING_API_KEY"),
		HTTP:   &http.Client{Timeout: 60 * time.Second},
	}
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func (c *Client) Mode() string {
	if c.URL == "" {
		return "placeholder"
	}
	return "real"
}

func (c *Client) Embed(text string) ([]float32, error) {
	if c.URL == "" {
		return PlaceholderEmbed(text), nil
	}
	body, _ := json.Marshal(map[string]any{
		"input": text,
		"model": c.Model,
	})
	req, err := http.NewRequest(http.MethodPost, c.URL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("embedding api %d: %s", resp.StatusCode, string(b))
	}
	var parsed struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}
	if len(parsed.Data) == 0 {
		return nil, fmt.Errorf("empty embedding response")
	}
	out := make([]float32, len(parsed.Data[0].Embedding))
	for i, v := range parsed.Data[0].Embedding {
		out[i] = float32(v)
	}
	return out, nil
}

func PlaceholderEmbed(text string) []float32 {
	const dim = 64
	vec := make([]float32, dim)
	h := fnv.New64a()
	h.Write([]byte(text))
	seed := h.Sum64()
	for i := 0; i < dim; i++ {
		seed = seed*6364136223846793005 + 1
		vec[i] = float32(int(seed%1000))/1000.0 - 0.5
	}
	var norm float32
	for _, v := range vec {
		norm += v * v
	}
	if norm == 0 {
		return vec
	}
	norm = float32(1.0 / (float64(norm) + 1e-8))
	for i := range vec {
		vec[i] *= norm
	}
	return vec
}
