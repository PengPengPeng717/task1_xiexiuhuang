package jsonl

import (
	"encoding/json"
	"os"
	"time"

	"mining-pipeline/internal/model"
)

func LoadManifest(path string) (model.Manifest, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return model.Manifest{FailedSources: []string{}, EmbeddingMode: "placeholder"}, nil
		}
		return model.Manifest{}, err
	}
	var m model.Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		return model.Manifest{}, err
	}
	if m.FailedSources == nil {
		m.FailedSources = []string{}
	}
	return m, nil
}

func SaveManifest(path string, docs []model.Document, failed []string, embeddingMode string) error {
	var news, policy, price int
	for _, d := range docs {
		switch d.SourceType {
		case model.SourceNews:
			news++
		case model.SourcePolicy:
			policy++
		case model.SourcePrice:
			price++
		}
	}
	m := model.Manifest{
		NewsCount:     news,
		PolicyCount:   policy,
		PriceCount:    price,
		TotalCount:    len(docs),
		LastIngestAt:  time.Now().UTC(),
		FailedSources: failed,
		EmbeddingMode: embeddingMode,
	}
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return WriteFileAtomic(path, append(b, '\n'))
}
