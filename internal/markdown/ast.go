package markdown

import (
	"fmt"
)

// NodeType identifies the kind of AST node.
type NodeType string

const (
	DocumentNode    NodeType = "Document"
	HeadingNode     NodeType = "Heading"
	ParagraphNode   NodeType = "Paragraph"
	HardBreakNode   NodeType = "HardBreak"
	ListNode        NodeType = "List"
	ListItemNode    NodeType = "ListItem"
	TextNode        NodeType = "Text"
	StrongNode      NodeType = "Strong"
	EmphasisNode    NodeType = "Emphasis"
	SpecialNode     NodeType = "Special"
	TableNode       NodeType = "Table"
	TableHeaderNode NodeType = "TableHeader"
	TableRowNode    NodeType = "TableRow"
	TableCellNode   NodeType = "TableCell"
)

var (
	activeFormatter Formatter = ReMarkdown{}
)

func SetFormatter(newFormatter Formatter) {
	activeFormatter = newFormatter
}

// Node is an element in the AST.
type Node interface {
	Type() NodeType
	Children() []Node
	String(int) string
}

// baseNode provides common fields.
type baseNode struct {
	nodeType     NodeType
	nodeChildren []Node
	level        int
	content      string
}

func (n *baseNode) Type() NodeType   { return n.nodeType }
func (n *baseNode) Children() []Node { return n.nodeChildren }
func (n *baseNode) String(depth int) string {
	ret := ``
	for _, c := range n.Children() {
		ret += c.String(depth + 1)
	}

	switch n.Type() {
	case DocumentNode:
		return activeFormatter.Document(ret, depth)
	case HeadingNode:
		return activeFormatter.Heading(ret, n.level)
	case ParagraphNode:
		return activeFormatter.Paragraph(ret, depth)
	case HardBreakNode:
		return activeFormatter.HardBreak(ret, depth)
	case ListNode:
		return activeFormatter.List(ret, depth)
	case ListItemNode:
		return activeFormatter.ListItem(ret, depth)
	case TextNode:
		return activeFormatter.Text(n.content+ret, depth)
	case StrongNode:
		return activeFormatter.Strong(ret, depth)
	case EmphasisNode:
		return activeFormatter.Emphasis(ret, depth)
	case SpecialNode:
		return activeFormatter.Special(ret, n.level)
	case TableNode:
		return activeFormatter.Table(ret, depth)
	case TableHeaderNode:
		return activeFormatter.TableHeader(ret, len(n.Children()))
	case TableRowNode:
		return activeFormatter.TableRow(ret, len(n.Children()))
	case TableCellNode:
		return activeFormatter.TableCell(ret, depth)
	default:
		return fmt.Sprintf(`[INVALID Node: type=%s depth=%d text=%s]`, n.Type(), depth, ret)
	}
}
