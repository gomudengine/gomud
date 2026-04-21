# Markdown Package Context

## Overview

The `internal/markdown` package is a custom, lightweight Markdown parser and formatter built specifically for GoMud. It parses a subset of Markdown into an AST and can render it to multiple output formats: ANSI-tagged terminal output, HTML, and clean Markdown. It is used to render help templates and other in-game text.

## Key Components

### AST (`ast.go`)

- **`NodeType`**: String enum for all node kinds (`DocumentNode`, `HeadingNode`, `ParagraphNode`, `HorizontalLineNode`, `HardBreakNode`, `ListNode`, `ListItemNode`, `TextNode`, `StrongNode`, `EmphasisNode`, `SpecialNode`)
- **`Node` interface**: `Type()`, `Children()`, `String(depth int)` — `String` dispatches to the active `Formatter`
- **`baseNode`**: Concrete implementation used for all node types; carries `level` (heading depth or special tilde count) and `content` (raw text for text/horizontal-line nodes)
- **`activeFormatter`**: Package-level `Formatter` variable, defaults to `ReMarkdown{}`. Change with `SetFormatter(f Formatter)`

### Parser (`parser.go`)

- **`Parser`**: Line-based parser; constructed with `NewParser(input string)` and produces a tree via `Parse() Node`
- **Supported constructs**:
  - Headings: `#`, `##`, `###`, …
  - Horizontal rules: lines starting with `---`, `===`, or `:::`
  - Unordered lists: `- item` with indent-based nesting
  - Paragraphs: runs of non-blank lines; hard line breaks via trailing two-space (`  `)
  - Inline bold: `**text**`
  - Inline emphasis: `*text*`
  - Inline special (tilde-delimited): `~text~`, `~~text~~`, etc. — level equals the tilde count

### Formatter Interface (`formatter.go`)

All formatters implement:

```go
type Formatter interface {
    Document(string, int) string
    Paragraph(string, int) string
    HorizontalLine(string, int) string
    HardBreak(string, int) string
    Heading(string, int) string
    List(string, int) string
    ListItem(string, int) string
    Text(string, int) string
    Strong(string, int) string
    Emphasis(string, int) string
    Special(string, int) string
}
```

### ANSITags Formatter (`formatter_ansitags.go`)

- Renders to GoMud ANSI tag markup (`<ansi fg="...">`)
- Uses named color aliases: `md`, `md-p`, `md-h1`–`md-hN`, `md-li`, `md-bold`, `md-em`, `md-sp1`–`md-spN`, `md-hr1`, `md-hr2`, plus `-bg` variants for each
- Three distinct horizontal-rule styles for `---`, `===`, and `:::`
- Heading level 1 gets a `.:` prefix rendered in `md-h1-prefix` color

### HTML Formatter (`formatter_html.go`)

- Renders to standard HTML tags (`<p>`, `<h1>`–`<hN>`, `<ul>`, `<li>`, `<strong>`, `<em>`, `<hr />`, `<br />`)
- `Special` nodes render as `<span data-special="N">…</span>`

### ReMarkdown Formatter (`formatter_remarkdown.go`)

- Round-trips Markdown back to clean Markdown text
- Default active formatter
- `Special` nodes render as `~…~` / `~~…~~` etc.

## Supported Markdown Subset

| Syntax | Node type |
|---|---|
| `# Heading` | `HeadingNode` (level = `#` count) |
| `---` / `===` / `:::` | `HorizontalLineNode` |
| `- item` | `ListNode` / `ListItemNode` |
| Trailing `  ` on a line | `HardBreakNode` |
| `**bold**` | `StrongNode` |
| `*em*` | `EmphasisNode` |
| `~special~` / `~~special~~` | `SpecialNode` (level = tilde count) |
| Everything else | `ParagraphNode` / `TextNode` |

Tables are **not** supported. Code blocks are **not** supported.

## Usage Patterns

```go
// Parse and render with the default (ReMarkdown) formatter
node := markdown.NewParser(input).Parse()
output := node.String(0)

// Render as ANSI tags for terminal output
markdown.SetFormatter(markdown.ANSITags{})
output := markdown.NewParser(input).Parse().String(0)

// Render as HTML
markdown.SetFormatter(markdown.HTML{})
output := markdown.NewParser(input).Parse().String(0)
```

`SetFormatter` is a package-level call — it changes the formatter globally. Callers that need to switch formatters should restore the previous one after use.

## Dependencies

- Standard library only (`regexp`, `strings`, `strconv`, `fmt`)
- No GoMud internal package dependencies

## Testing

Test coverage in `*_test.go` files covers:
- AST construction for all node types
- Inline parsing (bold, emphasis, special)
- List nesting
- Hard break detection
- Round-trip through all three formatters
