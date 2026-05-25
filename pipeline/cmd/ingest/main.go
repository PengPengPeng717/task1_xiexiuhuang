package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"mining-pipeline/internal/chunker"
	"mining-pipeline/internal/dedup"
	"mining-pipeline/internal/embedder"
	"mining-pipeline/internal/model"
	"mining-pipeline/internal/store/jsonl"
	"mining-pipeline/pipeline/collector/news"
	"mining-pipeline/pipeline/collector/policy"
	"mining-pipeline/pipeline/collector/price"
	"mining-pipeline/pipeline/fixtures"
	"mining-pipeline/pipeline/processor"
)

func main() {
	dataDir := flag.String("data-dir", "data", "data directory")
	seed := flag.Bool("seed", false, "use synthetic seed corpus (P0 demo)")
	seedP1 := flag.Bool("seed-p1", false, "synthetic 200 per category for manifest testing")
	maxNews := flag.Int("max-news-per-feed", 15, "max RSS items per news feed")
	flag.Parse()

	now := time.Now().UTC()
	docsPath := filepath.Join(*dataDir, "documents.jsonl")
	chunksPath := filepath.Join(*dataDir, "chunks.jsonl")
	manifestPath := filepath.Join(*dataDir, "manifest.json")

	var collected []model.Document
	var failed []string

	if *seed || *seedP1 {
		if *seedP1 {
			collected = fixtures.SeedP1(now)
		} else {
			collected = fixtures.SeedDocuments(now)
		}
	} else {
		newsDocs, w, err := news.Collect(now, *maxNews)
		if err != nil {
			log.Printf("news: %v", err)
		}
		collected = append(collected, newsDocs...)
		failed = append(failed, w...)

		policyDocs, w, perr := policy.Collect(now)
		if perr != nil {
			log.Printf("policy: %v", perr)
		}
		collected = append(collected, policyDocs...)
		failed = append(failed, w...)

		priceDocs, pf := price.Collect(now, 30)
		collected = append(collected, priceDocs...)
		failed = append(failed, pf...)
	}

	existing, _ := jsonl.LoadDocuments(docsPath)
	set := dedup.NewSet()
	for _, d := range existing {
		set.Add(d)
	}
	var merged []model.Document
	merged = append(merged, existing...)
	for _, d := range collected {
		if !processor.WithinLast30Days(d.PublishedAt, now) {
			continue
		}
		if set.Add(d) {
			merged = append(merged, d)
		}
	}

	emb := embedder.NewFromEnv()
	var allChunks []model.Chunk
	for _, d := range merged {
		parts := chunker.ChunkDocument(d, 800)
		embedFn := func(text string) []float32 {
			v, err := emb.Embed(text)
			if err != nil {
				log.Printf("embed %s: %v", d.DocumentID, err)
				return embedder.PlaceholderEmbed(text)
			}
			return v
		}
		allChunks = append(allChunks, chunker.BuildChunks(d, parts, embedFn)...)
	}

	if err := os.MkdirAll(*dataDir, 0o755); err != nil {
		log.Fatal(err)
	}
	if err := jsonl.WriteDocuments(docsPath, merged); err != nil {
		log.Fatal(err)
	}
	if err := jsonl.WriteChunks(chunksPath, allChunks); err != nil {
		log.Fatal(err)
	}
	if err := jsonl.SaveManifest(manifestPath, merged, failed, emb.Mode()); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("ingest done: documents=%d chunks=%d failed_sources=%d mode=%s\n",
		len(merged), len(allChunks), len(failed), emb.Mode())
}
