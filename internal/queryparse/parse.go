package queryparse

import (
	"regexp"
	"strings"

	"mining-pipeline/internal/model"
)

type Parsed struct {
	Days       int
	SourceType model.SourceType
	Keywords   []string
	Region     string
	Commodity  string
}

var (
	reDays = regexp.MustCompile(`近\s*(\d+)\s*天|last\s*(\d+)\s*days?`)
)

var commodityAliases = map[string]string{
	"锂": "锂", "lithium": "锂",
	"铜": "铜", "copper": "铜",
	"锌": "锌", "zinc": "锌",
	"镍": "镍", "nickel": "镍",
	"铁矿石": "铁矿石", "铁矿": "铁矿石", "iron ore": "铁矿石",
}

var regionAliases = []string{"澳洲", "澳大利亚", "australia", "中国", "china"}

func Parse(question string) Parsed {
	q := strings.ToLower(strings.TrimSpace(question))
	p := Parsed{Days: 0}

	if m := reDays.FindStringSubmatch(question); len(m) > 0 {
		for i := 1; i < len(m); i++ {
			if m[i] != "" {
				p.Days = atoi(m[i])
				break
			}
		}
	}

	switch {
	case strings.Contains(q, "政策") || strings.Contains(q, "policy") || strings.Contains(q, "出口"):
		p.SourceType = model.SourcePolicy
	case strings.Contains(q, "价格") || strings.Contains(q, "行情") || strings.Contains(q, "price"):
		p.SourceType = model.SourcePrice
	case strings.Contains(q, "新闻") || strings.Contains(q, "news"):
		p.SourceType = model.SourceNews
	}

	for _, r := range regionAliases {
		if strings.Contains(q, strings.ToLower(r)) {
			p.Region = r
			break
		}
	}
	for k, v := range commodityAliases {
		if strings.Contains(q, strings.ToLower(k)) {
			p.Commodity = v
			break
		}
	}

	words := regexp.MustCompile(`\p{Han}+|[a-zA-Z]{3,}`).FindAllString(question, -1)
	for _, w := range words {
		w = strings.TrimSpace(w)
		if len(w) < 2 {
			continue
		}
		if strings.Contains(w, "近") || strings.Contains(w, "天") {
			continue
		}
		p.Keywords = append(p.Keywords, w)
	}
	return p
}

func atoi(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			continue
		}
		n = n*10 + int(c-'0')
	}
	return n
}
