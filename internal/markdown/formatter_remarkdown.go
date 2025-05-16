package markdown

import "strings"

//
// Formats into a clean version of supported markdown
//

type ReMarkdown struct{}

func (r ReMarkdown) Document(contents string, depth int) string {
	return strings.TrimLeft(contents, "\n ")
}
func (r ReMarkdown) Paragraph(contents string, depth int) string { return "\n\n" + contents }
func (r ReMarkdown) HardBreak(contents string, depth int) string { return "\n" }
func (r ReMarkdown) Heading(contents string, depth int) string {
	return "\n\n" + strings.Repeat(`#`, depth) + " " + contents
}
func (r ReMarkdown) List(contents string, depth int) string {
	if depth == 0 {
		return "\n\n" + contents
	}
	return strings.Repeat(` `, depth) + contents
}
func (r ReMarkdown) ListItem(contents string, depth int) string {
	return "\n" + strings.Repeat(` `, depth) + "- " + contents
}
func (r ReMarkdown) Text(contents string, depth int) string {
	//return strings.TrimSpace(contents)
	return contents
}
func (r ReMarkdown) Strong(contents string, depth int) string   { return "**" + contents + "**" }
func (r ReMarkdown) Emphasis(contents string, depth int) string { return "*" + contents + "*" }
func (r ReMarkdown) Special(contents string, depth int) string {
	return strings.Repeat(`$`, depth) + contents + strings.Repeat(`$`, depth)
}
func (ReMarkdown) Table(contents string, _ int) string {
	return "\n" + contents
}
func (ReMarkdown) TableHeader(contents string, cellCount int) string {
	// we already want a leading pipe on each cell, so:
	return "\n" + contents + " |"
}

func (ReMarkdown) TableRow(contents string, cellCount int) string {
	return "\n" + contents + " |"
}

func (ReMarkdown) TableCell(contents string, _ int) string {
	return " | " + contents
}
