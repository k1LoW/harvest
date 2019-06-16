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

	"github.com/k1LoW/harvest/collector"
	"github.com/k1LoW/harvest/config"
	"github.com/k1LoW/harvest/logger"
	"github.com/k1LoW/harvest/parser"
	"github.com/labstack/gommon/color"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// configtestCmd represents the configtest command
var configtestCmd = &cobra.Command{
	Use:   "configtest",
	Short: "configtest",
	Long:  `configtest.`,
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

		l.Info("Test tmestamp parsing")
		fmt.Println("")

		cChan := make(chan struct{}, 1)
		var wg sync.WaitGroup

		failure := 0
		for _, t := range targets {
			wg.Add(1)
			cChan <- struct{}{}
			c, err := collector.NewCollector(ctx, t, true)
			if err != nil {
				failure++
				<-cChan
				wg.Done()
				l.Error("ConfigTest error", zap.String("host", t.Host), zap.String("path", t.Path), zap.String("error", err.Error()))
				continue
			}
			logChan := make(chan parser.Log)
			go func(t *config.Target, logChan chan parser.Log) {
				defer wg.Done()
				fmt.Printf("%s: ", t.Source)
				logRead := false
				for log := range logChan {
					if log.Timestamp > 0 {
						fmt.Printf("%s\n", color.Green("OK", color.B))
					} else if t.Type == "none" {
						fmt.Printf("%s\n", color.Yellow("Skip (because type=none)", color.B))
					} else {
						fmt.Printf("%s\n", color.Red("Timestamp parse error", color.B))
						fmt.Printf("    %s %s\n", color.Red("      Type:"), color.Red(t.Type))
						fmt.Printf("    %s %s\n", color.Red("    Regexp:"), color.Red(t.Regexp))
						fmt.Printf("    %s %s\n", color.Red("TimeFormat:"), color.Red(t.TimeFormat))
						fmt.Printf("    %s %s\n", color.Red(" MultiLine:"), color.Red(t.MultiLine))
						fmt.Printf("    %s %s\n", color.Red("       Log:"), color.Red(log.Content))
						fmt.Println("")
						failure++
					}
					logRead = true
				}
				if !logRead {
					fmt.Printf("%s\n", color.Red("Log read error", color.B))
					failure++
				}
			}(t, logChan)
			err = c.ConfigTest(logChan, t.MultiLine)
			if err != nil {
				failure++
				l.Error("ConfigTest error", zap.String("host", t.Host), zap.String("path", t.Path), zap.String("error", err.Error()))
			}
			<-cChan
		}

		wg.Wait()

		fmt.Println("")
		if failure > 0 {
			fmt.Println(color.Red(fmt.Sprintf("%d targets, %d failure\n", len(targets), failure), color.B))
		} else {
			fmt.Println(color.Green(fmt.Sprintf("%d targets, %d failure\n", len(targets), failure), color.B))
		}
	},
}

func init() {
	rootCmd.AddCommand(configtestCmd)
	configtestCmd.Flags().StringVarP(&configPath, "config", "c", "", "config file path")
	configtestCmd.Flags().StringVarP(&tag, "tag", "", "", "filter targets using tag")
	configtestCmd.Flags().StringVarP(&sourceRe, "source", "", "", "filter targets using source regexp")
	configtestCmd.Flags().BoolVarP(&presetSSHKeyPassphrase, "preset-ssh-key-passphrase", "", false, "preset SSH key passphrase")
}
