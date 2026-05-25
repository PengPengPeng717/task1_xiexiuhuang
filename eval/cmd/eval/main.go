package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"mining-pipeline/internal/embedder"
	"mining-pipeline/internal/retriever"
	"mining-pipeline/internal/store/jsonl"
	"mining-pipeline/eval/metrics"
)

type gtEntry struct {
	ID                  string   `json:"id"`
	Question            string   `json:"question"`
	ExpectedDocumentIDs []string `json:"expected_document_ids"`
	ExpectedPoints      []string `json:"expected_points"`
	Category            []string `json:"category"`
}

type report struct {
	RecallAt5        float64            `json:"recall_at_5"`
	FaithfulnessRate float64            `json:"faithfulness_rate"`
	PerQuestion      []questionResult   `json:"per_question"`
}

type questionResult struct {
	ID            string `json:"id"`
	RecallHit     bool   `json:"recall_hit"`
	Faithful      bool   `json:"faithful"`
	Coverage      string `json:"coverage"`
}

func main() {
	dataDir := flag.String("data-dir", "data", "data directory")
	gtPath := flag.String("ground-truth", "eval/ground_truth.json", "ground truth file")
	out := flag.String("out", "eval/report.json", "report output")
	flag.Parse()

	b, err := os.ReadFile(*gtPath)
	if err != nil {
		log.Fatal(err)
	}
	var entries []gtEntry
	if err := json.Unmarshal(b, &entries); err != nil {
		log.Fatal(err)
	}
	chunks, err := jsonl.LoadChunks(filepath.Join(*dataDir, "chunks.jsonl"))
	if err != nil {
		log.Fatal(err)
	}
	emb := embedder.NewFromEnv()
	rtr := retriever.New(chunks, emb)

	var recallHits, faithHits int
	var results []questionResult
	for _, e := range entries {
		hits, _, cov := rtr.Search(e.Question, 5)
		rh := metrics.RecallAt5(hits, e.ExpectedDocumentIDs, e.ExpectedPoints)
		fh := metrics.Faithfulness(hits, e.ExpectedPoints)
		if rh {
			recallHits++
		}
		if fh {
			faithHits++
		}
		results = append(results, questionResult{ID: e.ID, RecallHit: rh, Faithful: fh, Coverage: cov})
	}
	rep := report{
		RecallAt5:        float64(recallHits) / float64(len(entries)),
		FaithfulnessRate: float64(faithHits) / float64(len(entries)),
		PerQuestion:      results,
	}
	outB, _ := json.MarshalIndent(rep, "", "  ")
	if err := os.WriteFile(*out, append(outB, '\n'), 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("eval: recall@5=%.2f faithfulness=%.2f report=%s\n", rep.RecallAt5, rep.FaithfulnessRate, *out)
}
