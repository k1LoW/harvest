package client

import (
	"bufio"
	"context"
	"io"

	"go.uber.org/zap"
)

// Client ...
type Client interface {
	Read(ctx context.Context, path string) error
	Out() <-chan Line
}

// Line ...
type Line struct {
	Host     string
	Path     string
	Content  string
	Timezone string
}

func bindReaderAndChan(ctx context.Context, l *zap.Logger, r *io.Reader, lineChan chan Line, host string, path string, tz string) {
	scanner := bufio.NewScanner(*r)
	defer close(lineChan)
L:
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			break L
		default:
			lineChan <- Line{
				Host:     host,
				Path:     path,
				Content:  scanner.Text(),
				Timezone: tz,
			}
		}
	}
}
