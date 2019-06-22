package collector

import (
	"context"
	"fmt"
	"time"

	"github.com/k1LoW/harvest/client"
	"github.com/k1LoW/harvest/config"
	"github.com/k1LoW/harvest/logger"
	"github.com/k1LoW/harvest/parser"
	"go.uber.org/zap"
)

// Collector ...
type Collector struct {
	client client.Client
	parser parser.Parser
	target *config.Target
	ctx    context.Context
	logger *zap.Logger
}

// NewCollector ...
func NewCollector(ctx context.Context, t *config.Target, logSilent bool) (*Collector, error) {
	var (
		host string
		err  error
		c    client.Client
		p    parser.Parser
	)

	host = t.Host
	if host == "" {
		host = "localhost"
	}

	var l *zap.Logger
	if logSilent {
		l = logger.NewSilentLogger().With(zap.String("host", host), zap.String("path", t.Path))
	} else {
		l = logger.NewLogger().With(zap.String("host", host), zap.String("path", t.Path))
	}

	// Set client
	switch t.Scheme {
	case "ssh":
		sshc, err := client.NewSSHClient(l, host, t.User, t.Port, t.Path, t.SSHKeyPassphrase)
		if err != nil {
			return nil, err
		}
		c = sshc
	case "file":
		filec, err := client.NewFileClient(l, t.Path)
		if err != nil {
			return nil, err
		}
		c = filec
	case "k8s":
		k8sc, err := client.NewK8sClient(l, host, t.Path)
		if err != nil {
			return nil, err
		}
		c = k8sc
	default:
		return nil, fmt.Errorf("unsupport scheme: %s", t.Scheme)
	}

	// Set parser
	switch t.Type {
	case "syslog":
		p, err = parser.NewSyslogParser(t)
		if err != nil {
			return nil, err
		}
	case "combinedLog":
		p, err = parser.NewCombinedLogParser(t)
		if err != nil {
			return nil, err
		}
	case "none", "k8s":
		p, err = parser.NewNoneParser(t)
		if err != nil {
			return nil, err
		}
	default: // regexp
		p, err = parser.NewRegexpParser(t)
		if err != nil {
			return nil, err
		}
	}

	return &Collector{
		client: c,
		parser: p,
		target: t,
		ctx:    ctx,
		logger: l,
	}, nil
}

// Fetch ...
func (c *Collector) Fetch(dbChan chan parser.Log, st *time.Time, et *time.Time, multiLine bool) error {
	waiter := make(chan struct{})
	innerCtx, cancel := context.WithCancel(c.ctx)
	defer cancel()

	go func() {
		defer func() {
			cancel()
			waiter <- struct{}{}
		}()
	L:
		for log := range c.parser.Parse(innerCtx, cancel, c.client.Out(), c.target.TimeZone, st, et) {
			dbChan <- log
			select {
			case <-c.ctx.Done():
				break L
			case <-innerCtx.Done():
				break L
			default:
			}
		}
	}()

	err := c.client.Read(innerCtx, st, et)
	if err != nil {
		return err
	}

	<-waiter
	return nil
}

// Stream ...
func (c *Collector) Stream(logChan chan parser.Log, multiLine bool) error {
	waiter := make(chan struct{})
	innerCtx, cancel := context.WithCancel(c.ctx)
	defer cancel()

	go func() {
		defer func() {
			cancel()
			waiter <- struct{}{}
		}()
	L:
		for log := range c.parser.Parse(innerCtx, cancel, c.client.Out(), c.target.TimeZone, nil, nil) {
			logChan <- log
			select {
			case <-c.ctx.Done():
				break L
			case <-innerCtx.Done():
				break L
			default:
			}
		}
	}()

	err := c.client.Tailf(innerCtx)
	if err != nil {
		return err
	}

	<-waiter
	return nil
}

// LsLogs ...
func (c *Collector) LsLogs(logChan chan parser.Log, st *time.Time, et *time.Time) error {
	waiter := make(chan struct{})
	innerCtx, cancel := context.WithCancel(c.ctx)
	defer cancel()

	go func() {
		defer func() {
			cancel()
			waiter <- struct{}{}
		}()
	L:
		for line := range c.client.Out() {
			logChan <- parser.Log{
				Host:    line.Host,
				Path:    line.Path,
				Content: line.Content,
				Target:  c.target,
			}
			select {
			case <-c.ctx.Done():
				break L
			case <-innerCtx.Done():
				break L
			default:
			}
		}
	}()

	err := c.client.Ls(innerCtx, st, et)
	if err != nil {
		return err
	}

	<-waiter
	return nil
}

// Copy ...
func (c *Collector) Copy(logChan chan parser.Log, st *time.Time, et *time.Time, dstDir string) error {
	waiter := make(chan struct{})
	innerCtx, cancel := context.WithCancel(c.ctx)
	defer cancel()
	fileChan := make(chan parser.Log)

	go func() {
		defer func() {
			cancel()
			waiter <- struct{}{}
		}()
		files := []parser.Log{}
		for file := range fileChan {
			files = append(files, file)
		}
		for _, file := range files {
			filePath := file.Content
			c.logger.Info(fmt.Sprintf("Start copying %s", filePath), zap.String("host", c.target.Host), zap.String("path", c.target.Path))
			err := c.client.Copy(innerCtx, filePath, dstDir)
			if err != nil {
				c.logger.Error("Copy error", zap.String("host", c.target.Host), zap.String("path", c.target.Path), zap.String("error", err.Error()))
			} else {
				c.logger.Info(fmt.Sprintf("Copy %s finished", filePath), zap.String("host", c.target.Host), zap.String("path", c.target.Path))
			}
		}
	}()

	err := c.LsLogs(fileChan, st, et)
	if err != nil {
		return err
	}
	close(fileChan)

	<-waiter
	return nil
}

// ConfigTest ...
func (c *Collector) ConfigTest(logChan chan parser.Log, multiLine bool) error {
	waiter := make(chan struct{})
	innerCtx, cancel := context.WithCancel(c.ctx)
	defer cancel()

	go func() {
		defer func() {
			cancel()
			waiter <- struct{}{}
		}()
	L:
		for log := range c.parser.Parse(innerCtx, cancel, c.client.Out(), c.target.TimeZone, nil, nil) {
			logChan <- log
			select {
			case <-c.ctx.Done():
				break L
			case <-innerCtx.Done():
				break L
			default:
			}
		}
	}()

	err := c.client.RandomOne(innerCtx)
	if err != nil {
		return err
	}

	<-waiter
	close(logChan)
	return nil
}
