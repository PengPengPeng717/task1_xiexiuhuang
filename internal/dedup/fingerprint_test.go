package dedup_test

import (
	"testing"

	"mining-pipeline/internal/dedup"
	"mining-pipeline/internal/model"
)

func TestDedupSet(t *testing.T) {
	s := dedup.NewSet()
	d := model.Document{SourceType: model.SourceNews, CanonicalURL: "https://a.com/1", ContentSHA256: "x"}
	if !s.Add(d) {
		t.Fatal("first add")
	}
	if s.Add(d) {
		t.Fatal("duplicate should be rejected")
	}
}
