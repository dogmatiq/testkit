package report

import "fmt"

// Content is arbitrary content that can be added to various parts of a report.
type Content struct {
	// Heading is a heading for the content.
	//
	// It may be empty. It should be given in lower case without a trailing
	// period, exclamation or question mark, similar to how Go error messages
	// are formatted.
	Heading FormatString

	// Body is the content itself.
	Body string
}

// FormatString captures the arguments for a printf-like function.
type FormatString struct {
	Format    string
	Arguments []interface{}
}

func (s FormatString) String() string {
	return fmt.Sprintf(s.Format, s.Arguments...)
}
