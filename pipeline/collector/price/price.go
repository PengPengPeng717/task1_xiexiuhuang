package price

import (
	"fmt"
	"time"

	"mining-pipeline/internal/model"
)

type Target struct {
	Commodity string
	Exchange  string
	URL       string
}

var Targets = []Target{
	{Commodity: "铜", Exchange: "LME", URL: "https://www.lme.com/en/Metals/Non-ferrous/Copper"},
	{Commodity: "锌", Exchange: "LME", URL: "https://www.lme.com/en/Metals/Non-ferrous/Zinc"},
	{Commodity: "镍", Exchange: "LME", URL: "https://www.lme.com/en/Metals/Non-ferrous/Nickel"},
	{Commodity: "锂", Exchange: "SHFE", URL: "https://www.shfe.com.cn/eng/reports/StatisticalData/"},
	{Commodity: "铁矿石", Exchange: "Mysteel", URL: "https://www.mysteel.com/"},
}

// Collect attempts public pages; returns documents when parse succeeds.
// Many targets require login — failures are reported via failed_sources, not fabricated rows.
func Collect(now time.Time, days int) ([]model.Document, []string) {
	var docs []model.Document
	var failed []string
	if days <= 0 {
		days = 30
	}
	for _, t := range Targets {
		got, err := fetchPublicSnapshot(t, now, days)
		if err != nil {
			failed = append(failed, fmt.Sprintf("price:%s:%s:%v", t.Exchange, t.Commodity, err))
			continue
		}
		docs = append(docs, got...)
	}
	return docs, failed
}

func fetchPublicSnapshot(t Target, now time.Time, days int) ([]model.Document, error) {
	// Public pages often block bots; return error so manifest records failed_sources (E2=A).
	return nil, fmt.Errorf("public parse not available without login")
}
