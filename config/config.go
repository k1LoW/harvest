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

// Target ...
type Target struct {
	URL         string   `yaml:"url"`
	Description string   `yaml:"description"`
	Type        string   `yaml:"type"`
	Regexp      string   `yaml:"regexp"`
	TimeFormat  string   `yaml:"timeFormat"`
	TimeZone    string   `yaml:"timeZone"`
	Tags        []string `yaml:"tags"`
	Scheme      string
	Host        string
	User        string
	Port        int
	Path        string
}

// Config ...
type Config struct {
	Targets []Target `yaml:"logs"`
}

// NewConfig ...
func NewConfig() (*Config, error) {
	return &Config{
		Targets: []Target{},
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
	for i, t := range c.Targets {
		u, err := url.Parse(t.URL)
		if err != nil {
			return err
		}
		c.Targets[i].Scheme = u.Scheme
		c.Targets[i].Path = u.Path
		c.Targets[i].User = u.User.Username()
		if strings.Contains(u.Host, ":") {
			splited := strings.Split(u.Host, ":")
			c.Targets[i].Host = splited[0]
			c.Targets[i].Port, _ = strconv.Atoi(splited[1])
		} else {
			c.Targets[i].Host = u.Host
			c.Targets[i].Port = 0
		}
	}
	return nil
}
