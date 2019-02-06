package client

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

const (
	maxScanTokenSize = 128 * 1024
	startBufSize     = 4096
)

// Client ...
type Client interface {
	Read(ctx context.Context, path string, st time.Time) error
	Out() <-chan Line
}

// Line ...
type Line struct {
	Host     string
	Path     string
	Content  string
	TimeZone string
}

// buildCommand ...
func buildCommand(path string, st time.Time) string {
	dir := filepath.Dir(path)
	base := filepath.Base(path)

	stStr := st.Format("2006-01-02 15:04:05 MST")

	cmd := fmt.Sprintf("sudo find %s -type f -name '%s' -newermt '%s' -exec zcat -f {} \\;", dir, base, stStr)

	return cmd
}

func bindReaderAndChan(ctx context.Context, l *zap.Logger, r *io.Reader, lineChan chan Line, host string, path string, tz string) {
	scanner := bufio.NewScanner(*r)
	buf := make([]byte, startBufSize)
	scanner.Buffer(buf, maxScanTokenSize)

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
				TimeZone: tz,
			}
		}
	}
	if scanner.Err() != nil {
		l.Fatal("Fetch error", zap.Error(scanner.Err()))
	}
}
