package client

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// FileClient ...
type FileClient struct {
	lineChan chan Line
	logger   *zap.Logger
}

// NewFileClient ...
func NewFileClient(l *zap.Logger) (Client, error) {
	return &FileClient{
		lineChan: make(chan Line),
		logger:   l,
	}, nil
}

// Read ...
func (c *FileClient) Read(ctx context.Context, path string, st time.Time) error {
	if _, err := os.Lstat(path); err != nil {
		return errors.Wrap(err, fmt.Sprintf("%s not exists", path))
	}
	tzCmd := exec.Command("date", `+"%z"`)
	tzOut, err := tzCmd.Output()
	if err != nil {
		return err
	}

	cmd := exec.Command("sh", "-c", buildCommand(path, st))

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

	go bindReaderAndChan(ctx, c.logger, &r, c.lineChan, "localhost", path, strings.TrimRight(string(tzOut), "\n"))

	err = cmd.Start()
	if err != nil {
		return err
	}
	c.logger.Info("Start reading ...")

	err = cmd.Wait()
	if err != nil {
		return err
	}

	<-ctx.Done()
	c.logger.Info("Read finished.")

	return nil
}

// Out ...
func (c *FileClient) Out() <-chan Line {
	return c.lineChan
}
