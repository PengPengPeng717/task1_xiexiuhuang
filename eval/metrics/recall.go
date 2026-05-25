package metrics

import (
	"mining-pipeline/internal/retriever"
)

func RecallAt5(hits []retriever.Hit, expectedDocIDs []string, expectedPoints []string) bool {
	set := make(map[string]struct{}, len(expectedDocIDs))
	for _, id := range expectedDocIDs {
		set[id] = struct{}{}
	}
	for _, h := range hits {
		if _, ok := set[h.Chunk.DocumentID]; ok {
			return true
		}
	}
	if len(expectedPoints) > 0 && Faithfulness(hits, expectedPoints) {
		return true
	}
	return false
}
