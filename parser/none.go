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
		return p.parseMultipleLine(ctx, cancel, lineChan, tz)
	}
	return p.parseSingleLine(ctx, cancel, lineChan, tz)
}

func (p *NoneParser) parseSingleLine(ctx context.Context, cancel context.CancelFunc, lineChan <-chan client.Line, tz string) <-chan Log {
	logChan := make(chan Log)
	var (
		prevTs *time.Time
	)

	go func() {
		defer close(logChan)
	L:
		for line := range lineChan {
			var ts *time.Time

			if line.TimestampViaClient != nil {
				ts = line.TimestampViaClient
				prevTs = ts
			} else {
				ts = prevTs
			}

			logChan <- Log{
				Host:           line.Host,
				Path:           line.Path,
				Timestamp:      ts,
				FilledByPrevTs: false,
				Content:        line.Content,
				Target:         p.target,
			}

			select {
			case <-ctx.Done():
				break L
			default:
			}
		}
	}()

	return logChan
}

func (p *NoneParser) parseMultipleLine(ctx context.Context, cancel context.CancelFunc, lineChan <-chan client.Line, tz string) <-chan Log {
	logChan := make(chan Log)
	contentStash := []string{}

	var (
		hostStash string
		pathStash string
		prevTs    *time.Time
	)

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
	L:
		for line := range lineChan {
			hostStash = line.Host
			pathStash = line.Path
			var ts *time.Time

			if line.TimestampViaClient != nil {
				ts := line.TimestampViaClient
				prevTs = ts
			}

			if strings.HasPrefix(line.Content, " ") || strings.HasPrefix(line.Content, "\t") {
				contentStash = append(contentStash, line.Content)
				if len(contentStash) > maxContentStash {
					logChan <- Log{
						Host:           line.Host,
						Path:           line.Path,
						Timestamp:      ts,
						FilledByPrevTs: false,
						Content:        strings.Join(contentStash, "\n"),
						Target:         p.target,
					}
					logChan <- Log{
						Host:           line.Host,
						Path:           line.Path,
						Timestamp:      ts,
						FilledByPrevTs: false,
						Content:        "Harvest parse error: too many rows",
						Target:         p.target,
					}
					contentStash = nil
				}
				continue
			}

			if len(contentStash) > 0 {
				logChan <- Log{
					Host:           line.Host,
					Path:           line.Path,
					Timestamp:      ts,
					FilledByPrevTs: false,
					Content:        strings.Join(contentStash, "\n"),
					Target:         p.target,
				}
			}

			contentStash = nil
			contentStash = append(contentStash, line.Content)

			select {
			case <-ctx.Done():
				break L
			default:
			}
		}
	}()

	return logChan
}
