package collector

import (
	"context"
	"fmt"

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
func NewCollector(ctx context.Context, t *config.Target) (*Collector, error) {
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

	l := logger.NewLogger().With(zap.String("host", host), zap.String("path", t.Path))

	// Set client
	switch t.Scheme {
	case "ssh":
		sshc, err := client.NewSSHClient(l, host, t.User, t.Port)
		if err != nil {
			return nil, err
		}
		c = sshc
	case "file":
		filec, err := client.NewFileClient(l)
		if err != nil {
			return nil, err
		}
		c = filec
	default:
		return nil, fmt.Errorf("unsupport scheme: %s", t.Scheme)
	}

	// Set parser
	switch t.Type {
	case "syslog":
		p, err = parser.NewSyslogParser()
		if err != nil {
			return nil, err
		}
	case "combinedLog":
		p, err = parser.NewCombinedLogParser()
		if err != nil {
			return nil, err
		}
	default: // regexp
		p, err = parser.NewRegexpParser(t.Regexp, t.TimeFormat)
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

// Collect ...
func (c *Collector) Collect(dbChan chan parser.Log) error {
	innerCtx, cancel := context.WithCancel(c.ctx)
	defer cancel()

	go func() {
	L:
		for log := range c.parser.Parse(c.client.Out(), c.target.TimeZone, c.target.Tags) {
			select {
			case <-c.ctx.Done():
				break L
			case <-innerCtx.Done():
				break L
			default:
				dbChan <- log
			}
		}
		cancel()
	}()

	err := c.client.Read(innerCtx, c.target.Path)
	if err != nil {
		return err
	}

	return nil
}
