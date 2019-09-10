package parser

import (
	"github.com/k1LoW/harvest/config"
	"go.uber.org/zap"
)

// NewCombinedLogParser ...
func NewCombinedLogParser(t *config.Target, l *zap.Logger) (Parser, error) {
	t.Regexp = `^[\d\.]+ - [^ ]+ \[(.+)\] .+$`
	t.TimeFormat = "02/Jan/2006:15:04:05 -0700"
	t.MultiLine = false
	return NewRegexpParser(t, l)
}
