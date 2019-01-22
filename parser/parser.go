package parser

import (
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
	Parse(lineChan <-chan client.Line) <-chan Log
}
