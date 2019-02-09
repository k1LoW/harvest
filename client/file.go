package client

import (
	"context"
	"io"
	"os/exec"
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
	return c.Exec(ctx, cmd)
}

// Exec ...
func (c *FileClient) Exec(ctx context.Context, cmdStr string) error {
	defer close(c.lineChan)
	tzCmd := exec.Command("date", `+"%z"`)
	tzOut, err := tzCmd.Output()
	if err != nil {
		return err
	}

	innerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	cmd := exec.CommandContext(innerCtx, "sh", "-c", cmdStr)

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

	go bindReaderAndChan(innerCtx, cancel, c.logger, &r, c.lineChan, "localhost", c.path, strings.TrimRight(string(tzOut), "\n"))

	err = cmd.Start()
	if err != nil {
		return err
	}
	c.logger.Info("Start reading ...")

	err = cmd.Wait()
	if err != nil {
		return err
	}

	<-innerCtx.Done()
	c.logger.Info("Read finished.")

	return nil
}

// Out ...
func (c *FileClient) Out() <-chan Line {
	return c.lineChan
}
