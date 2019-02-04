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
	"github.com/logrusorgru/aurora"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	withTimestamp     bool
	withTimestampNano bool
	withHost          bool
	withPath          bool
	match             string
	st                string
	et                string
	noColors          bool
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

		var hFmt string
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

		au := aurora.NewAurora(!noColors)

		for log := range d.Cat(cond) {
			var (
				ts   string
				host string
			)
			if withTimestamp {
				ts = fmt.Sprintf("%s ", time.Unix(0, log.Timestamp).Format("2006-01-02T15:04:05-07:00"))
			}
			if withTimestampNano {
				ts = fmt.Sprintf("%s ", time.Unix(0, log.Timestamp).Format("2006-01-02T15:04:05.000000000-07:00"))
			}
			if withHost && withPath {
				host = fmt.Sprintf(hFmt, fmt.Sprintf("%s:%s", log.Host, log.Path))
			} else if withHost {
				host = fmt.Sprintf(hFmt, log.Host)
			} else if withPath {
				host = fmt.Sprintf(hFmt, log.Path)
			}

			fmt.Printf("%s%s%s\n", au.Brown(ts), au.Gray(host), log.Content)
		}
	},
}

// buildCondition ...
func buildCondition() (string, error) {
	cond := []string{}
	if match != "" {
		cond = append(cond, fmt.Sprintf("content MATCH '%s'", match))
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
	if len(cond) == 0 {
		return "", nil
	}
	return fmt.Sprintf(" WHERE %s", strings.Join(cond, " AND ")), nil
}

func init() {
	rootCmd.AddCommand(catCmd)
	catCmd.Flags().BoolVarP(&withTimestamp, "with-ts", "", false, "with timestamp")
	catCmd.Flags().BoolVarP(&withTimestampNano, "with-ts-nano", "", false, "with timestamp nano sec")
	catCmd.Flags().BoolVarP(&withHost, "with-host", "", false, "with host")
	catCmd.Flags().BoolVarP(&withPath, "with-path", "", false, "with path")
	catCmd.Flags().StringVarP(&match, "match", "", "", "MATCH Query")
	catCmd.Flags().StringVarP(&st, "start-time", "", "", "start time")
	catCmd.Flags().StringVarP(&et, "end-time", "", "", "end time")
	catCmd.Flags().BoolVarP(&noColors, "no-colors", "", false, "disable colors")
}
