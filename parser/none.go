package parser

import (
	"context"
	"strings"
	"time"

	"github.com/k1LoW/harvest/client"
	"github.com/k1LoW/harvest/config"
	"go.uber.org/zap"
)

// NoneParser ...
type NoneParser struct {
	target *config.Target
	logger *zap.Logger
}

// NewNoneParser ...
func NewNoneParser(t *config.Target, l *zap.Logger) (Parser, error) {
	return &NoneParser{
		target: t,
		logger: l,
	}, nil
}

// Parse ...
func (p *NoneParser) Parse(ctx context.Context, cancel context.CancelFunc, lineChan <-chan client.Line, tz string, st *time.Time, et *time.Time) <-chan Log {
	if p.target.MultiLine {
		return p.parseMultipleLine(ctx, cancel, lineChan, tz, st, et)
	}
	return p.parseSingleLine(ctx, cancel, lineChan, tz, st, et)
}

func (p *NoneParser) parseSingleLine(ctx context.Context, cancel context.CancelFunc, lineChan <-chan client.Line, tz string, st *time.Time, et *time.Time) <-chan Log {
	logChan := make(chan Log)
	logStarted := false
	logEnded := false

	var prevTs *time.Time

	if st == nil {
		logStarted = true
	}

	go func() {
		defer func() {
			p.logger.Debug("Close chan parser.Log")
			close(logChan)
		}()

		for line := range lineChan {
			if logEnded {
				continue
			}
			var (
				ts             *time.Time
				filledByPrevTs bool
			)

			if line.TimestampViaClient != nil {
				ts = line.TimestampViaClient
				prevTs = ts
			} else {
				ts = prevTs
				if ts != nil {
					filledByPrevTs = true
				}
			}
			if ts == nil {
				logStarted = true
			}

			if !logStarted && ts != nil && ts.UnixNano() > st.UnixNano() {
				logStarted = true
			}

			if !logStarted {
				continue
			}

			if ts != nil && et != nil && ts.UnixNano() > et.UnixNano() {
				p.logger.Debug("Cancel parse, because timestamp period out")
				logEnded = true
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

func (p *NoneParser) parseMultipleLine(ctx context.Context, cancel context.CancelFunc, lineChan <-chan client.Line, tz string, st *time.Time, et *time.Time) <-chan Log {
	logChan := make(chan Log)
	logStarted := false
	logEnded := false
	contentStash := []string{}

	var (
		hostStash string
		pathStash string
		prevTs    *time.Time
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
			close(logChan)
		}()

		for line := range lineChan {
			if logEnded {
				continue
			}
			hostStash = line.Host
			pathStash = line.Path
			var ts *time.Time

			if line.TimestampViaClient != nil {
				ts = line.TimestampViaClient
			} else {
				logStarted = true
			}

			if !logStarted && ts != nil && ts.UnixNano() > st.UnixNano() {
				logStarted = true
			}

			if !logStarted {
				continue
			}

			if ts != nil && et != nil && ts.UnixNano() > et.UnixNano() {
				p.logger.Debug("Cancel parse, because timestamp period out")
				logEnded = true
				cancel()
				continue
			}

			if ts == nil && (strings.HasPrefix(line.Content, " ") || strings.HasPrefix(line.Content, "\t")) {
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
						Timestamp:      prevTs,
						FilledByPrevTs: false,
						Content:        "Harvest parse error: too many rows",
						Target:         p.target,
					}
					contentStash = nil
				}
				continue
			}

			// ts > 0 or ^.+

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
