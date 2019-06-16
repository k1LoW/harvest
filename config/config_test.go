package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFilterTargets(t *testing.T) {
	var tests = []struct {
		tagExpr     string
		sourceRe string
		want        int
	}{
		{"webproxy", "", 2},
		{"webproxy and syslog", "", 1},
		{"webproxy or app", "", 5},
		{"!app", "", 5},
		{"webproxy,app", "", 5},
	}

	for _, tt := range tests {
		c, err := NewConfig()
		if err != nil {
			t.Fatalf("%v", err)
		}
		err = c.LoadConfigFile(filepath.Join(testdataDir(), "test_config.yml"))
		if err != nil {
			t.Fatalf("%v", err)
		}
		filtered, err := c.FilterTargets(tt.tagExpr, tt.sourceRe)
		if err != nil {
			t.Fatalf("%v", err)
		}
		got := len(filtered)
		if got != tt.want {
			t.Errorf("\ngot %v\nwant %v", got, tt.want)
		}
	}
}

func testdataDir() string {
	wd, _ := os.Getwd()
	dir, _ := filepath.Abs(filepath.Join(filepath.Dir(wd), "testdata"))
	return dir
}
