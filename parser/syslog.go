package parser

// NewSyslogParser ...
func NewSyslogParser() (Parser, error) {
	r := `^(\w{3} \d{2} \d{2}:\d{2}:\d{2}) .+$`
	tf := "Jan 02 15:04:05"
	return NewRegexpParser(r, tf)
}
