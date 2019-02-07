package parser

import (
	"context"
	"fmt"
	"time"

	"github.com/k1LoW/harvest/client"
)

// Log ...
type Log struct {
	Host         string `db:"host"`
	Path         string `db:"path"`
	Tag          string `db:"tag"`
	Timestamp    int64  `db:"ts"`
	FilledByPrev bool   `db:"filled_by_prev"`
	Content      string `db:"content"`
}

// Parser ...
type Parser interface {
	Parse(ctx context.Context, cancel context.CancelFunc, lineChan <-chan client.Line, tz string, tag []string, st *time.Time, et *time.Time) <-chan Log
	ParseMultipleLine(ctx context.Context, cancel context.CancelFunc, lineChan <-chan client.Line, tz string, tag []string, st *time.Time, et *time.Time) <-chan Log
}

// parseTime ...
func parseTime(tf string, tz string, content string) (time.Time, error) {
	if tz == "" {
		return time.Parse(fmt.Sprintf("2006-01-02 %s", tf), fmt.Sprintf("%s %s", time.Now().Format("2006-01-02"), content))
	}
	return time.Parse(fmt.Sprintf("2006-01-02 -0700 %s", tf), fmt.Sprintf("%s %s %s", time.Now().Format("2006-01-02"), tz, content))
}
