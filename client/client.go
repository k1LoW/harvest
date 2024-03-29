package client

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"math/rand"
	"path/filepath"
	"regexp"
	"time"

	"go.uber.org/zap"
)

const (
	initialScanTokenSize = 4096
	maxScanTokenSize     = 1024 * 1024
)

// Client ...
type Client interface {
	Read(ctx context.Context, st, et *time.Time, timeFormat, timeZone string) error
	Tailf(ctx context.Context) error
	RandomOne(ctx context.Context) error
	Ls(ctx context.Context, st *time.Time, et *time.Time) error
	Copy(ctx context.Context, filePath string, dstDir string) error
	Out() <-chan Line
}

// Line ...
type Line struct {
	Host               string
	Path               string
	Content            string
	TimeZone           string
	TimestampViaClient *time.Time
}

var syslogTimestampAMRe = regexp.MustCompile(`^([a-zA-Z]{3}) ([0-9] .+)$`)

// buildReadCommand ...
func buildReadCommand(path string, st, et *time.Time, timeFormat, timeZone string) string {
	dir := filepath.Dir(path)
	base := filepath.Base(path)

	stRunes := []rune(st.Format(fmt.Sprintf("%s %s", timeFormat, timeZone)))
	etRunes := []rune(et.Format(fmt.Sprintf("%s %s", timeFormat, timeZone)))

	matches := []rune{}
	for idx, r := range stRunes {
		if r != etRunes[idx] {
			break
		}
		matches = append(matches, r)
	}

	grepStr := string(matches)
	// for syslog timestamp
	if syslogTimestampAMRe.MatchString(grepStr) {
		grepStr = syslogTimestampAMRe.ReplaceAllString(string(matches), "$1  $2")
	}

	findStart := st.Format("2006-01-02 15:04:05 MST")

	cmd := fmt.Sprintf("sudo find %s/ -type f -name '%s' -newermt '%s' | xargs sudo ls -tr | xargs sudo zcat -f | grep -a '%s'", dir, base, findStart, grepStr)
	if timeFormat == "unixtime" {
		cmd = fmt.Sprintf("sudo find %s/ -type f -name '%s' -newermt '%s' | xargs sudo ls -tr | xargs sudo zcat -f", dir, base, findStart)
	}

	return cmd
}

// buildTailfCommand ...
func buildTailfCommand(path string) string {
	dir := filepath.Dir(path)
	base := filepath.Base(path)

	cmd := fmt.Sprintf("sudo find %s/ -type f -name '%s' | xargs sudo ls -tr | tail -1 | xargs sudo tail -F", dir, base)

	return cmd
}

// buildLsCommand ...
func buildLsCommand(path string, st *time.Time) string {
	dir := filepath.Dir(path)
	base := filepath.Base(path)

	stStr := st.Format("2006-01-02 15:04:05 MST")

	cmd := fmt.Sprintf("sudo find %s/ -type f -name '%s' -newermt '%s' | xargs sudo ls -tr", dir, base, stStr)

	return cmd
}

// buildRandomOneCommand ...
func buildRandomOneCommand(path string) string {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	rand.Seed(time.Now().UnixNano())

	// why tail -2 -> for 0 line log
	cmd := fmt.Sprintf("sudo find %s/ -type f -name '%s' | xargs sudo ls -tr | tail -2 | xargs sudo zcat -f | head -%d | tail -1", dir, base, rand.Intn(100)) // #nosec

	return cmd
}

func bindReaderAndChan(ctx context.Context, l *zap.Logger, r *io.Reader, lineChan chan Line, host string, path string, tz string) {
	defer func() {
		l.Debug("Close chan client.Line")
		close(lineChan)
	}()
	scanner := bufio.NewScanner(*r)
	buf := make([]byte, initialScanTokenSize)
	scanner.Buffer(buf, maxScanTokenSize)
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
		l.Error("Fetch error", zap.Error(scanner.Err()))
	}
}
