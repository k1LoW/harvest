package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/k1LoW/harvest/client"
)

// RegexpParser ...
type RegexpParser struct {
	regexp     *regexp.Regexp
	timeFormat string
}

// NewRegexpParser ...
func NewRegexpParser(r string, tf string) (Parser, error) {
	re, err := regexp.Compile(r)
	if err != nil {
		return nil, err
	}
	return &RegexpParser{
		regexp:     re,
		timeFormat: tf,
	}, nil
}

// Parse ...
func (p *RegexpParser) Parse(lineChan <-chan client.Line, tz string, tag []string) <-chan Log {
	logChan := make(chan Log)

	go func() {
		lineTZ := tz
		for line := range lineChan {
			var ts int64
			ts = 0
			if tz == "" {
				lineTZ = line.TimeZone
			}
			if p.timeFormat != "" {
				m := p.regexp.FindStringSubmatch(line.Content)
				if len(m) > 1 {
					t, err := parseTime(p.timeFormat, lineTZ, m[1])
					if err == nil {
						ts = t.UnixNano()
					}
				}
			}
			tStr := ""
			if len(tag) > 0 {
				tStr = fmt.Sprintf("[%s]", strings.Join(tag, "]["))
			}

			logChan <- Log{
				Host:      line.Host,
				Path:      line.Path,
				Tag:       tStr,
				Timestamp: ts,
				Content:   line.Content,
			}
		}
		close(logChan)
	}()

	return logChan
}
