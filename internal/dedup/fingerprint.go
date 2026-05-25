package dedup

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"mining-pipeline/internal/model"
)

func ContentSHA256(text string) string {
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:])
}

func DedupKey(d model.Document) string {
	if d.CanonicalURL != "" {
		return fmt.Sprintf("%s|%s", d.SourceType, d.CanonicalURL)
	}
	return fmt.Sprintf("%s|%s", d.SourceType, d.ContentSHA256)
}

type Set struct {
	keys map[string]struct{}
}

func NewSet() *Set {
	return &Set{keys: make(map[string]struct{})}
}

func (s *Set) Add(d model.Document) bool {
	k := DedupKey(d)
	if _, ok := s.keys[k]; ok {
		return false
	}
	s.keys[k] = struct{}{}
	return true
}
