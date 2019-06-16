package config

import (
	"io/ioutil"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/antonmedv/expr"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// TargetSet ...
type TargetSet struct {
	Sources     []string `yaml:"sources"`
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
	Source           string
	Description      string
	Type             string
	Regexp           string
	MultiLine        bool
	TimeFormat       string
	TimeZone         string
	Tags             []string
	Scheme           string
	Host             string
	User             string
	Port             int
	Path             string
	SSHKeyPassphrase []byte
}

type Tags map[string]int

// Config ...
type Config struct {
	Targets    []Target
	TargetSets []TargetSet `yaml:"targetSets"`
}

// NewConfig ...
func NewConfig() (*Config, error) {
	return &Config{
		Targets:    []Target{},
		TargetSets: []TargetSet{},
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
	for _, t := range c.TargetSets {
		for _, src := range t.Sources {
			target := Target{}
			target.Source = src
			target.Description = t.Description
			target.Type = t.Type
			target.Regexp = t.Regexp
			target.MultiLine = t.MultiLine
			target.TimeFormat = t.TimeFormat
			target.TimeZone = t.TimeZone
			target.Tags = t.Tags

			u, err := url.Parse(src)
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

func (c *Config) Tags() Tags {
	tags := map[string]int{}
	for _, t := range c.TargetSets {
		for _, tag := range t.Tags {
			if _, ok := tags[tag]; !ok {
				tags[tag] = 0
			}
			tags[tag] = tags[tag] + 1
		}
	}
	return tags
}

func (c *Config) FilterTargets(tagExpr, sourceRe string) ([]Target, error) {
	allTags := c.Tags()
	targets := []Target{}
	tagExpr = strings.Replace(tagExpr, ",", " or ", -1)
	if tagExpr != "" || sourceRe != "" {
		re := regexp.MustCompile(sourceRe)
		for _, target := range c.Targets {
			tags := map[string]interface{}{}
			for tag, _ := range allTags {
				if contains(target.Tags, tag) {
					tags[tag] = true
				} else {
					tags[tag] = false
				}
			}
			out, err := expr.Eval(tagExpr, tags)
			if err != nil {
				return targets, err
			}
			if out.(bool) && (sourceRe == "" || re.MatchString(target.Source)) {
				targets = append(targets, target)
			}
		}
	} else {
		for _, target := range c.Targets {
			targets = append(targets, target)
		}
	}
	return targets, nil
}

func contains(ss []string, t string) bool {
	for _, s := range ss {
		if s == t {
			return true
		}
	}
	return false
}
