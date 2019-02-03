package parser

import (
	"testing"
	"time"
)

var parseTimeTests = []struct {
	tf      string
	tz      string
	content string
	want    string
}{
	{
		"Jan 02 15:04:05",
		"+0900",
		"Mar 05 23:59:59",
		"2019-03-05T23:59:59.000000000 +09:00",
	},
	{
		"Jan 02 15:04:05",
		"+0000",
		"Mar 05 23:59:59",
		"2019-03-05T23:59:59.000000000 +00:00",
	},
	{
		"Jan 02 15:04:05",
		"",
		"Mar 05 23:59:59",
		"2019-03-05T23:59:59.000000000 +00:00",
	},
	{
		"02/Jan/2006:15:04:05 -0700",
		"",
		"04/Feb/2019:00:13:49 +0900",
		"2019-02-04T00:13:49.000000000 +09:00",
	},
	{
		"02/Jan/2006:15:04:05 -0700",
		"+0500",
		"04/Feb/2019:00:13:49 +0900",
		"2019-02-04T00:13:49.000000000 +09:00",
	},
}

func TestParseTime(t *testing.T) {
	for _, tt := range parseTimeTests {
		got, err := parseTime(tt.tf, tt.tz, tt.content)
		if err != nil {
			t.Errorf("%v", err)
		}
		want, err := time.Parse("2006-01-02T15:04:05.999999999 -07:00", tt.want)
		if err != nil {
			t.Errorf("%v", err)
		}
		if got.UnixNano() != want.UnixNano() {
			t.Errorf("\ngot %s\nwant %s", got, want)
		}
	}
}
