package client

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"go.uber.org/zap"
)

// FileClient ...
type FileClient struct {
	path     string
	lineChan chan Line
	logger   *zap.Logger
}

// NewFileClient ...
func NewFileClient(l *zap.Logger, path string) (Client, error) {
	return &FileClient{
		path:     path,
		lineChan: make(chan Line),
		logger:   l,
	}, nil
}

// Read ...
func (c *FileClient) Read(ctx context.Context, st *time.Time, et *time.Time) error {
	cmd := buildReadCommand(c.path, st)
	if runtime.GOOS == "darwin" {
		cmd = strings.Replace(cmd, "zcat", "gzcat", -1)
	}

	return c.Exec(ctx, cmd)
}

// Tailf ...
func (c *FileClient) Tailf(ctx context.Context) error {
	cmd := buildTailfCommand(c.path)
	return c.Exec(ctx, cmd)
}

// Ls ...
func (c *FileClient) Ls(ctx context.Context, st *time.Time, et *time.Time) error {
	cmd := buildLsCommand(c.path, st)
	return c.Exec(ctx, cmd)
}

// Copy ...
func (c *FileClient) Copy(ctx context.Context, filePath string, dstDir string) error {
	dstLogFilePath := filepath.Join(dstDir, filePath)
	dstLogDir := filepath.Dir(dstLogFilePath)
	err := os.MkdirAll(dstLogDir, 0755) // #nosec
	if err != nil {
		return err
	}
	catCmd := fmt.Sprintf("sudo cat %s > %s", filePath, dstLogFilePath)
	cmd := exec.CommandContext(ctx, "sh", "-c", catCmd) // #nosec
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// RandomOne ...
func (c *FileClient) RandomOne(ctx context.Context) error {
	cmd := buildRandomOneCommand(c.path)
	if runtime.GOOS == "darwin" {
		cmd = strings.Replace(cmd, "zcat", "gzcat", -1)
	}
	return c.Exec(ctx, cmd)
}

// Exec ...
func (c *FileClient) Exec(ctx context.Context, cmdStr string) error {
	defer func() {
		c.logger.Debug("Close chan client.Line")
		close(c.lineChan)
	}()
	c.logger.Info("Create new local exec session")
	tzCmd := exec.Command("date", `+%z`) // #nosec
	tzOut, err := tzCmd.Output()
	if err != nil {
		return err
	}

	innerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	cmd := exec.CommandContext(innerCtx, "sh", "-c", cmdStr) // #nosec

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	// FIXME
	// _, err = cmd.StderrPipe()
	// if err != nil {
	// 	return err
	// }

	r := stdout.(io.Reader)

	err = cmd.Start()
	if err != nil {
		return err
	}

	bindReaderAndChan(innerCtx, cancel, c.logger, &r, c.lineChan, "localhost", c.path, strings.TrimRight(string(tzOut), "\n"))

	err = cmd.Wait()
	if err != nil {
		return err
	}
	<-innerCtx.Done()
	c.logger.Info("Close local exec session")
	return nil
}

// Out ...
func (c *FileClient) Out() <-chan Line {
	return c.lineChan
}
