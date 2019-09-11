package parser

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/k1LoW/harvest/client"
	"github.com/k1LoW/harvest/config"
	"go.uber.org/zap"
)

// RegexpParser ...
type RegexpParser struct {
	target *config.Target
	logger *zap.Logger
}

// NewRegexpParser ...
func NewRegexpParser(t *config.Target, l *zap.Logger) (Parser, error) {
	return &RegexpParser{
		target: t,
		logger: l,
	}, nil
}

// Parse ...
func (p *RegexpParser) Parse(ctx context.Context, cancel context.CancelFunc, lineChan <-chan client.Line, tz string, st *time.Time, et *time.Time) <-chan Log {
	if p.target.MultiLine {
		return p.parseMultipleLine(ctx, cancel, lineChan, tz, st, et)
	}
	return p.parseSingleLine(ctx, cancel, lineChan, tz, st, et)
}

func (p *RegexpParser) parseSingleLine(ctx context.Context, cancel context.CancelFunc, lineChan <-chan client.Line, tz string, st *time.Time, et *time.Time) <-chan Log {
	logChan := make(chan Log)
	logStarted := false
	re := regexp.MustCompile(p.target.Regexp)

	var prevTs int64

	if st == nil {
		logStarted = true
	}

	go func() {
		defer func() {
			p.logger.Debug("Close chan parser.Log")
			close(logChan)
		}()
		lineTZ := tz
		for line := range lineChan {
			var (
				ts             int64
				filledByPrevTs bool
			)
			ts = 0
			if tz == "" {
				lineTZ = line.TimeZone
			}
			if p.target.TimeFormat != "" {
				m := re.FindStringSubmatch(line.Content)
				if len(m) > 1 {
					t, err := parseTime(p.target.TimeFormat, lineTZ, m[1])
					if err == nil {
						ts = t.UnixNano()
						if !logStarted && (st == nil || ts > st.UnixNano()) {
							logStarted = true
						}
						prevTs = ts
					}
				}
			}
			if ts == 0 {
				if line.TimestampViaClient != nil {
					tsC := line.TimestampViaClient
					ts = tsC.UnixNano()
					prevTs = ts
				} else {
					ts = prevTs
					filledByPrevTs = true
				}
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
				Timestamp:      ts,
				FilledByPrevTs: filledByPrevTs,
				Content:        line.Content,
				Target:         p.target,
			}
		}
	}()

	return logChan
}

func (p *RegexpParser) parseMultipleLine(ctx context.Context, cancel context.CancelFunc, lineChan <-chan client.Line, tz string, st *time.Time, et *time.Time) <-chan Log {
	logChan := make(chan Log)
	logStarted := false
	re := regexp.MustCompile(p.target.Regexp)
	contentStash := []string{}
	var (
		prevTs    int64
		hostStash string
		pathStash string
	)

	if st == nil {
		logStarted = true
	}

	go func() {
		defer func() {
			logChan <- Log{
				Host:           hostStash,
				Path:           pathStash,
				Timestamp:      prevTs,
				FilledByPrevTs: false,
				Content:        strings.Join(contentStash, "\n"),
				Target:         p.target,
			}
			p.logger.Debug("Close chan parser.Log")
			close(logChan)
		}()

		lineTZ := tz
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
			if p.target.TimeFormat != "" {
				m := re.FindStringSubmatch(line.Content)
				if len(m) > 1 {
					t, err := parseTime(p.target.TimeFormat, lineTZ, m[1])
					if err == nil {
						ts = t.UnixNano()
						if !logStarted && (st == nil || ts > st.UnixNano()) {
							logStarted = true
						}
					}
				}
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
						Timestamp:      prevTs,
						FilledByPrevTs: false,
						Content:        strings.Join(contentStash, "\n"),
						Target:         p.target,
					}
					logChan <- Log{
						Host:           line.Host,
						Path:           line.Path,
						Timestamp:      0,
						FilledByPrevTs: false,
						Content:        "Harvest parse error: too many rows",
						Target:         p.target,
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
					Timestamp:      prevTs,
					FilledByPrevTs: false,
					Content:        strings.Join(contentStash, "\n"),
					Target:         p.target,
				}
			}

			contentStash = nil
			contentStash = append(contentStash, line.Content)
			prevTs = ts
		}
	}()

	return logChan
}
