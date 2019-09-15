package config

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/antonmedv/expr"
	"github.com/k1LoW/harvest/client/k8s"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// TargetSet ...
type TargetSet struct {
	Sources     []string `yaml:"sources"`
	Description string   `yaml:"description,omitempty"`
	Type        string   `yaml:"type"`
	Regexp      string   `yaml:"regexp,omitempty"`
	MultiLine   bool     `yaml:"multiLine,omitempty"`
	TimeFormat  string   `yaml:"timeFormat,omitempty"`
	TimeZone    string   `yaml:"timeZone,omitempty"`
	Tags        []string `yaml:"tags"`
}

// Target ...
type Target struct {
	Source           string `db:"source"`
	Description      string `db:"description"`
	Type             string `db:"type"`
	Regexp           string `db:"regexp"`
	MultiLine        bool   `db:"multi_line"`
	TimeFormat       string `db:"time_format"`
	TimeZone         string `db:"time_zone"`
	Tags             []string
	Scheme           string `db:"scheme"`
	Host             string `db:"host"`
	User             string `db:"user"`
	Port             int    `db:"port"`
	Path             string `db:"path"`
	SSHKeyPassphrase []byte
	Id               int64 `db:"id"`
}

func (t *Target) GetHostLength() int {
	return len(t.Host)
}

func (t *Target) GetPathLength() (int, error) {
	if t.Scheme == "k8s" {
		contextName := t.Host
		splited := strings.Split(t.Path, "/")
		namespace := splited[1]
		podFilter := regexp.MustCompile(strings.Replace(strings.Replace(splited[2], ".*", "*", -1), "*", ".*", -1))
		containers, err := k8s.GetContainers(contextName, namespace, podFilter)
		if err != nil {
			return 0, err
		}
		length := 0
		for _, c := range containers {
			if length < len(c) {
				length = len(c)
			}
		}
		return length, nil
	}
	return len(t.Path), nil
}

type Tags map[string]int

// Config ...
type Config struct {
	Targets    []*Target    `yaml:"-"`
	TargetSets []*TargetSet `yaml:"targetSets"`
}

// NewConfig ...
func NewConfig() (*Config, error) {
	return &Config{
		Targets:    []*Target{},
		TargetSets: []*TargetSet{},
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
	buf, err := ioutil.ReadFile(filepath.Clean(fullPath))
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
			if target.Host == "" {
				target.Host = "localhost"
			}

			c.Targets = append(c.Targets, &target)
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

func (c *Config) FilterTargets(tagExpr, sourceRe string) ([]*Target, error) {
	allTags := c.Tags()
	targets := []*Target{}
	tagExpr = strings.Replace(tagExpr, ",", " or ", -1)
	for _, target := range c.Targets {
		tags := map[string]interface{}{
			"hrv_source": target.Source,
		}
		for tag := range allTags {
			if contains(target.Tags, tag) {
				tags[tag] = true
			} else {
				tags[tag] = false
			}
		}
		targetExpr := []string{"true"}
		if tagExpr != "" {
			targetExpr = append(targetExpr, fmt.Sprintf("(%s)", tagExpr))
		}
		if sourceRe != "" {
			targetExpr = append(targetExpr, fmt.Sprintf(`(hrv_source matches "%s")`, sourceRe))
		}
		out, err := expr.Eval(strings.Join(targetExpr, " and "), tags)
		if err != nil {
			return targets, err
		}
		if out.(bool) {
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
