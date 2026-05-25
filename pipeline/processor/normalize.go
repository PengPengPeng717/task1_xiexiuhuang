package processor

import (
	"strings"
	"time"
)

var ShanghaiLoc *time.Location

func init() {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*3600)
	}
	ShanghaiLoc = loc
}

func NormalizeURL(u string) string {
	u = strings.TrimSpace(u)
	u = strings.TrimSuffix(u, "/")
	return u
}

func WithinLast30Days(t time.Time, now time.Time) bool {
	cutoff := now.In(ShanghaiLoc).AddDate(0, 0, -30)
	return !t.Before(cutoff)
}

func WithinLastNDays(t time.Time, now time.Time, days int) bool {
	cutoff := now.In(ShanghaiLoc).AddDate(0, 0, -days)
	return !t.Before(cutoff)
}
