package retriever

import (
	"math"
	"sort"
	"strings"
	"time"

	"mining-pipeline/internal/embedder"
	"mining-pipeline/internal/model"
	"mining-pipeline/internal/queryparse"
	"mining-pipeline/pipeline/processor"
)

type Hit struct {
	Chunk model.Chunk
	Score float32
}

type Retriever struct {
	Chunks []model.Chunk
	Now    time.Time
	Embed  *embedder.Client
}

func New(chunks []model.Chunk, embed *embedder.Client) *Retriever {
	return &Retriever{Chunks: chunks, Now: time.Now().UTC(), Embed: embed}
}

func (r *Retriever) Search(question string, k int) ([]Hit, queryparse.Parsed, string) {
	if k <= 0 {
		k = 5
	}
	p := queryparse.Parse(question)
	qvec, _ := r.Embed.Embed(question)

	var candidates []model.Chunk
	for _, c := range r.Chunks {
		if p.SourceType != "" && c.SourceType != p.SourceType {
			continue
		}
		if p.Days > 0 && !processor.WithinLastNDays(c.PublishedAt, r.Now, p.Days) {
			continue
		}
		if p.Commodity != "" && !strings.Contains(c.Text+c.SourceURL, p.Commodity) {
			continue
		}
		if p.Region != "" && !strings.Contains(strings.ToLower(c.Text+c.SourceURL), strings.ToLower(p.Region)) {
			continue
		}
		candidates = append(candidates, c)
	}

	if len(candidates) == 0 {
		if p.Days > 0 {
			return nil, p, "no_recent"
		}
		return nil, p, "no_data"
	}

	var hits []Hit
	for _, c := range candidates {
		score := cosine(qvec, c.Embedding)
		if score < 0.01 && len(p.Keywords) > 0 {
			for _, kw := range p.Keywords {
				if strings.Contains(c.Text, kw) {
					score += 0.2
				}
			}
		}
		hits = append(hits, Hit{Chunk: c, Score: score})
	}
	sort.Slice(hits, func(i, j int) bool { return hits[i].Score > hits[j].Score })
	if len(hits) > k {
		hits = hits[:k]
	}
	if len(hits) == 0 || hits[0].Score < 0.05 {
		return nil, p, "no_data"
	}
	return hits, p, "ok"
}

func cosine(a, b []float32) float32 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dot, na, nb float32
	for i := 0; i < n; i++ {
		dot += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (float32(math.Sqrt(float64(na))) * float32(math.Sqrt(float64(nb))))
}
