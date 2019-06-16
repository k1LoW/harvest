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
	"sync"
	"time"

	"github.com/k1LoW/harvest/collector"
	"github.com/k1LoW/harvest/config"
	"github.com/k1LoW/harvest/logger"
	"github.com/k1LoW/harvest/parser"
	"github.com/k1LoW/harvest/stdout"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// streamCmd represents the stream command
var streamCmd = &cobra.Command{
	Use:   "stream",
	Short: "output stream from targets",
	Long:  `output stream from targets.`,
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

		hLen, tLen, err := getStreamStdoutLengthes(targets, withHost, withPath, withTag)
		if err != nil {
			l.Error("option error", zap.String("error", err.Error()))
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
			l.Error("fetch error", zap.String("error", err.Error()))
			os.Exit(1)
		}

		hosts := getHosts(targets)
		logChan := make(chan parser.Log)

		go sout.Out(logChan, hosts)

		var wg sync.WaitGroup

		for _, t := range targets {
			wg.Add(1)
			go func(t *config.Target) {
				defer wg.Done()
				c, err := collector.NewCollector(ctx, t, false)
				if err != nil {
					l.Error("Stream error", zap.String("host", t.Host), zap.String("path", t.Path), zap.String("error", err.Error()))
				}
				err = c.Stream(logChan, t.MultiLine)
				if err != nil {
					l.Error("Stream error", zap.String("host", t.Host), zap.String("path", t.Path), zap.String("error", err.Error()))
				}
			}(t)
			time.Sleep(100 * time.Millisecond)
		}

		wg.Wait()
	},
}

func getHosts(targets []*config.Target) []string {
	hosts := []string{}
	for _, target := range targets {
		hosts = append(hosts, target.Host)
	}
	return hosts
}

func getStreamStdoutLengthes(targets []*config.Target, withHost, withPath, withTag bool) (int, int, error) {
	var (
		hLen int
		tLen int
	)
	if withHost && withPath {
		hLen = getMaxLength(targets, "hostpath")
	} else if withHost {
		hLen = getMaxLength(targets, "host")
	} else if withPath {
		hLen = getMaxLength(targets, "path")
	}
	if withTag {
		tLen = getMaxLength(targets, "tags")
	}
	return hLen, tLen, nil
}

func getMaxLength(targets []*config.Target, key string) int {
	var length int
	for _, target := range targets {
		var c int
		switch key {
		case "host":
			c = len(target.Host)
		case "path":
			c = len(target.Path)
		case "hostpath":
			c = len(target.Host) + len(target.Path)
		case "tags":
			c = len(fmt.Sprintf("[%s]", strings.Join(target.Tags, "][")))
		}
		if length < c {
			length = c
		}
	}
	return length
}

func init() {
	rootCmd.AddCommand(streamCmd)
	streamCmd.Flags().StringVarP(&configPath, "config", "c", "", "config file path")
	streamCmd.Flags().BoolVarP(&withTimestamp, "with-timestamp", "", false, "output with timestamp")
	streamCmd.Flags().BoolVarP(&withTimestampNano, "with-timestamp-nano", "", false, "output with timestamp nano sec")
	streamCmd.Flags().BoolVarP(&withHost, "with-host", "", false, "output with host")
	streamCmd.Flags().BoolVarP(&withPath, "with-path", "", false, "output with path")
	streamCmd.Flags().BoolVarP(&withTag, "with-tag", "", false, "output with tag")
	streamCmd.Flags().BoolVarP(&withoutMark, "without-mark", "", false, "output without prefix mark")
	streamCmd.Flags().StringVarP(&tag, "tag", "", "", "filter targets using tag (format: foo,bar)")
	streamCmd.Flags().StringVarP(&sourceRe, "source", "", "", "filter targets using source regexp")
	streamCmd.Flags().BoolVarP(&noColor, "no-color", "", false, "disable colorize output")
	streamCmd.Flags().BoolVarP(&presetSSHKeyPassphrase, "preset-ssh-key-passphrase", "", false, "preset SSH key passphrase")
}
