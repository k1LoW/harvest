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
	"time"

	"github.com/Songmu/prompter"
	"github.com/k1LoW/harvest/config"
	"github.com/spf13/cobra"
)

var (
	tag                    string
	configPath             string
	sourceRe               string
	withTimestamp          bool
	withTimestampNano      bool
	withHost               bool
	withPath               bool
	withTag                bool
	withoutMark            bool
	noColor                bool
	presetSSHKeyPassphrase bool
	verbose                bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "hrv",
	Short: "Portable log aggregation tool for middle-scale system operation/troubleshooting",
	Long:  `Portable log aggregation tool for middle-scale system operation/troubleshooting.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {}

type hostPassphrase struct {
	host       string
	passphrase []byte
}

func presetSSHKeyPassphraseToTargets(targets []*config.Target) error {
	hpMap := map[string]hostPassphrase{}
	var defaultPassohrase []byte

	yn := prompter.YN("Do you preset default passphrase for all targets?", true)
	if yn {
		fmt.Println("Preset default passphrase")
		defaultPassohrase = []byte(prompter.Password("Enter default passphrase"))
	} else {
		fmt.Println("Preset passphrase for each target")
	}

	for i, target := range targets {
		if target.Scheme != "ssh" {
			continue
		}
		if yn {
			targets[i].SSHKeyPassphrase = defaultPassohrase
			continue
		}
		if hp, ok := hpMap[target.Host]; ok {
			targets[i].SSHKeyPassphrase = hp.passphrase
			continue
		}
		passphrase := []byte(prompter.Password(fmt.Sprintf("Enter passphrase for host '%s'", target.Host)))
		targets[i].SSHKeyPassphrase = passphrase
		hpMap[target.Host] = hostPassphrase{
			host:       target.Host,
			passphrase: passphrase,
		}
	}
	return nil
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
