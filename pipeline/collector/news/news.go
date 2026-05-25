package news

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"mining-pipeline/internal/dedup"
	"mining-pipeline/internal/httputil"
	"mining-pipeline/internal/model"
	"mining-pipeline/pipeline/processor"
)

var feeds = []struct {
	Name string
	URL  string
}{
	{"mining.com", "https://www.mining.com/feed/"},
	{"spglobal", "https://www.spglobal.com/marketintelligence/en/news-insights/latest-news/rss"},
}

type rssItem struct {
	Title   string `xml:"title"`
	Link    string `xml:"link"`
	PubDate string `xml:"pubDate"`
}

func Collect(now time.Time, maxPerFeed int) ([]model.Document, []string, error) {
	var docs []model.Document
	var warnings []string
	for _, f := range feeds {
		items, err := fetchRSS(f.URL, maxPerFeed)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("news:%s:%v", f.Name, err))
			continue
		}
		for i, it := range items {
			doc, err := fetchArticle(f.Name, it, i, now)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("news:%s:item:%v", f.Name, err))
				continue
			}
			if doc != nil {
				docs = append(docs, *doc)
			}
		}
	}
	return docs, warnings, nil
}

func fetchRSS(url string, limit int) ([]rssItem, error) {
	b, err := httputil.Get(url)
	if err != nil {
		return nil, err
	}
	var channel struct {
		Items []rssItem `xml:"channel>item"`
	}
	if err := xml.Unmarshal(b, &channel); err != nil {
		return nil, err
	}
	if limit > 0 && len(channel.Items) > limit {
		channel.Items = channel.Items[:limit]
	}
	return channel.Items, nil
}

func fetchArticle(source string, it rssItem, idx int, now time.Time) (*model.Document, error) {
	link := strings.TrimSpace(it.Link)
	if link == "" {
		return nil, fmt.Errorf("empty link")
	}
	body, err := httputil.Get(link)
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	title := strings.TrimSpace(it.Title)
	if title == "" {
		title = doc.Find("h1").First().Text()
	}
	content := doc.Find("article").Text()
	if strings.TrimSpace(content) == "" {
		content = doc.Find("main").Text()
	}
	if strings.TrimSpace(content) == "" {
		content = doc.Find("body").Text()
	}
	content = strings.TrimSpace(content)
	if len(content) < 80 {
		return nil, fmt.Errorf("content too short")
	}
	pub := parsePubDate(it.PubDate, now)
	if !processor.WithinLast30Days(pub, now) {
		return nil, nil
	}
	id := fmt.Sprintf("news-%s-%d", source, idx)
	d := model.Document{
		DocumentID:    id,
		SourceType:    model.SourceNews,
		CanonicalURL:  processor.NormalizeURL(link),
		Title:         title,
		PublishedAt:   pub,
		ContentText:   content,
		ContentSHA256: dedup.ContentSHA256(content),
	}
	return &d, nil
}

func parsePubDate(s string, fallback time.Time) time.Time {
	layouts := []string{time.RFC1123, time.RFC1123Z, time.RFC3339, "Mon, 02 Jan 2006 15:04:05 MST"}
	for _, l := range layouts {
		if t, err := time.Parse(l, strings.TrimSpace(s)); err == nil {
			return t.UTC()
		}
	}
	return fallback.UTC()
}
