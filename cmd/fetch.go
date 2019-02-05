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
	dbPath      string
	configPath  string
	concurrency int
)

// fetchCmd represents the fetch command
var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "fetch from targets",
	Long:  `fetch from targets.`,
	Run: func(cmd *cobra.Command, args []string) {
		l := logger.NewLogger()

		cfg, err := config.NewConfig()
		if err != nil {
			l.Error("Config error", zap.Error(err))
			os.Exit(1)
		}
		err = cfg.LoadConfigFile(configPath)
		if err != nil {
			l.Error("Config error", zap.Error(err))
			os.Exit(1)
		}

		if _, err := os.Lstat(dbPath); err == nil {
			l.Error(fmt.Sprintf("%s already exists", dbPath), zap.Error(err))
			os.Exit(1)
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		d, err := db.NewDB(ctx, l, dbPath)
		if err != nil {
			l.Error("DB initialize error", zap.Error(err))
			os.Exit(1)
		}

		l.Info("Start fetching from targets.")

		var st time.Time
		if stStr != "" {
			loc, err := time.LoadLocation("Local")
			if err != nil {
				l.Error("option error", zap.Error(err))
				os.Exit(1)
			}
			st, err = time.ParseInLocation("2006-01-02 15:04:05", stStr, loc)
			if err != nil {
				l.Error("option error", zap.Error(err))
				os.Exit(1)
			}
		} else {
			st = time.Now().Add(-time.Hour * 6)
		}

		go d.StartInsert()

		cChan := make(chan struct{}, concurrency)
		var wg sync.WaitGroup

		for _, t := range cfg.Targets {
			wg.Add(1)
			go func(t config.Target) {
				cChan <- struct{}{}
				defer wg.Done()
				c, err := collector.NewCollector(ctx, &t)
				if err != nil {
					l.Error("Fetch error", zap.String("host", t.Host), zap.String("path", t.Path), zap.Error(err))
				}
				err = c.Collect(d.In(), st)
				if err != nil {
					l.Error("Fetch error", zap.String("host", t.Host), zap.String("path", t.Path), zap.Error(err))
				}
				<-cChan
			}(t)
		}

		wg.Wait()

		l.Info("Fetch finished.")
	},
}

func init() {
	rootCmd.AddCommand(fetchCmd)
	fetchCmd.Flags().StringVarP(&dbPath, "out", "o", "harvest.db", "db path")
	fetchCmd.Flags().StringVarP(&configPath, "config", "c", "", "config file path")
	fetchCmd.Flags().IntVarP(&concurrency, "concurrency", "C", 5, "concurrency")
	fetchCmd.Flags().StringVarP(&stStr, "start-time", "", "", "log start time (defalt: 6 hours ago) (format: 2006-01-02 15:04:05)")
}
