package queryparse_test

import (
	"testing"

	"mining-pipeline/internal/model"
	"mining-pipeline/internal/queryparse"
)

func TestParsePolicyWeek(t *testing.T) {
	p := queryparse.Parse("近 7 天澳洲锂出口政策有何变化？")
	if p.Days != 7 {
		t.Fatalf("days=%d", p.Days)
	}
	if p.SourceType != model.SourcePolicy {
		t.Fatalf("source=%s", p.SourceType)
	}
	if p.Commodity != "锂" {
		t.Fatalf("commodity=%s", p.Commodity)
	}
}
