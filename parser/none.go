package parser

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/k1LoW/harvest/client"
)

// NoneParser ...
type NoneParser struct {
	multiLine bool
}

// NewNoneParser ...
func NewNoneParser(multiLine bool) (Parser, error) {
	return &NoneParser{
		multiLine: multiLine,
	}, nil
}

// Parse ...
func (p *NoneParser) Parse(ctx context.Context, cancel context.CancelFunc, lineChan <-chan client.Line, tz string, tag []string, st *time.Time, et *time.Time) <-chan Log {
	if p.multiLine {
		return p.parseMultipleLine(ctx, cancel, lineChan, tz, tag)
	}
	return p.parseSingleLine(ctx, cancel, lineChan, tz, tag)
}

func (p *NoneParser) parseSingleLine(ctx context.Context, cancel context.CancelFunc, lineChan <-chan client.Line, tz string, tag []string) <-chan Log {
	logChan := make(chan Log)
	var tStr string

	if len(tag) > 0 {
		tStr = fmt.Sprintf("[%s]", strings.Join(tag, "]["))
	}

	go func() {
		defer close(logChan)
	L:
		for line := range lineChan {
			logChan <- Log{
				Host:           line.Host,
				Path:           line.Path,
				Tag:            tStr,
				Timestamp:      0,
				FilledByPrevTs: false,
				Content:        line.Content,
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

func (p *NoneParser) parseMultipleLine(ctx context.Context, cancel context.CancelFunc, lineChan <-chan client.Line, tz string, tag []string) <-chan Log {
	logChan := make(chan Log)
	contentStash := []string{}

	var (
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
				Timestamp:      0,
				FilledByPrevTs: false,
				Content:        strings.Join(contentStash, "\n"),
			}
			close(logChan)
		}()
	L:
		for line := range lineChan {
			hostStash = line.Host
			pathStash = line.Path

			if strings.HasPrefix(line.Content, " ") || strings.HasPrefix(line.Content, "\t") {
				contentStash = append(contentStash, line.Content)
				if len(contentStash) > maxContentStash {
					logChan <- Log{
						Host:           line.Host,
						Path:           line.Path,
						Tag:            tStr,
						Timestamp:      0,
						FilledByPrevTs: false,
						Content:        "Harvest parse error: too many rows", // FIXME
					}
					contentStash = nil
				}
				continue
			}

			if len(contentStash) > 0 {
				logChan <- Log{
					Host:           line.Host,
					Path:           line.Path,
					Tag:            tStr,
					Timestamp:      0,
					FilledByPrevTs: false,
					Content:        strings.Join(contentStash, "\n"),
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
