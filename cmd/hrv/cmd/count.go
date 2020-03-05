/*
Copyright © 2019 Ken'ichiro Oyama <k1lowxb@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/k1LoW/harvest/db"
	"github.com/k1LoW/harvest/logger"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	groups    []string
	matches   []string
	delimiter string
)

// countCmd represents the count command
var countCmd = &cobra.Command{
	Use:   "count [DB_FILE]",
	Short: "count logs from harvest-*.db",
	Long:  `count logs from harvest-*.db.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(runCount(args, groups, matches))
	},
}

// runCount ...
func runCount(args, groups, matches []string) int {
	l := logger.NewLogger(verbose)
	dbPath := args[0]

	if _, err := os.Lstat(dbPath); err != nil {
		l.Error(fmt.Sprintf("%s not exists", dbPath), zap.String("error", err.Error()))
		return 1
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	d, err := db.AttachDB(ctx, l, dbPath)
	if err != nil {
		l.Error("DB attach error", zap.String("error", err.Error()))
		return 1
	}

	results, err := d.Count(groups, matches)
	if err != nil {
		l.Error("DB attach error", zap.String("error", err.Error()))
		return 1
	}

	for _, line := range results {
		fmt.Println(strings.Join(line, delimiter))
	}

	return 0
}

func init() {
	countCmd.Flags().StringSliceVarP(&groups, "group-by", "g", []string{}, "group logs using time, host, desctiption, and tag")
	countCmd.Flags().StringSliceVarP(&matches, "match", "m", []string{}, "group logs using SQLite `%LIKE%` query")
	countCmd.Flags().StringVarP(&delimiter, "delimiter", "d", "\t", "delmiter")
	countCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "print debugging messages.")
	err := countCmd.MarkZshCompPositionalArgumentFile(1)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
	rootCmd.AddCommand(countCmd)
}
