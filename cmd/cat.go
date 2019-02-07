// Copyright © 2019 Ken'ichiro Oyama <k1lowxb@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/k1LoW/harvest/db"
	"github.com/k1LoW/harvest/logger"
	"github.com/labstack/gommon/color"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

const (
	tsParseFmt     = "2006-01-02T15:04:05-07:00"
	tsNanoParseFmt = "2006-01-02T15:04:05.000000000-07:00"
	delimiter      = ","
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

var (
	withTimestamp     bool
	withTimestampNano bool
	withHost          bool
	withPath          bool
	withTag           bool
	match             string
	tag               string
	st                string
	et                string
	noColor           bool
)

// catCmd represents the cat command
var catCmd = &cobra.Command{
	Use:   "cat [DB_FILE]",
	Short: "cat",
	Long:  `cat`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.WithStack(errors.New("requires one arg"))
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		l := logger.NewLogger()
		dbPath := args[0]

		if _, err := os.Lstat(dbPath); err != nil {
			l.Error(fmt.Sprintf("%s not exists", dbPath), zap.Error(err))
			os.Exit(1)
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		d, err := db.AttachDB(ctx, l, dbPath)
		if err != nil {
			l.Error("DB attach error", zap.Error(err))
			os.Exit(1)
		}

		cond, err := buildCondition()
		if err != nil {
			l.Error("option error", zap.Error(err))
			os.Exit(1)
		}

		var (
			hFmt string
			tFmt string
		)
		if withHost && withPath {
			hLen, err := d.GetColumnMaxLength("host", "path")
			if err != nil {
				l.Error("option error", zap.Error(err))
				os.Exit(1)
			}
			hFmt = fmt.Sprintf("%%-%ds ", hLen)
		} else if withHost {
			hLen, err := d.GetColumnMaxLength("host")
			if err != nil {
				l.Error("option error", zap.Error(err))
				os.Exit(1)
			}
			hFmt = fmt.Sprintf("%%-%ds ", hLen)
		} else if withPath {
			hLen, err := d.GetColumnMaxLength("path")
			if err != nil {
				l.Error("option error", zap.Error(err))
				os.Exit(1)
			}
			hFmt = fmt.Sprintf("%%-%ds ", hLen)
		}

		if withTag {
			tLen, err := d.GetColumnMaxLength("tag")
			if err != nil {
				l.Error("option error", zap.Error(err))
				os.Exit(1)
			}
			tFmt = fmt.Sprintf("%%-%ds ", tLen)
		}

		if noColor {
			color.Disable()
		} else {
			color.Enable()
		}

		hosts, err := d.GetHosts()
		if err != nil {
			l.Error("DB query error", zap.Error(err))
			os.Exit(1)
		}

		for log := range d.Cat(cond) {
			var (
				bar          string
				ts           string
				filledByPrev string
				host         string
				tag          string
			)

			colorFunc := func(msg interface{}, styles ...string) string {
				return msg.(string)
			}

			if withTimestamp {
				if log.Timestamp == 0 {
					ts = fmt.Sprintf(fmt.Sprintf("%%-%ds", len(tsParseFmt)), "-")
				} else {
					ts = time.Unix(0, log.Timestamp).Format(tsParseFmt)
				}
			}
			if withTimestampNano {
				if log.Timestamp == 0 {
					ts = fmt.Sprintf(fmt.Sprintf("%%-%ds", len(tsNanoParseFmt)), "-")
				} else {
					ts = time.Unix(0, log.Timestamp).Format(tsNanoParseFmt)
				}
			}
			if withTimestamp || withTimestampNano {
				if log.FilledByPrev {
					filledByPrev = "* "
				} else {
					filledByPrev = "  "
				}
			}

			if withHost && withPath {
				host = fmt.Sprintf(hFmt, fmt.Sprintf("%s:%s", log.Host, log.Path))
			} else if withHost {
				host = fmt.Sprintf(hFmt, log.Host)
			} else if withPath {
				host = fmt.Sprintf(hFmt, log.Path)
			}
			if withTag {
				tag = fmt.Sprintf(tFmt, log.Tag)
			}

			if withTimestamp || withTimestampNano || withHost || withPath {
				for i, h := range hosts {
					if h == log.Host {
						colorFunc = colorizeMap[i%len(colorizeMap)].colorFunc
						bar = colorFunc(colorizeMap[i%len(colorizeMap)].bar)
					}
				}
			}

			fmt.Printf("%s%s%s%s%s%s\n", bar, colorFunc(ts), color.White(filledByPrev, color.B), colorizeTag(colorFunc, tag), color.Grey(host), log.Content)
		}
	},
}

func colorizeTag(colorFunc func(interface{}, ...string) string, tag string) string {
	colorized := []string{}
	tags := strings.Split(tag, " ")
	for _, t := range tags {
		colorized = append(colorized, colorFunc(t, color.B))
	}
	return strings.Join(colorized, " ")
}

// buildCondition ...
func buildCondition() (string, error) {
	matchCond := []string{}
	cond := []string{}
	if match != "" {
		matchCond = append(matchCond, match)
	}
	loc, err := time.LoadLocation("Local")
	if err != nil {
		return "", err
	}
	if st != "" {
		t, err := time.ParseInLocation("2006-01-02 15:04:05", st, loc)
		if err != nil {
			return "", err
		}
		cond = append(cond, fmt.Sprintf("ts >= %d", t.UnixNano()))
	}
	if et != "" {
		t, err := time.ParseInLocation("2006-01-02 15:04:05", et, loc)
		if err != nil {
			return "", err
		}
		cond = append(cond, fmt.Sprintf("ts <= %d", t.UnixNano()))
	}
	if tag != "" {
		cond = append(cond, fmt.Sprintf("( tag LIKE '%%[%s]%%' )", strings.Join(strings.Split(tag, delimiter), "]%' OR tag LIKE '%[")))
	}

	if len(matchCond) > 0 {
		cond = append(cond, fmt.Sprintf("content MATCH '%s'", strings.Join(matchCond, " AND ")))
	}

	if len(cond) == 0 {
		return "", nil
	}

	return fmt.Sprintf(" WHERE %s", strings.Join(cond, " AND ")), nil
}

func init() {
	rootCmd.AddCommand(catCmd)
	catCmd.Flags().BoolVarP(&withTimestamp, "with-timestamp", "", false, "output with timestamp")
	catCmd.Flags().BoolVarP(&withTimestampNano, "with-timestamp-nano", "", false, "output with timestamp nano sec")
	catCmd.Flags().BoolVarP(&withHost, "with-host", "", false, "output with host")
	catCmd.Flags().BoolVarP(&withPath, "with-path", "", false, "output with path")
	catCmd.Flags().BoolVarP(&withTag, "with-tag", "", false, "output with tag")
	catCmd.Flags().StringVarP(&match, "match", "", "", "filter logs using SQLite FTS `MATCH` query")
	catCmd.Flags().StringVarP(&tag, "tag", "", "", "filter logs using tag")
	catCmd.Flags().StringVarP(&st, "start-time", "", "", "log start time (format: 2006-01-02 15:04:05)")
	catCmd.Flags().StringVarP(&et, "end-time", "", "", "log end time (format: 2006-01-02 15:04:05)")
	catCmd.Flags().BoolVarP(&noColor, "no-color", "", false, "disable colorize output")
}
