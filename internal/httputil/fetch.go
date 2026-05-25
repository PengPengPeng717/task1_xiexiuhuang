package httputil

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

var DefaultClient = &http.Client{Timeout: 30 * time.Second}

func Get(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "MiningPipelineBot/1.0 (+research; contact=local)")
	resp, err := DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("http %d for %s", resp.StatusCode, url)
	}
	return io.ReadAll(resp.Body)
}
