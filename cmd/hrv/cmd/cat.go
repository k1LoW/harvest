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
	"strconv"
	"strings"

	"github.com/antonmedv/expr"
	"github.com/k1LoW/harvest/db"
	"github.com/k1LoW/harvest/logger"
	"github.com/k1LoW/harvest/stdout"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	match string
)

// catCmd represents the cat command
var catCmd = &cobra.Command{
	Use:   "cat [DB_FILE]",
	Short: "cat logs from harvest-*.db",
	Long:  `cat logs from harvest-*.db`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		l := logger.NewLogger(verbose)
		dbPath := args[0]

		if _, err := os.Lstat(dbPath); err != nil {
			l.Error(fmt.Sprintf("%s not exists", dbPath), zap.String("error", err.Error()))
			os.Exit(1)
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		d, err := db.AttachDB(ctx, l, dbPath)
		if err != nil {
			l.Error("DB attach error", zap.String("error", err.Error()))
			os.Exit(1)
		}

		cond, err := buildCondition(d)
		if err != nil {
			l.Error("option error", zap.String("error", err.Error()))
			os.Exit(1)
		}

		hLen, tLen, err := getCatStdoutLengthes(d, withHost, withPath, withTag)
		if err != nil {
			l.Error("option error", zap.String("error", err.Error()))
			os.Exit(1)
		}

		hosts, err := d.GetHosts()
		if err != nil {
			l.Error("DB query error", zap.String("error", err.Error()))
			os.Exit(1)
		}

		sout, err := stdout.NewStdout(
			withTimestamp,
			withTimestampNano,
			withHost,
			withPath,
			withTag,
			withoutMark,
			hLen,
			tLen,
			noColor,
		)
		if err != nil {
			l.Error("cat error", zap.String("error", err.Error()))
			os.Exit(1)
		}

		sout.Out(d.Cat(cond), hosts)
	},
}

func buildCondition(db *db.DB) (string, error) {
	matchCond := []string{}
	cond := []string{}
	if match != "" {
		matchCond = append(matchCond, match)
	}

	if stStr != "" || etStr != "" || duStr != "" {
		st, et, err := parseTimes(stStr, etStr, duStr)
		if err != nil {
			return "", err
		}
		cond = append(cond, fmt.Sprintf("ts_unixnano >= %d", st.UnixNano()))
		cond = append(cond, fmt.Sprintf("ts_unixnano <= %d", et.UnixNano()))
	}

	if tag != "" {
		tagExpr := strings.Replace(tag, ",", " or ", -1)
		allTags, err := db.GetTags()
		if err != nil {
			return "", err
		}
		tt, err := db.GetTargetIdAndTags()
		if err != nil {
			return "", err
		}
		targetIds := []string{}
		for targetId, targetTags := range tt {
			targetExpr := []string{"true"}
			targetExpr = append(targetExpr, fmt.Sprintf("(%s)", tagExpr))
			tags := map[string]interface{}{}
			for _, tag := range allTags {
				if contains(targetTags, tag) {
					tags[tag] = true
				} else {
					tags[tag] = false
				}
			}
			out, err := expr.Eval(strings.Join(targetExpr, " and "), tags)
			if err != nil {
				return "", err
			}
			if out.(bool) {
				targetIds = append(targetIds, strconv.FormatInt(targetId, 10))
			}
		}
		cond = append(cond, fmt.Sprintf("( target_id IN (%s) )", strings.Join(targetIds, ", ")))
	}

	if len(matchCond) > 0 {
		cond = append(cond, fmt.Sprintf("content MATCH '%s'", strings.Join(matchCond, " AND ")))
	}

	if len(cond) == 0 {
		return "", nil
	}

	return fmt.Sprintf(" WHERE %s", strings.Join(cond, " AND ")), nil // #nosec
}

func getCatStdoutLengthes(d *db.DB, withHost, withPath, withTag bool) (int, int, error) {
	var (
		hLen int
		tLen int
		err  error
	)
	if withHost && withPath {
		hLen, err = d.GetColumnMaxLength("host", "path")
		if err != nil {
			return 0, 0, err
		}
	} else if withHost {
		hLen, err = d.GetColumnMaxLength("host")
		if err != nil {
			return 0, 0, err
		}
	} else if withPath {
		hLen, err = d.GetColumnMaxLength("path")
		if err != nil {
			return 0, 0, err
		}
	}
	if withTag {
		tLen, err = d.GetTagMaxLength()
		if err != nil {
			return 0, 0, err
		}
	}
	return hLen, tLen, nil
}

func contains(ss []string, t string) bool {
	for _, s := range ss {
		if s == t {
			return true
		}
	}
	return false
}

func init() {
	rootCmd.AddCommand(catCmd)
	catCmd.Flags().BoolVarP(&withTimestamp, "with-timestamp", "", false, "output with timestamp")
	catCmd.Flags().BoolVarP(&withTimestampNano, "with-timestamp-nano", "", false, "output with timestamp nano sec")
	catCmd.Flags().BoolVarP(&withHost, "with-host", "", false, "output with host")
	catCmd.Flags().BoolVarP(&withPath, "with-path", "", false, "output with path")
	catCmd.Flags().BoolVarP(&withTag, "with-tag", "", false, "output with tag")
	catCmd.Flags().BoolVarP(&withoutMark, "without-mark", "", false, "output without prefix mark")
	catCmd.Flags().StringVarP(&match, "match", "", "", "filter logs using SQLite FTS `MATCH` query")
	catCmd.Flags().StringVarP(&tag, "tag", "", "", "filter logs using tag")
	catCmd.Flags().StringVarP(&stStr, "start-time", "", "", "log start time (format: 2006-01-02 15:04:05)")
	catCmd.Flags().StringVarP(&etStr, "end-time", "", "", "log end time (format: 2006-01-02 15:04:05)")
	catCmd.Flags().StringVarP(&duStr, "duration", "", "", "log duration")
	catCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "print debugging messages.")
	catCmd.Flags().BoolVarP(&noColor, "no-color", "", false, "disable colorize output")
	err := catCmd.MarkZshCompPositionalArgumentFile(1)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
