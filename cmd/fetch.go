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
	"regexp"
	"strings"
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

		if dbPath == "" {
			dbPath = fmt.Sprintf("harvest-%s.db", time.Now().Format("20060102T150405-0700"))
		}
		if _, err := os.Lstat(dbPath); err == nil {
			l.Error(fmt.Sprintf("%s already exists", dbPath), zap.String("error", err.Error()))
			os.Exit(1)
		}
		d, err := db.NewDB(ctx, l, dbPath)
		if err != nil {
			l.Error("DB initialize error", zap.String("error", err.Error()))
			os.Exit(1)
		}

		targets := filterTargets(cfg.Targets)
		if len(targets) == 0 {
			l.Error("No targets")
			os.Exit(1)
		}
		l.Info(fmt.Sprintf("Target count: %d", len(targets)))

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

		l.Info(fmt.Sprintf("Client concurrency: %d", concurrency))
		if et != nil {
			l.Info(fmt.Sprintf("Log timestamp: %s - %s", st.Format("2006-01-02 15:04:05-0700"), et.Format("2006-01-02 15:04:05-0700")))
		} else {
			l.Info(fmt.Sprintf("Log timestamp: %s - latest", st.Format("2006-01-02 15:04:05-0700")))
		}

		l.Info("Start fetching from targets.")

		go d.StartInsert()

		cChan := make(chan struct{}, concurrency)
		var wg sync.WaitGroup

		for _, t := range targets {
			wg.Add(1)
			go func(t config.Target) {
				cChan <- struct{}{}
				defer wg.Done()
				c, err := collector.NewCollector(ctx, &t, false)
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

		l.Info("Fetch finished.")
	},
}

// filterTargets ...
func filterTargets(cfgTargets []config.Target) []config.Target {
	targets := []config.Target{}
	ignoreTags := strings.Split(ignoreTag, ",")
	if tag != "" || urlRegexp != "" {
		tags := strings.Split(tag, ",")
		re := regexp.MustCompile(urlRegexp)
		for _, target := range cfgTargets {
			if contains(target.Tags, tags) || (urlRegexp != "" && re.MatchString(target.URL)) {
				if contains(target.Tags, ignoreTags) {
					continue
				}
				targets = append(targets, target)
			}
		}
	} else {
		for _, target := range cfgTargets {
			if contains(target.Tags, ignoreTags) {
				continue
			}
			targets = append(targets, target)
		}
	}
	return targets
}

func setStartTime(stStr string) (*time.Time, error) {
	var st *time.Time
	if stStr != "" {
		loc, err := time.LoadLocation("Local")
		if err != nil {
			return nil, err
		}
		stt, err := time.ParseInLocation("2006-01-02 15:04:05", stStr, loc)
		if err != nil {
			return nil, err
		}
		st = &stt
	} else {
		stt := time.Now().Add(defaultStartTimeDuration)
		st = &stt
	}
	return st, nil
}

// setEndTime ...
func setEndTime(etStr string) (*time.Time, error) {
	var et *time.Time
	if etStr != "" {
		loc, err := time.LoadLocation("Local")
		if err != nil {
			return nil, err
		}
		ett, err := time.ParseInLocation("2006-01-02 15:04:05", etStr, loc)
		if err != nil {
			return nil, err
		}
		et = &ett
	} else {
		et = nil
	}
	return et, nil
}

// contains ...
func contains(ss1 []string, ss2 []string) bool {
	for _, s1 := range ss1 {
		for _, s2 := range ss2 {
			if s1 == s2 {
				return true
			}
		}
	}
	return false
}

func init() {
	rootCmd.AddCommand(fetchCmd)
	fetchCmd.Flags().StringVarP(&dbPath, "out", "o", "", "db path")
	fetchCmd.Flags().StringVarP(&configPath, "config", "c", "", "config file path")
	fetchCmd.Flags().IntVarP(&concurrency, "concurrency", "C", defaultConcurrency, "concurrency")
	fetchCmd.Flags().StringVarP(&tag, "tag", "", "", "filter targets using tag (format: foo,bar)")
	fetchCmd.Flags().StringVarP(&ignoreTag, "ignore-tag", "", "", "ignore targets using tag (format: foo,bar)")
	fetchCmd.Flags().StringVarP(&urlRegexp, "url-regexp", "", "", "filter targets using url regexp")
	fetchCmd.Flags().StringVarP(&stStr, "start-time", "", "", "log start time (default: 1 hours ago) (format: 2006-01-02 15:04:05)")
	fetchCmd.Flags().StringVarP(&etStr, "end-time", "", "", "log end time (default: latest) (format: 2006-01-02 15:04:05)")
}
