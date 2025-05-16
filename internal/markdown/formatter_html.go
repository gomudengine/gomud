package markdown

import (
	"strconv"
	"strings"
)

//
// Formats into HTML tags
//

type HTML struct{}

func (r HTML) Document(contents string, depth int) string {
	return strings.TrimLeft(contents, "\n ")
}
func (r HTML) Paragraph(contents string, depth int) string { return "\n<p>\n" + contents + "\n</p>" }
func (r HTML) HardBreak(contents string, depth int) string { return "\n<br />\n" }
func (r HTML) Heading(contents string, depth int) string {
	return "\n<h" + strconv.Itoa(depth) + ">" + contents + "</h" + strconv.Itoa(depth) + ">"
}
func (r HTML) List(contents string, depth int) string {
	return "\n" + strings.Repeat("\t", depth) + "<ul>" + contents + "\n" + strings.Repeat("\t", depth) + "</ul>"
}
func (r HTML) ListItem(contents string, depth int) string {
	return "\n" + strings.Repeat("\t", depth) + "<li>" + contents + "\n" + strings.Repeat("\t", depth) + "</li>"
}
func (r HTML) Text(contents string, depth int) string {
	return contents
}
func (r HTML) Strong(contents string, depth int) string   { return "<strong>" + contents + "</strong>" }
func (r HTML) Emphasis(contents string, depth int) string { return "<em>" + contents + "</em>" }
func (r HTML) Special(contents string, depth int) string {
	return "<special" + strconv.Itoa(depth) + ">" + contents + "</special" + strconv.Itoa(depth) + ">"
}
func (HTML) Table(contents string, _ int) string {
	return "\n<table>\n" + contents + "\n</table>\n"
}
func (HTML) TableHeader(contents string, _ int) string {
	return "<thead>\n<tr>" + contents + "</tr>\n</thead>\n"
}
func (HTML) TableRow(contents string, _ int) string {
	return "<tr>" + contents + "</tr>\n"
}
func (HTML) TableCell(contents string, _ int) string {
	return "<td>" + contents + "</td>"
}
