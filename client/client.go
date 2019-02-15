package client

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"math/rand"
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
	Read(ctx context.Context, st *time.Time, et *time.Time) error
	Tailf(ctx context.Context) error
	RandomOne(ctx context.Context) error
	Out() <-chan Line
}

// Line ...
type Line struct {
	Host     string
	Path     string
	Content  string
	TimeZone string
}

// buildReadCommand ...
func buildReadCommand(path string, st *time.Time) string {
	dir := filepath.Dir(path)
	base := filepath.Base(path)

	stStr := st.Format("2006-01-02 15:04:05 MST")

	cmd := fmt.Sprintf("sudo find %s/ -type f -name '%s' -newermt '%s' | xargs sudo ls -tr | xargs sudo zcat -f", dir, base, stStr)

	return cmd
}

// buildTailfCommand ...
func buildTailfCommand(path string) string {
	dir := filepath.Dir(path)
	base := filepath.Base(path)

	cmd := fmt.Sprintf("sudo find %s/ -type f -name '%s' | xargs sudo ls -tr | tail -1 | xargs sudo tail -F", dir, base)

	return cmd
}

// buildRandomOneCommand ...
func buildRandomOneCommand(path string) string {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	rand.Seed(time.Now().UnixNano())

	// why tail -2 -> for 0 line log
	cmd := fmt.Sprintf("sudo find %s/ -type f -name '%s' | xargs sudo ls -tr | tail -2 | xargs sudo zcat -f | head -%d | tail -1", dir, base, rand.Intn(100))

	return cmd
}

func bindReaderAndChan(ctx context.Context, cancel context.CancelFunc, l *zap.Logger, r *io.Reader, lineChan chan Line, host string, path string, tz string) {
	defer cancel()

	scanner := bufio.NewScanner(*r)
	buf := make([]byte, startBufSize)
	scanner.Buffer(buf, maxScanTokenSize)
L:
	for scanner.Scan() {
		lineChan <- Line{
			Host:     host,
			Path:     path,
			Content:  scanner.Text(),
			TimeZone: tz,
		}
		select {
		case <-ctx.Done():
			break L
		default:
		}
	}
	if scanner.Err() != nil {
		l.Error("Fetch error", zap.Error(scanner.Err()))
	}
}
