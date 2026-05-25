package model_test

import (
	"encoding/json"
	"testing"
	"time"

	"mining-pipeline/internal/model"
)

func TestDocumentJSONRoundTrip(t *testing.T) {
	doc := model.Document{
		DocumentID:    "news-1",
		SourceType:    model.SourceNews,
		CanonicalURL:  "https://www.mining.com/article/1",
		Title:         "Test",
		PublishedAt:   time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC),
		ContentText:   "body",
		ContentSHA256: "abc",
	}
	b, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	var got model.Document
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got.DocumentID != doc.DocumentID {
		t.Fatalf("got %s", got.DocumentID)
	}
}
