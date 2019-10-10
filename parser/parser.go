package parser

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/k1LoW/harvest/client"
	"github.com/k1LoW/harvest/config"
)

const maxContentStash = 1000

// Log ...
type Log struct {
	Host              string `db:"host"`
	Path              string `db:"path"`
	Timestamp         *time.Time
	TimestampUnixNano int64          `db:"ts_unixnano"`
	FilledByPrevTs    bool           `db:"filled_by_prev_ts"`
	Content           string         `db:"content"`
	Target            *config.Target `db:"target"`
}

// Parser ...
type Parser interface {
	Parse(ctx context.Context, cancel context.CancelFunc, lineChan <-chan client.Line, tz string, st *time.Time, et *time.Time) <-chan Log
}

func parseTime(tf string, tz string, content string) (*time.Time, error) {
	if tf == "unixtime" {
		ui, _ := strconv.ParseInt(content, 10, 64)
		ut := time.Unix(ui, 0)
		return &ut, nil
	}
	if tz == "" {
		t, err := time.Parse(fmt.Sprintf("2006-01-02 %s", tf), fmt.Sprintf("%s %s", time.Now().Format("2006-01-02"), content))
		return &t, err
	}
	t, err := time.Parse(fmt.Sprintf("2006-01-02 -0700 %s", tf), fmt.Sprintf("%s %s %s", time.Now().Format("2006-01-02"), tz, content))
	return &t, err
}
