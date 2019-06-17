package parser

import "github.com/k1LoW/harvest/config"

// NewSyslogParser ...
func NewSyslogParser(t *config.Target) (Parser, error) {
	t.Regexp = `^(\w{3}  ?\d{1,2} \d{2}:\d{2}:\d{2}) .+$`
	t.TimeFormat = "Jan 2 15:04:05"
	t.MultiLine = false
	return NewRegexpParser(t)
}
