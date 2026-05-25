package fixtures

import (
	"fmt"
	"time"

	"mining-pipeline/internal/dedup"
	"mining-pipeline/internal/model"
)

// SeedDocuments generates demo corpus for P0 when live collectors are insufficient.
// Clearly synthetic titles; used only with -seed flag documented in README.
func SeedDocuments(now time.Time) []model.Document {
	var docs []model.Document
	for i := 0; i < 30; i++ {
		pub := now.AddDate(0, 0, -(i % 28))
		body := fmt.Sprintf("Mining industry update sample %d about copper and lithium markets.", i)
		docs = append(docs, model.Document{
			DocumentID:    fmt.Sprintf("seed-news-%d", i),
			SourceType:    model.SourceNews,
			CanonicalURL:  fmt.Sprintf("https://example.com/mining/news/%d", i),
			Title:         fmt.Sprintf("Mining News Sample %d", i),
			PublishedAt:   pub,
			ContentText:   body,
			ContentSHA256: dedup.ContentSHA256(body),
		})
	}
	for i := 0; i < 30; i++ {
		pub := now.AddDate(0, 0, -(i % 25))
		body := fmt.Sprintf("稀土与锂出口政策公告样本 %d，涉及澳洲与中国监管动态。", i)
		docs = append(docs, model.Document{
			DocumentID:    fmt.Sprintf("seed-policy-%d", i),
			SourceType:    model.SourcePolicy,
			CanonicalURL:  fmt.Sprintf("https://example.com/policy/%d", i),
			Title:         fmt.Sprintf("政策公告样本 %d", i),
			PublishedAt:   pub,
			ContentText:   body,
			ContentSHA256: dedup.ContentSHA256(body),
			Region:        "澳洲",
			Commodity:     "锂",
		})
	}
	for i := 0; i < 30; i++ {
		pub := now.AddDate(0, 0, -(i % 20))
		td := pub.Format("2006-01-02")
		commodities := []string{"铜", "锌", "镍", "锂", "铁矿石"}
		exchanges := []string{"LME", "LME", "LME", "SHFE", "Mysteel"}
		c := commodities[i%len(commodities)]
		ex := exchanges[i%len(exchanges)]
		body := fmt.Sprintf("%s %s 收盘价样本 %d: 1000+%d", ex, c, i, i)
		docs = append(docs, model.Document{
			DocumentID:    fmt.Sprintf("seed-price-%d", i),
			SourceType:    model.SourcePrice,
			CanonicalURL:  fmt.Sprintf("https://example.com/price/%s/%s", ex, td),
			Title:         fmt.Sprintf("%s %s %s", ex, c, td),
			PublishedAt:   pub,
			ContentText:   body,
			ContentSHA256: dedup.ContentSHA256(body),
			Commodity:     c,
			Exchange:      ex,
			TradeDate:     td,
		})
	}
	return docs
}

// SeedP1 generates exactly 200 documents per category (600 total) for manifest testing.
func SeedP1(now time.Time) []model.Document {
	var out []model.Document
	for n := 0; n < 200; n++ {
		pub := now.AddDate(0, 0, -(n % 29))
		body := fmt.Sprintf("Extended mining news record %d for volume testing copper lithium.", n)
		out = append(out, model.Document{
			DocumentID:    fmt.Sprintf("seed-news-%d", n),
			SourceType:    model.SourceNews,
			CanonicalURL:  fmt.Sprintf("https://example.com/mining/news/%d", n),
			Title:         fmt.Sprintf("Mining News %d", n),
			PublishedAt:   pub,
			ContentText:   body,
			ContentSHA256: dedup.ContentSHA256(body),
		})
	}
	for n := 0; n < 200; n++ {
		pub := now.AddDate(0, 0, -(n % 28))
		body := fmt.Sprintf("政策样本 %d：近7天澳洲锂出口政策监测文本。", n)
		out = append(out, model.Document{
			DocumentID:    fmt.Sprintf("seed-policy-%d", n),
			SourceType:    model.SourcePolicy,
			CanonicalURL:  fmt.Sprintf("https://example.com/policy/%d", n),
			Title:         fmt.Sprintf("政策 %d", n),
			PublishedAt:   pub,
			ContentText:   body,
			ContentSHA256: dedup.ContentSHA256(body),
			Region:        "澳洲",
			Commodity:     "锂",
		})
	}
	for n := 0; n < 200; n++ {
		pub := now.AddDate(0, 0, -(n % 27))
		td := pub.Format("2006-01-02")
		c := []string{"铜", "锌", "镍", "锂", "铁矿石"}[n%5]
		ex := []string{"LME", "LME", "LME", "SHFE", "Mysteel"}[n%5]
		body := fmt.Sprintf("%s %s 行情 %d", ex, c, n)
		out = append(out, model.Document{
			DocumentID:    fmt.Sprintf("seed-price-%d", n),
			SourceType:    model.SourcePrice,
			CanonicalURL:  fmt.Sprintf("https://example.com/price/%s/%d", ex, n),
			Title:         fmt.Sprintf("%s %s %s", ex, c, td),
			PublishedAt:   pub,
			ContentText:   body,
			ContentSHA256: dedup.ContentSHA256(body),
			Commodity:     c,
			Exchange:      ex,
			TradeDate:     td,
		})
	}
	return out
}
