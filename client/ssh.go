package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/k1LoW/sshc"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
)

// SSHClient ...
type SSHClient struct {
	host     string
	client   *ssh.Client
	lineChan chan Line
	logger   *zap.Logger
}

// NewSSHClient ...
func NewSSHClient(l *zap.Logger, host string, user string, port int) (Client, error) {
	options := []sshc.Option{}
	if user != "" {
		options = append(options, sshc.User(user))
	}
	if port > 0 {
		options = append(options, sshc.Port(port))
	}
	client, err := sshc.NewClient(host, options...)
	if err != nil {
		return nil, err
	}
	return &SSHClient{
		client:   client,
		host:     host,
		lineChan: make(chan Line),
		logger:   l,
	}, nil
}

// Read ...
func (c *SSHClient) Read(ctx context.Context, path string) error {
	session, err := c.client.NewSession()
	if err != nil {
		return err
	}
	c.logger.Info("Create new SSH session")
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

	cmd := fmt.Sprintf("sudo cat %s", path)

	stdout, err := session.StdoutPipe()
	if err != nil {
		return err
	}
	// FIXME
	// _, err = session.StderrPipe()
	// if err != nil {
	// 	return err
	// }

	go bindReaderAndChan(ctx, c.logger, &stdout, c.lineChan, c.host, path, strings.TrimRight(string(tzOut), "\n"))

	err = session.Start(cmd)
	if err != nil {
		return err
	}
	c.logger.Info("Start reading ...")

	err = session.Wait()
	if err != nil {
		return err
	}

	<-ctx.Done()
	c.logger.Info("Read finished.")

	return nil
}

// Out ...
func (c *SSHClient) Out() <-chan Line {
	return c.lineChan
}
