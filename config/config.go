package config

import (
	"io/ioutil"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// Log ...
type Log struct {
	URLs        []string `yaml:"urls"`
	Description string   `yaml:"description"`
	Type        string   `yaml:"type"`
	Regexp      string   `yaml:"regexp"`
	MultiLine   bool     `yaml:"multiLine"`
	TimeFormat  string   `yaml:"timeFormat"`
	TimeZone    string   `yaml:"timeZone"`
	Tags        []string `yaml:"tags"`
}

// Target ...
type Target struct {
	URL         string
	Description string
	Type        string
	Regexp      string
	MultiLine   bool
	TimeFormat  string
	TimeZone    string
	Tags        []string
	Scheme      string
	Host        string
	User        string
	Port        int
	Path        string
}

// Config ...
type Config struct {
	Targets []Target
	Logs    []Log `yaml:"logs"`
}

// NewConfig ...
func NewConfig() (*Config, error) {
	return &Config{
		Targets: []Target{},
		Logs:    []Log{},
	}, nil
}

// LoadConfigFile ...
func (c *Config) LoadConfigFile(path string) error {
	if path == "" {
		return errors.New("failed to load config file")
	}
	fullPath, err := filepath.Abs(path)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to load config file")
	}
	buf, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to load config file")
	}
	err = yaml.Unmarshal(buf, c)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to load config file")
	}
	for _, l := range c.Logs {
		for _, URL := range l.URLs {
			target := Target{}
			target.Description = l.Description
			target.Type = l.Type
			target.Regexp = l.Regexp
			target.MultiLine = l.MultiLine
			target.TimeFormat = l.TimeFormat
			target.TimeZone = l.TimeZone
			target.Tags = l.Tags

			u, err := url.Parse(URL)
			if err != nil {
				return err
			}
			target.Scheme = u.Scheme
			target.Path = u.Path
			target.User = u.User.Username()
			if strings.Contains(u.Host, ":") {
				splited := strings.Split(u.Host, ":")
				target.Host = splited[0]
				target.Port, _ = strconv.Atoi(splited[1])
			} else {
				target.Host = u.Host
				target.Port = 0
			}
			c.Targets = append(c.Targets, target)
		}
	}
	return nil
}
