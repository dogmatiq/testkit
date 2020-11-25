package testkit

import (
	"io"
	"strings"

	"github.com/dogmatiq/iago/count"
	"github.com/dogmatiq/iago/indent"
	"github.com/dogmatiq/iago/must"
)

const (
	// suggestionsSection is the heading for the section of the test report
	// where suggestions about how to fix failed tests are shown.
	suggestionsSection = "Suggestions"

	// logSection is the heading for the section of the test report where
	// log messages from user-defined expectations are shown.
	logSection = "Log Messages"
)

// Report is a report on the outcome of an expectation.
type Report struct {
	// TreeOk is true if the "tree" that the expectation belongs to passed.
	TreeOk bool

	// Ok is true if this expectation passed.
	Ok bool

	// Criteria is a brief description of the expectation's requirement to pass.
	Criteria string

	// Outcome is a brief description of the outcome of the expectation.
	Outcome string

	// Explanation is a brief description of what actually happened during the
	// test as it relates to this expectation.
	Explanation string

	// Sections contains arbitrary "sections" that are added to the report by
	// the expectation.
	Sections []*ReportSection

	// SubReports contains the reports of any child expectations.
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
	defer must.Recover(&err)
	w := count.NewWriter(next)

	if r.Ok {
		must.WriteString(w, "✓ ")
	} else {
		must.WriteString(w, "✗ ")
	}

	must.WriteString(w, r.Criteria)

	if r.Outcome != "" {
		must.WriteString(w, " (")
		must.WriteString(w, r.Outcome)
		must.WriteByte(w, ')')
	}

	must.WriteByte(w, '\n')

	if len(r.Sections) != 0 || r.Explanation != "" {
		must.WriteByte(w, '\n')

		iw := indent.NewIndenter(w, sectionsIndent)

		if r.Explanation != "" {
			must.WriteString(iw, "EXPLANATION\n")

			must.WriteString(
				iw,
				indent.String(r.Explanation, sectionContentIndent),
			)

			must.WriteByte(iw, '\n')

			if len(r.Sections) != 0 {
				must.WriteByte(iw, '\n')
			}
		}

		for i, s := range r.Sections {
			must.WriteString(iw, strings.ToUpper(s.Title))
			must.WriteString(iw, "\n")

			must.WriteString(
				iw,
				indent.String(
					strings.TrimSpace(s.Content.String()),
					sectionContentIndent,
				),
			)

			must.WriteByte(iw, '\n')

			if i < len(r.Sections)-1 {
				must.WriteByte(iw, '\n')
			}
		}
	}

	if len(r.SubReports) != 0 {
		iw := indent.NewIndenter(w, subReportsIndent)
		for _, sr := range r.SubReports {
			must.WriteTo(iw, sr)
		}
	}

	return int64(w.Count()), nil
}

// ReportSection is a section of a report containing additional information
// about the expectation.
type ReportSection struct {
	Title   string
	Content strings.Builder
}

// Append appends a line of text to the section's content.
func (s *ReportSection) Append(f string, v ...interface{}) {
	must.Fprintf(&s.Content, f+"\n", v...)
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
