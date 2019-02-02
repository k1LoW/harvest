package parser

import (
	"fmt"
	"regexp"
	"time"

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
func (p *RegexpParser) Parse(lineChan <-chan client.Line) <-chan Log {
	logChan := make(chan Log)
	go func() {
		for line := range lineChan {
			var ts int64
			ts = 0
			if p.timeFormat != "" {
				m := p.regexp.FindStringSubmatch(line.Content)
				if len(m) > 1 {
					if line.Timezone != "" {
						t, _ := time.Parse(fmt.Sprintf("2006-01-02 %s %s", line.Timezone, p.timeFormat), fmt.Sprintf("%s %s", time.Now().Format("2006-01-02 -0700"), m[1]))
						ts = t.UnixNano()
					} else {
						t, _ := time.Parse(fmt.Sprintf("2006-01-02 %s", p.timeFormat), fmt.Sprintf("%s %s", time.Now().Format("2006-01-02"), m[1]))
						ts = t.UnixNano()
					}
				}
			}
			logChan <- Log{
				Host:      line.Host,
				Path:      line.Path,
				Timestamp: ts,
				Content:   line.Content,
			}
		}
		close(logChan)
	}()

	return logChan
}
