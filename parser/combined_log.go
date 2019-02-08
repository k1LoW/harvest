package parser

// NewCombinedLogParser ...
func NewCombinedLogParser() (Parser, error) {
	r := `^[\d\.]+ - [^ ]+ \[(.+)\] .+$`
	tf := "02/Jan/2006:15:04:05 -0700"
	multiLine := false
	return NewRegexpParser(r, tf, multiLine)
}
