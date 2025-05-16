package markdown

type Formatter interface {
	Document(string, int) string
	Paragraph(string, int) string
	HardBreak(string, int) string
	Heading(string, int) string
	List(string, int) string
	ListItem(string, int) string
	Text(string, int) string
	Strong(string, int) string
	Emphasis(string, int) string
	Special(string, int) string
	Table(string, int) string
	TableHeader(string, int) string
	TableRow(string, int) string
	TableCell(string, int) string
}
