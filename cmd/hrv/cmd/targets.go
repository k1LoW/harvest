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
	"fmt"
	"os"

	"github.com/k1LoW/harvest/config"
	"github.com/k1LoW/harvest/logger"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// targetsCmd represents the targets command
var targetsCmd = &cobra.Command{
	Use:   "targets",
	Short: "list targets",
	Long:  `list targets.`,
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
		targets, err := cfg.FilterTargets(tag, sourceRe)
		if err != nil {
			l.Error("tag option error", zap.String("error", err.Error()))
			os.Exit(1)
		}
		if len(targets) == 0 {
			l.Error("No targets")
			os.Exit(1)
		}
		for _, t := range targets {
			host := t.Host
			if host == "" {
				host = "localhost"
			}
			fmt.Printf("%s:%s\n", host, t.Path)
		}
	},
}

func init() {
	rootCmd.AddCommand(targetsCmd)
	targetsCmd.Flags().StringVarP(&configPath, "config", "c", "", "config file path")
	targetsCmd.Flags().StringVarP(&tag, "tag", "", "", "filter targets using tag (format: foo,bar)")
	targetsCmd.Flags().StringVarP(&sourceRe, "source", "", "", "filter targets using source regexp")
}
