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
	"sync"
	"time"

	"github.com/k1LoW/harvest/collector"
	"github.com/k1LoW/harvest/config"
	"github.com/k1LoW/harvest/db"
	"github.com/k1LoW/harvest/logger"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	stStr       string
	etStr       string
	dbPath      string
	concurrency int
)

const (
	defaultConcurrency       = 10
	defaultStartTimeDuration = -1 * time.Hour
)

// fetchCmd represents the fetch command
var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "fetch from targets",
	Long:  `fetch from targets.`,
	Run: func(cmd *cobra.Command, args []string) {
		l := logger.NewLogger(verbose)

		cfg, err := config.NewConfig()
		if err != nil {
			l.Error("Config error", zap.String("error", err.Error()))
			os.Exit(1)
		}
		err = cfg.LoadConfigFile(configPath)
		if err != nil {
			l.Error("Config error", zap.String("error", err.Error()))
			os.Exit(1)
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		if dbPath == "" {
			dbPath = fmt.Sprintf("harvest-%s.db", time.Now().Format("20060102T150405-0700"))
		}
		if _, err := os.Lstat(dbPath); err == nil {
			l.Error(fmt.Sprintf("%s already exists", dbPath))
			os.Exit(1)
		}
		d, err := db.NewDB(ctx, l, cfg, dbPath)
		if err != nil {
			l.Error("DB initialize error", zap.String("error", err.Error()))
			os.Exit(1)
		}

		targets, err := cfg.FilterTargets(tag, sourceRe)
		if err != nil {
			l.Error("tag option error", zap.String("error", err.Error()))
			os.Exit(1)
		}
		if len(targets) == 0 {
			l.Error("No targets")
			os.Exit(1)
		}
		l.Info(fmt.Sprintf("Target count: %d", len(targets)))

		if presetSSHKeyPassphrase {
			err = presetSSHKeyPassphraseToTargets(targets)
			if err != nil {
				l.Error("option error", zap.String("error", err.Error()))
				os.Exit(1)
			}
		}

		st, err := setStartTime(stStr)
		if err != nil {
			l.Error("option error", zap.String("error", err.Error()))
			os.Exit(1)
		}

		et, err := setEndTime(etStr)
		if err != nil {
			l.Error("option error", zap.String("error", err.Error()))
			os.Exit(1)
		}

		l.Debug(fmt.Sprintf("Client concurrency: %d", concurrency))
		if et != nil {
			l.Info(fmt.Sprintf("Log timestamp: %s - %s", st.Format("2006-01-02 15:04:05-0700"), et.Format("2006-01-02 15:04:05-0700")))
		} else {
			l.Info(fmt.Sprintf("Log timestamp: %s - latest", st.Format("2006-01-02 15:04:05-0700")))
		}

		l.Debug("Start fetching from targets")

		go d.StartInsert()

		cChan := make(chan struct{}, concurrency)
		var wg sync.WaitGroup

		for _, t := range targets {
			wg.Add(1)
			go func(t *config.Target) {
				cChan <- struct{}{}
				defer wg.Done()
				c, err := collector.NewCollector(ctx, t, l)
				if err != nil {
					l.Error("Fetch error", zap.String("host", t.Host), zap.String("path", t.Path), zap.String("error", err.Error()))
				}
				err = c.Fetch(d.In(), st, et, t.MultiLine)
				if err != nil {
					l.Error("Fetch error", zap.String("host", t.Host), zap.String("path", t.Path), zap.String("error", err.Error()))
				}
				<-cChan
			}(t)
		}

		wg.Wait()

		l.Debug("Fetch finished")
	},
}

func init() {
	rootCmd.AddCommand(fetchCmd)
	fetchCmd.Flags().StringVarP(&dbPath, "out", "o", "", "db path")
	fetchCmd.Flags().StringVarP(&configPath, "config", "c", "", "config file path")
	fetchCmd.Flags().IntVarP(&concurrency, "concurrency", "C", defaultConcurrency, "concurrency")
	fetchCmd.Flags().StringVarP(&tag, "tag", "", "", "filter targets using tag")
	fetchCmd.Flags().StringVarP(&sourceRe, "source", "", "", "filter targets using source regexp")
	fetchCmd.Flags().StringVarP(&stStr, "start-time", "", "", "log start time (default: 1 hours ago) (format: 2006-01-02 15:04:05)")
	fetchCmd.Flags().StringVarP(&etStr, "end-time", "", "", "log end time (default: latest) (format: 2006-01-02 15:04:05)")
	fetchCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "print debugging messages.")
	fetchCmd.Flags().BoolVarP(&presetSSHKeyPassphrase, "preset-ssh-key-passphrase", "", false, "preset SSH key passphrase")
}
