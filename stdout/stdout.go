package stdout

import (
	"fmt"
	"strings"
	"time"

	"github.com/k1LoW/harvest/parser"
	"github.com/labstack/gommon/color"
)

const (
	tsParseFmt     = "2006-01-02T15:04:05-07:00"
	tsNanoParseFmt = "2006-01-02T15:04:05.000000000-07:00"
)

var colorizeMap = []struct {
	colorFunc func(interface{}, ...string) string
	bar       string
}{
	{color.Yellow, "█ "},
	{color.Magenta, "█ "},
	{color.Green, "█ "},
	{color.Cyan, "█ "},
	{color.Yellow, "▚ "},
	{color.Magenta, "▚ "},
	{color.Green, "▚ "},
	{color.Cyan, "▚ "},
	{color.Yellow, "║ "},
	{color.Magenta, "║ "},
	{color.Green, "║ "},
	{color.Cyan, "║ "},
	{color.Yellow, "▒ "},
	{color.Magenta, "▒ "},
	{color.Green, "▒ "},
	{color.Cyan, "▒ "},
	{color.Yellow, "▓ "},
	{color.Magenta, "▓ "},
	{color.Green, "▓ "},
	{color.Cyan, "▓ "},
}

// Stdout ...
type Stdout struct {
	withTimestamp     bool
	withTimestampNano bool
	withHost          bool
	withPath          bool
	withTag           bool
	withoutMark       bool
	hFmt              string
	tFmt              string
	noColor           bool
}

// NewStdout ...
func NewStdout(withTimestamp bool,
	withTimestampNano bool,
	withHost bool,
	withPath bool,
	withTag bool,
	withoutMark bool,
	hLen int,
	tLen int,
	noColor bool,
) (*Stdout, error) {
	return &Stdout{
		withTimestamp:     withTimestamp,
		withTimestampNano: withTimestampNano,
		withHost:          withHost,
		withPath:          withPath,
		withTag:           withTag,
		withoutMark:       withoutMark,
		hFmt:              fmt.Sprintf("%%-%ds ", hLen),
		tFmt:              fmt.Sprintf("%%-%ds ", tLen),
		noColor:           noColor,
	}, nil
}

// Out ...
func (s *Stdout) Out(logChan chan parser.Log, hosts []string) error {
	if s.noColor {
		color.Disable()
	} else {
		color.Enable()
	}

	for log := range logChan {
		var (
			bar            string
			ts             string
			filledByPrevTs string
			host           string
			tag            string
		)

		colorFunc := func(msg interface{}, styles ...string) string {
			return msg.(string)
		}

		if s.withTimestamp {
			if log.Timestamp == 0 {
				ts = fmt.Sprintf(fmt.Sprintf("%%-%ds", len(tsParseFmt)), "-")
			} else {
				ts = time.Unix(0, log.Timestamp).Format(tsParseFmt)
			}
		}
		if s.withTimestampNano {
			if log.Timestamp == 0 {
				ts = fmt.Sprintf(fmt.Sprintf("%%-%ds", len(tsNanoParseFmt)), "-")
			} else {
				ts = time.Unix(0, log.Timestamp).Format(tsNanoParseFmt)
			}
		}
		if s.withTimestamp || s.withTimestampNano {
			if log.FilledByPrevTs {
				filledByPrevTs = "* "
			} else {
				filledByPrevTs = "  "
			}
		}

		if s.withHost && s.withPath {
			host = fmt.Sprintf(s.hFmt, fmt.Sprintf("%s:%s", log.Host, log.Path))
		} else if s.withHost {
			host = fmt.Sprintf(s.hFmt, log.Host)
		} else if s.withPath {
			host = fmt.Sprintf(s.hFmt, log.Path)
		}
		if s.withTag {
			tag = fmt.Sprintf(s.tFmt, fmt.Sprintf("%v", log.Target.Tags))
		}

		if s.withTimestamp || s.withTimestampNano || s.withHost || s.withPath {
			for i, h := range hosts {
				if h == log.Host {
					colorFunc = colorizeMap[i%len(colorizeMap)].colorFunc
					if s.withoutMark {
						bar = ""
					} else {
						bar = colorFunc(colorizeMap[i%len(colorizeMap)].bar)
					}
				}
			}
		}

		fmt.Printf("%s%s%s%s%s%s\n", bar, colorFunc(ts), color.White(filledByPrevTs, color.B), colorizeTag(colorFunc, tag), color.Grey(host), log.Content)
	}
	return nil
}

func colorizeTag(colorFunc func(interface{}, ...string) string, tag string) string {
	colorized := []string{}
	tags := strings.Split(tag, " ")
	for _, t := range tags {
		colorized = append(colorized, colorFunc(t, color.B))
	}
	return strings.Join(colorized, " ")
}
