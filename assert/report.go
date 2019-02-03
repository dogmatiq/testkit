package assert

import (
	"io"
	"strings"

	"github.com/dogmatiq/iago"
	"github.com/dogmatiq/iago/count"
	"github.com/dogmatiq/iago/indent"
)

// Report is a report on the outcome of an assertion.
type Report struct {
	// TreeOk is true if the "tree" that the assertion belongs to passed.
	TreeOk bool

	// Ok is true if this assertion passed.
	Ok bool

	// Criteria is a brief description of the assertion's requirement to pass.
	Criteria string

	// Outcome is a brief description of the outcome of the assertion.
	Outcome string

	// Explanation is a brief description of what actually happened during the test
	// as it relates to this assertion.
	Explanation string

	// Sections contains arbitrary "sections" that are added to the report by the
	// assertion.
	Sections []*ReportSection

	// SubReports contains the reports of any sub-assertions.
	SubReports []*Report
}

// Section adds an arbitrary "section" to the report.
func (r *Report) Section(title string) *ReportSection {
	for _, s := range r.Sections {
		if s.Title == title {
			return s
		}
	}

	s := &ReportSection{
		Title: title,
	}

	r.Sections = append(r.Sections, s)

	return s
}

// Append adds sr as a sub-report of s.
func (r *Report) Append(sr *Report) {
	r.SubReports = append(r.SubReports, sr)
}

// WriteTo writes the report to the given writer.
func (r *Report) WriteTo(next io.Writer) (_ int64, err error) {
	defer iago.Recover(&err)
	w := count.NewWriter(next)

	writeIcon(w, r.Ok)

	iago.MustWriteByte(w, ' ')

	iago.MustWriteString(w, r.Criteria)

	if r.Outcome != "" {
		iago.MustWriteString(w, " (")
		iago.MustWriteString(w, r.Outcome)
		iago.MustWriteByte(w, ')')
	}

	iago.MustWriteByte(w, '\n')

	if len(r.Sections) != 0 || r.Explanation != "" {
		iago.MustWriteByte(w, '\n')

		iw := indent.NewIndenter(w, sectionsIndent)

		if r.Explanation != "" {
			iago.MustWriteString(iw, "EXPLANATION\n")

			iago.MustWriteString(
				iw,
				indent.String(r.Explanation, sectionContentIndent),
			)

			iago.MustWriteByte(iw, '\n')

			if len(r.Sections) != 0 {
				iago.MustWriteByte(iw, '\n')
			}
		}

		for i, s := range r.Sections {
			iago.MustWriteString(iw, strings.ToUpper(s.Title))
			iago.MustWriteString(iw, "\n")

			iago.MustWriteString(
				iw,
				indent.String(
					strings.TrimSpace(s.Content.String()),
					sectionContentIndent,
				),
			)

			iago.MustWriteByte(iw, '\n')

			if i < len(r.Sections)-1 {
				iago.MustWriteByte(iw, '\n')
			}
		}

		iago.MustWriteByte(w, '\n')
	}

	if len(r.SubReports) != 0 {
		iw := indent.NewIndenter(w, subReportsIndent)
		for _, sr := range r.SubReports {
			iago.MustWriteTo(iw, sr)
		}
	}

	return int64(w.Count()), nil
}

// ReportSection is a section of a report containing additional information
// about the assertion.
type ReportSection struct {
	Title   string
	Content strings.Builder
}

// Append appends a line of text to the section's content.
func (s *ReportSection) Append(f string, v ...interface{}) {
	iago.MustFprintf(&s.Content, f+"\n", v...)
}

// AppendListItem appends a line of text prefixed with a bullet.
func (s *ReportSection) AppendListItem(f string, v ...interface{}) {
	s.Append("• "+f, v...)
}

var (
	sectionsIndent       = []byte("  | ")
	sectionContentIndent = "    "
	subReportsIndent     = []byte("    ")
)

const (
	suggestionsSection     = "Suggestions"
	messageDiffSection     = "Message Diff"
	messageTypeDiffSection = "Message Type Diff"
)
