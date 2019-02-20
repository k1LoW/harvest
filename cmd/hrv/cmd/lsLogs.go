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
	"path/filepath"
	"sync"
	"time"

	"github.com/k1LoW/harvest/collector"
	"github.com/k1LoW/harvest/config"
	"github.com/k1LoW/harvest/logger"
	"github.com/k1LoW/harvest/parser"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// lsLogsCmd represents the lsLogs command
var lsLogsCmd = &cobra.Command{
	Use:   "ls-logs",
	Short: "list target logs",
	Long:  `list target logs.`,
	Run: func(cmd *cobra.Command, args []string) {
		l := logger.NewLogger()

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

		if dstDir == "" {
			dstDir = fmt.Sprintf("harvest-%s", time.Now().Format("20060102T150405-0700"))
		}
		dstDir, err = filepath.Abs(dstDir)
		if err != nil {
			l.Error("option error", zap.String("error", err.Error()))
			os.Exit(1)
		}

		if _, err := os.Lstat(dstDir); err == nil {
			l.Error(fmt.Sprintf("%s already exists", dstDir), zap.String("error", err.Error()))
			os.Exit(1)
		}

		targets := filterTargets(cfg.Targets)
		if len(targets) == 0 {
			l.Error("No targets")
			os.Exit(1)
		}

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

		logChan := make(chan parser.Log)

		waiter := make(chan struct{})

		go func() {
			defer func() {
				waiter <- struct{}{}
			}()
			for log := range logChan {
				fmt.Printf("%s:%s\n", log.Host, log.Content)
			}
		}()

		cChan := make(chan struct{}, concurrency)
		var wg sync.WaitGroup

		for _, t := range targets {
			wg.Add(1)
			go func(t config.Target) {
				cChan <- struct{}{}
				defer wg.Done()
				c, err := collector.NewCollector(ctx, &t, true)
				if err != nil {
					l.Error("Copy error", zap.String("host", t.Host), zap.String("path", t.Path), zap.String("error", err.Error()))
				}
				err = c.LsLogs(logChan, st, et)
				if err != nil {
					l.Error("Copy error", zap.String("host", t.Host), zap.String("path", t.Path), zap.String("error", err.Error()))
				}
				<-cChan
			}(t)
		}

		wg.Wait()
		close(logChan)
		<-waiter
	},
}

func init() {
	rootCmd.AddCommand(lsLogsCmd)
	lsLogsCmd.Flags().StringVarP(&configPath, "config", "c", "", "config file path")
	lsLogsCmd.Flags().IntVarP(&concurrency, "concurrency", "C", defaultConcurrency, "concurrency")
	lsLogsCmd.Flags().StringVarP(&tag, "tag", "", "", "filter targets using tag (format: foo,bar)")
	lsLogsCmd.Flags().StringVarP(&ignoreTag, "ignore-tag", "", "", "ignore targets using tag (format: foo,bar)")
	lsLogsCmd.Flags().StringVarP(&urlRegexp, "url-regexp", "", "", "filter targets using url regexp")
	lsLogsCmd.Flags().StringVarP(&stStr, "start-time", "", "", "log start time (default: 1 hours ago) (format: 2006-01-02 15:04:05)")
	lsLogsCmd.Flags().StringVarP(&etStr, "end-time", "", "", "log end time (default: latest) (format: 2006-01-02 15:04:05)")
	lsLogsCmd.Flags().BoolVarP(&presetSSHKeyPassphrase, "preset-ssh-key-passphrase", "", false, "preset SSH key passphrase")
}
