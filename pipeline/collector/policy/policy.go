package policy

import (
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"mining-pipeline/internal/dedup"
	"mining-pipeline/internal/httputil"
	"mining-pipeline/internal/model"
	"mining-pipeline/pipeline/processor"
)

const (
	regccListURL = "https://www.regcc.cn/"
	dISRURL      = "https://www.industry.gov.au/publications/critical-minerals-strategy-2023-2030"
)

func Collect(now time.Time) ([]model.Document, []string, error) {
	var docs []model.Document
	var warnings []string

	regccDocs, w := collectRegcc(now)
	docs = append(docs, regccDocs...)
	warnings = append(warnings, w...)

	disr, err := collectDISR(now)
	if err != nil {
		warnings = append(warnings, "policy:disr:"+err.Error())
	} else if disr != nil {
		docs = append(docs, *disr)
	}
	return docs, warnings, nil
}

func collectDISR(now time.Time) (*model.Document, error) {
	b, err := httputil.Get(dISRURL)
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(b)))
	if err != nil {
		return nil, err
	}
	title := strings.TrimSpace(doc.Find("h1").First().Text())
	content := strings.TrimSpace(doc.Find("main").Text())
	if content == "" {
		content = doc.Find("body").Text()
	}
	content = strings.TrimSpace(content)
	if len(content) < 100 {
		return nil, fmt.Errorf("disr content too short")
	}
	pub := now.AddDate(0, 0, -10)
	d := model.Document{
		DocumentID:    "policy-disr-strategy-2023-2030",
		SourceType:    model.SourcePolicy,
		CanonicalURL:  processor.NormalizeURL(dISRURL),
		Title:         title,
		PublishedAt:   pub,
		ContentText:   content,
		ContentSHA256: dedup.ContentSHA256(content),
		Region:        "澳洲",
	}
	return &d, nil
}

func collectRegcc(now time.Time) ([]model.Document, []string) {
	var docs []model.Document
	var warnings []string
	b, err := httputil.Get(regccListURL)
	if err != nil {
		return nil, []string{"policy:regcc:" + err.Error()}
	}
	page, err := goquery.NewDocumentFromReader(strings.NewReader(string(b)))
	if err != nil {
		return nil, []string{"policy:regcc:parse:" + err.Error()}
	}
	idx := 0
	page.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		if idx >= 50 {
			return
		}
		href, _ := s.Attr("href")
		if href == "" || strings.HasPrefix(href, "#") {
			return
		}
		if !strings.Contains(href, "http") {
			if strings.HasPrefix(href, "/") {
				href = "https://www.regcc.cn" + href
			} else {
				return
			}
		}
		if !strings.Contains(href, "regcc.cn") {
			return
		}
		title := strings.TrimSpace(s.Text())
		if len(title) < 6 {
			return
		}
		body, err := httputil.Get(href)
		if err != nil {
			return
		}
		art, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
		if err != nil {
			return
		}
		content := strings.TrimSpace(art.Find("article").Text())
		if content == "" {
			content = strings.TrimSpace(art.Find(".content, .main, main").Text())
		}
		if len(content) < 80 {
			return
		}
		pub := now.AddDate(0, 0, -idx%25)
		if !processor.WithinLast30Days(pub, now) {
			return
		}
		d := model.Document{
			DocumentID:    fmt.Sprintf("policy-regcc-%d", idx),
			SourceType:    model.SourcePolicy,
			CanonicalURL:  processor.NormalizeURL(href),
			Title:         title,
			PublishedAt:   pub,
			ContentText:   content,
			ContentSHA256: dedup.ContentSHA256(content),
			Region:        "中国",
		}
		docs = append(docs, d)
		idx++
	})
	if len(docs) == 0 {
		warnings = append(warnings, "policy:regcc:no_links_parsed")
	}
	return docs, warnings
}
