package client

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/k1LoW/sshc"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
)

// SSHClient ...
type SSHClient struct {
	host     string
	path     string
	client   *ssh.Client
	lineChan chan Line
	logger   *zap.Logger
}

// NewSSHClient ...
func NewSSHClient(l *zap.Logger, host string, user string, port int, path string, passphrase []byte) (Client, error) {
	options := []sshc.Option{}
	if user != "" {
		options = append(options, sshc.User(user))
	}
	if port > 0 {
		options = append(options, sshc.Port(port))
	}
	options = append(options, sshc.Passphrase(passphrase))

	client, err := sshc.NewClient(host, options...)
	if err != nil {
		return nil, err
	}
	return &SSHClient{
		client:   client,
		host:     host,
		path:     path,
		lineChan: make(chan Line),
		logger:   l,
	}, nil
}

// Read ...
func (c *SSHClient) Read(ctx context.Context, st, et *time.Time, timeFormat, timeZone string) error {
	cmd := buildReadCommand(c.path, st, et, timeFormat, timeZone)
	return c.Exec(ctx, cmd)
}

// Tailf ...
func (c *SSHClient) Tailf(ctx context.Context) error {
	cmd := buildTailfCommand(c.path)
	return c.Exec(ctx, cmd)
}

// Ls ...
func (c *SSHClient) Ls(ctx context.Context, st *time.Time, et *time.Time) error {
	cmd := buildLsCommand(c.path, st)
	return c.Exec(ctx, cmd)
}

// Copy ...
func (c *SSHClient) Copy(ctx context.Context, filePath string, dstDir string) error {
	dstLogFilePath := filepath.Join(dstDir, c.host, filePath)
	dstLogDir := filepath.Dir(dstLogFilePath)
	err := os.MkdirAll(dstLogDir, 0755) // #nosec
	if err != nil {
		return err
	}
	catCmd := fmt.Sprintf("ssh %s sudo cat %s > %s", c.host, filePath, dstLogFilePath)
	cmd := exec.CommandContext(ctx, "sh", "-c", catCmd) // #nosec
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// RandomOne ...
func (c *SSHClient) RandomOne(ctx context.Context) error {
	cmd := buildRandomOneCommand(c.path)

	return c.Exec(ctx, cmd)
}

// Exec ...
func (c *SSHClient) Exec(ctx context.Context, cmd string) error {
	session, err := c.client.NewSession()
	if err != nil {
		return err
	}
	c.logger.Debug("Create new SSH session")
	defer session.Close()

	var tzOut []byte
	err = func() error {
		session, err := c.client.NewSession()
		if err != nil {
			return err
		}
		defer session.Close()
		tzCmd := `date +"%z"`
		tzOut, err = session.Output(tzCmd)
		if err != nil {
			return err
		}
		return nil
	}()
	if err != nil {
		return err
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		return err
	}
	// FIXME
	// _, err = session.StderrPipe()
	// if err != nil {
	// 	return err
	// }

	innerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go bindReaderAndChan(innerCtx, c.logger, &stdout, c.lineChan, c.host, c.path, strings.TrimRight(string(tzOut), "\n"))

	err = session.Start(cmd)
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		err = session.Close()
		if err != nil && err != io.EOF {
			c.logger.Error(fmt.Sprintf("%s", err))
		}
	}()

	// TODO: use session.Signal()
	// https://github.com/golang/go/issues/16597
	_ = session.Wait()

	c.logger.Debug("Close SSH session")

	return nil
}

// Out ...
func (c *SSHClient) Out() <-chan Line {
	return c.lineChan
}
