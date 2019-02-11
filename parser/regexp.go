package parser

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/k1LoW/harvest/client"
)

const maxContentStash = 1000

// RegexpParser ...
type RegexpParser struct {
	regexp     *regexp.Regexp
	timeFormat string
	multiLine  bool
}

// NewRegexpParser ...
func NewRegexpParser(r string, tf string, multiLine bool) (Parser, error) {
	re, err := regexp.Compile(r)
	if err != nil {
		return nil, err
	}
	return &RegexpParser{
		regexp:     re,
		timeFormat: tf,
		multiLine:  multiLine,
	}, nil
}

// Parse ...
func (p *RegexpParser) Parse(ctx context.Context, cancel context.CancelFunc, lineChan <-chan client.Line, tz string, tag []string, st *time.Time, et *time.Time) <-chan Log {
	if p.multiLine {
		return p.parseMultipleLine(ctx, cancel, lineChan, tz, tag, st, et)
	}
	return p.parseSingleLine(ctx, cancel, lineChan, tz, tag, st, et)
}

func (p *RegexpParser) parseSingleLine(ctx context.Context, cancel context.CancelFunc, lineChan <-chan client.Line, tz string, tag []string, st *time.Time, et *time.Time) <-chan Log {
	logChan := make(chan Log)
	logStarted := false
	var prevTs int64
	var tStr string

	if len(tag) > 0 {
		tStr = fmt.Sprintf("[%s]", strings.Join(tag, "]["))
	}

	go func() {
		defer close(logChan)
		lineTZ := tz
	L:
		for line := range lineChan {
			var (
				ts             int64
				filledByPrevTs bool
			)
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
						if !logStarted && (st == nil || ts > st.UnixNano()) {
							logStarted = true
						}
						prevTs = ts
					}
				} else {
					ts = prevTs
					filledByPrevTs = true
				}
			}

			select {
			case <-ctx.Done():
				break L
			default:
			}

			if !logStarted {
				continue
			}

			if et != nil && ts > et.UnixNano() {
				cancel()
				continue
			}

			logChan <- Log{
				Host:           line.Host,
				Path:           line.Path,
				Tag:            tStr,
				Timestamp:      ts,
				FilledByPrevTs: filledByPrevTs,
				Content:        line.Content,
			}
		}
	}()

	return logChan
}

func (p *RegexpParser) parseMultipleLine(ctx context.Context, cancel context.CancelFunc, lineChan <-chan client.Line, tz string, tag []string, st *time.Time, et *time.Time) <-chan Log {
	logChan := make(chan Log)
	logStarted := false
	contentStash := []string{}
	var (
		prevTs    int64
		tStr      string
		hostStash string
		pathStash string
	)
	if len(tag) > 0 {
		tStr = fmt.Sprintf("[%s]", strings.Join(tag, "]["))
	}

	go func() {
		defer func() {
			logChan <- Log{
				Host:           hostStash,
				Path:           pathStash,
				Tag:            tStr,
				Timestamp:      prevTs,
				FilledByPrevTs: false,
				Content:        strings.Join(contentStash, "\n"),
			}
			close(logChan)
		}()

		lineTZ := tz
	L:
		for line := range lineChan {
			var (
				ts int64
			)

			hostStash = line.Host
			pathStash = line.Path

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
						if !logStarted && (st == nil || ts > st.UnixNano()) {
							logStarted = true
						}
					}
				}
			}

			select {
			case <-ctx.Done():
				break L
			default:
			}

			if !logStarted {
				continue
			}
			if et != nil && ts > et.UnixNano() {
				cancel()
				continue
			}

			if ts == 0 {
				contentStash = append(contentStash, line.Content)
				if len(contentStash) > maxContentStash {
					logChan <- Log{
						Host:           line.Host,
						Path:           line.Path,
						Tag:            tStr,
						Timestamp:      0,
						FilledByPrevTs: false,
						Content:        "Harvest parse error: too many rows.", // FIXME
					}
					contentStash = nil
				}
				continue
			}

			// ts > 0
			if len(contentStash) > 0 {
				logChan <- Log{
					Host:           line.Host,
					Path:           line.Path,
					Tag:            tStr,
					Timestamp:      prevTs,
					FilledByPrevTs: false,
					Content:        strings.Join(contentStash, "\n"),
				}
			}

			contentStash = nil
			contentStash = append(contentStash, line.Content)
			prevTs = ts
		}
	}()

	return logChan
}
