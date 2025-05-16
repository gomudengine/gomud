package markdown

import (
	"strconv"
	"strings"
)

// Formats into HTML tags
//
// Expected ansitags color aliases:
// md
// md-bg
// md-p
// md-p-bg
// md-h1-prefix
// md-h1-prefix-bg
// md-h1, md-h2, md-h3 etc.
// md-h1-bg, md-h2-bg, md-h3-bg etc.
// md-li
// md-li-bg
// md-bold
// md-bold-bg
// md-em
// md-em-bg
// md-sp1, md-sp2, md-sp3, etc.
// md-sp1-bg, md-sp2-bg, md-sp3-bg, etc.
type ANSITags struct{}

func (r ANSITags) Document(contents string, depth int) string {
	return "<ansi fg=\"md\" bg=\"md-bg\">" + strings.TrimLeft(contents, "\n ") + "</ansi>"
}
func (r ANSITags) Paragraph(contents string, depth int) string {
	return "\n\n<ansi fg=\"md-p\" bg=\"md-p-bg\">" + contents + "</ansi>"
}
func (r ANSITags) HardBreak(contents string, depth int) string { return "\n" }
func (r ANSITags) Heading(contents string, depth int) string {
	if depth == 1 {
		contents = "<ansi fg=\"md-h1-prefix\" bg=\"md-h1-prefix-bg\">.:</ansi> " + contents
	}
	return "\n\n<ansi fg=\"md-h" + strconv.Itoa(depth) + "\" bg=\"md-h" + strconv.Itoa(depth) + "-bg\">" + contents + "</ansi>"
}
func (r ANSITags) List(contents string, depth int) string {
	if depth == 0 {
		return "\n\n" + contents
	}
	return strings.Repeat(` `, depth) + contents
}
func (r ANSITags) ListItem(contents string, depth int) string {
	return "\n" + strings.Repeat(` `, depth) + "<ansi fg=\"md-li\" bg=\"md-li-bg\">- " + contents + "</ansi>"
}
func (r ANSITags) Text(contents string, depth int) string {
	//return strings.TrimSpace(contents)
	return contents
}
func (r ANSITags) Strong(contents string, depth int) string {
	return "<ansi fg=\"md-bold\" bg=\"md-bold-bg\">" + contents + "</ansi>"
}
func (r ANSITags) Emphasis(contents string, depth int) string {
	return "<ansi fg=\"md-em\" bg=\"md-em-bg\">" + contents + "</ansi>"
}
func (r ANSITags) Special(contents string, depth int) string {
	return "<ansi fg=\"md-sp" + strconv.Itoa(depth) + "\" bg=\"md-sp" + strconv.Itoa(depth) + "-bg\">" + contents + "</ansi>"
}
