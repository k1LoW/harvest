package parser

// NewSyslogParser ...
func NewSyslogParser() (Parser, error) {
	r := `^(\w{3}  ?\d{1,2} \d{2}:\d{2}:\d{2}) .+$`
	tf := "Jan 2 15:04:05"
	multiLine := false
	return NewRegexpParser(r, tf, multiLine)
}
