package parser

import (
	"fmt"
	"time"

	"github.com/k1LoW/harvest/client"
)

// Log ...
type Log struct {
	Host      string `db:"host"`
	Path      string `db:"path"`
	Timestamp int64  `db:"ts"`
	Content   string `db:"content"`
}

// Parser ...
type Parser interface {
	Parse(lineChan <-chan client.Line, tz string) <-chan Log
}

// parseTime ...
func parseTime(tf string, tz string, content string) (time.Time, error) {
	if tz == "" {
		return time.Parse(fmt.Sprintf("2006-01-02 %s", tf), fmt.Sprintf("%s %s", time.Now().Format("2006-01-02"), content))
	}
	return time.Parse(fmt.Sprintf("2006-01-02 -0700 %s", tf), fmt.Sprintf("%s %s %s", time.Now().Format("2006-01-02"), tz, content))
}
