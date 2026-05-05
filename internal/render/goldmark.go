package render

import (
	"fmt"
	"strings"

	"github.com/yuin/goldmark/ast"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// inlineRenderer implements renderer.NodeRenderer and emits HTML with inline
// styles drawn from a Theme. It is deliberately narrow: only the node kinds
// actually produced by the markdown we care about are handled. Unknown kinds
// fall through to goldmark's default renderer via higher priority numbers.
type inlineRenderer struct {
	theme      Theme
	imageState *imagePlaceholderState
}

func newInlineRenderer(theme Theme, imgs *imagePlaceholderState) *inlineRenderer {
	return &inlineRenderer{theme: theme, imageState: imgs}
}

func (r *inlineRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindHeading, r.renderHeading)
	reg.Register(ast.KindParagraph, r.renderParagraph)
	reg.Register(ast.KindTextBlock, r.renderTextBlock)
	reg.Register(ast.KindText, r.renderText)
	reg.Register(ast.KindString, r.renderString)
	reg.Register(ast.KindEmphasis, r.renderEmphasis)
	reg.Register(ast.KindLink, r.renderLink)
	reg.Register(ast.KindAutoLink, r.renderAutoLink)
	reg.Register(ast.KindImage, r.renderImage)
	reg.Register(ast.KindList, r.renderList)
	reg.Register(ast.KindListItem, r.renderListItem)
	reg.Register(ast.KindBlockquote, r.renderBlockquote)
	reg.Register(ast.KindCodeSpan, r.renderCodeSpan)
	reg.Register(ast.KindFencedCodeBlock, r.renderFencedCodeBlock)
	reg.Register(ast.KindCodeBlock, r.renderCodeBlock)
	reg.Register(ast.KindThematicBreak, r.renderThematicBreak)
	reg.Register(ast.KindHTMLBlock, r.renderHTMLBlock)
	reg.Register(ast.KindRawHTML, r.renderRawHTML)

	reg.Register(extast.KindTable, r.renderTable)
	reg.Register(extast.KindTableHeader, r.renderTableHeader)
	reg.Register(extast.KindTableRow, r.renderTableRow)
	reg.Register(extast.KindTableCell, r.renderTableCell)
}

func styleAttr(s string) string {
	if s == "" {
		return ""
	}
	return ` style="` + s + `"`
}

func (r *inlineRenderer) elementStyle(e Element) string {
	return styleAttr(r.theme.Style(e))
}

// ---------- block elements ----------

func (r *inlineRenderer) renderHeading(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	h := node.(*ast.Heading)
	tag := fmt.Sprintf("h%d", h.Level)
	var elem Element
	switch h.Level {
	case 1:
		elem = ElemH1
	case 2:
		elem = ElemH2
	case 3:
		elem = ElemH3
	case 4:
		elem = ElemH4
	case 5:
		elem = ElemH5
	default:
		elem = ElemH6
	}
	if entering {
		_, _ = fmt.Fprintf(w, "<%s%s>", tag, r.elementStyle(elem))
	} else {
		_, _ = fmt.Fprintf(w, "</%s>\n", tag)
	}
	return ast.WalkContinue, nil
}

func (r *inlineRenderer) renderParagraph(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	// Suppress the paragraph wrapper inside list items — the bullet/number
	// already occupies the line, and goldmark wraps "loose" items in <p>
	// which looks awkward with our inline-styled lists.
	if _, ok := node.Parent().(*ast.ListItem); ok {
		return ast.WalkContinue, nil
	}
	if entering {
		_, _ = fmt.Fprintf(w, "<p%s>", r.elementStyle(ElemParagraph))
	} else {
		_, _ = w.WriteString("</p>\n")
	}
	return ast.WalkContinue, nil
}

func (r *inlineRenderer) renderTextBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		return ast.WalkContinue, nil
	}
	if node.NextSibling() != nil && !node.HasBlankPreviousLines() {
		_, _ = w.WriteString("\n")
	}
	return ast.WalkContinue, nil
}

func (r *inlineRenderer) renderBlockquote(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = fmt.Fprintf(w, "<blockquote%s>", r.elementStyle(ElemBlockquote))
	} else {
		_, _ = w.WriteString("</blockquote>\n")
	}
	return ast.WalkContinue, nil
}

func (r *inlineRenderer) renderList(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	l := node.(*ast.List)
	tag := "ul"
	elem := ElemUL
	if l.IsOrdered() {
		tag = "ol"
		elem = ElemOL
	}
	if entering {
		_, _ = fmt.Fprintf(w, "<%s%s>\n", tag, r.elementStyle(elem))
	} else {
		_, _ = fmt.Fprintf(w, "</%s>\n", tag)
	}
	return ast.WalkContinue, nil
}

func (r *inlineRenderer) renderListItem(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = fmt.Fprintf(w, "<li%s>", r.elementStyle(ElemLI))
	} else {
		_, _ = w.WriteString("</li>\n")
	}
	return ast.WalkContinue, nil
}

func (r *inlineRenderer) renderFencedCodeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	_, _ = fmt.Fprintf(w, "<pre%s><code>", r.elementStyle(ElemCodeBlock))
	lines := node.Lines()
	for i := 0; i < lines.Len(); i++ {
		seg := lines.At(i)
		_, _ = w.Write(wechatSafeCode(util.EscapeHTML(seg.Value(source))))
	}
	_, _ = w.WriteString("</code></pre>\n")
	return ast.WalkSkipChildren, nil
}

// wechatSafeCode rewrites escaped code-block bytes so WeChat's sanitizer
// can't collapse whitespace or drop newlines.
//
// Why: inside <pre><code> WeChat strips `white-space: pre`-style inline CSS
// and then normalises the result as regular prose, which merges runs of
// spaces and silently drops raw '\n' characters. The result is a single
// unreadable line. The standard workaround — used by bm.md, MD2WeChat, and
// several other WeChat markdown typesetters — is to replace spaces with
// &nbsp; (non-breaking space; WeChat treats each as a real character) and
// newlines with <br /> (a whitelisted tag). Tabs expand to four &nbsp; for
// visual consistency with the gofmt convention our users mostly encounter.
func wechatSafeCode(escaped []byte) []byte {
	const nbsp = "&nbsp;"
	const br = "<br />"
	const tabWidth = 4

	out := make([]byte, 0, len(escaped)+len(escaped)/4)
	for i := 0; i < len(escaped); i++ {
		b := escaped[i]
		switch b {
		case ' ':
			out = append(out, nbsp...)
		case '\t':
			for t := 0; t < tabWidth; t++ {
				out = append(out, nbsp...)
			}
		case '\n':
			out = append(out, br...)
			out = append(out, '\n')
		default:
			out = append(out, b)
		}
	}
	return out
}

func (r *inlineRenderer) renderCodeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	return r.renderFencedCodeBlock(w, source, node, entering)
}

func (r *inlineRenderer) renderThematicBreak(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = fmt.Fprintf(w, "<hr%s />\n", r.elementStyle(ElemHR))
	}
	return ast.WalkContinue, nil
}

func (r *inlineRenderer) renderHTMLBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	b := node.(*ast.HTMLBlock)
	for i := 0; i < b.Lines().Len(); i++ {
		seg := b.Lines().At(i)
		_, _ = w.Write(seg.Value(source))
	}
	if b.HasClosure() {
		seg := b.ClosureLine
		_, _ = w.Write(seg.Value(source))
	}
	return ast.WalkSkipChildren, nil
}

// ---------- inline elements ----------

func (r *inlineRenderer) renderText(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	t := node.(*ast.Text)
	seg := t.Segment
	if t.IsRaw() {
		_, _ = w.Write(seg.Value(source))
	} else {
		_, _ = w.Write(util.EscapeHTML(seg.Value(source)))
		if t.HardLineBreak() {
			_, _ = w.WriteString("<br />\n")
		} else if t.SoftLineBreak() {
			_, _ = w.WriteString("\n")
		}
	}
	return ast.WalkContinue, nil
}

func (r *inlineRenderer) renderString(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	s := node.(*ast.String)
	if s.IsCode() {
		_, _ = w.Write(s.Value)
	} else {
		_, _ = w.Write(util.EscapeHTML(s.Value))
	}
	return ast.WalkContinue, nil
}

func (r *inlineRenderer) renderEmphasis(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	e := node.(*ast.Emphasis)
	tag := "em"
	elem := ElemEmphasis
	if e.Level == 2 {
		tag = "strong"
		elem = ElemStrong
	}
	if entering {
		_, _ = fmt.Fprintf(w, "<%s%s>", tag, r.elementStyle(elem))
	} else {
		_, _ = fmt.Fprintf(w, "</%s>", tag)
	}
	return ast.WalkContinue, nil
}

func (r *inlineRenderer) renderLink(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	l := node.(*ast.Link)
	if entering {
		_, _ = fmt.Fprintf(w, `<a href="%s"%s>`, escapeURL(string(l.Destination)), r.elementStyle(ElemLink))
	} else {
		_, _ = w.WriteString("</a>")
	}
	return ast.WalkContinue, nil
}

func (r *inlineRenderer) renderAutoLink(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	l := node.(*ast.AutoLink)
	label := l.Label(source)
	_, _ = fmt.Fprintf(w, `<a href="%s"%s>%s</a>`, escapeURL(string(label)), r.elementStyle(ElemLink), util.EscapeHTML(label))
	return ast.WalkContinue, nil
}

func (r *inlineRenderer) renderImage(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	img := node.(*ast.Image)
	src := string(img.Destination)
	placeholder := r.imageState.Add(src)
	_, _ = w.WriteString(placeholder)
	return ast.WalkSkipChildren, nil
}

func (r *inlineRenderer) renderCodeSpan(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = fmt.Fprintf(w, "<code%s>", r.elementStyle(ElemCodeInline))
		for c := node.FirstChild(); c != nil; c = c.NextSibling() {
			if t, ok := c.(*ast.Text); ok {
				_, _ = w.Write(util.EscapeHTML(t.Segment.Value(source)))
			}
		}
		_, _ = w.WriteString("</code>")
		return ast.WalkSkipChildren, nil
	}
	return ast.WalkContinue, nil
}

func (r *inlineRenderer) renderRawHTML(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	raw := node.(*ast.RawHTML)
	for i := 0; i < raw.Segments.Len(); i++ {
		seg := raw.Segments.At(i)
		_, _ = w.Write(seg.Value(source))
	}
	return ast.WalkSkipChildren, nil
}

// ---------- tables ----------

func (r *inlineRenderer) renderTable(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = fmt.Fprintf(w, "<table%s>\n", r.elementStyle(ElemTable))
	} else {
		_, _ = w.WriteString("</table>\n")
	}
	return ast.WalkContinue, nil
}

func (r *inlineRenderer) renderTableHeader(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = fmt.Fprintf(w, "<thead%s><tr%s>", r.elementStyle(ElemTHead), r.elementStyle(ElemTR))
	} else {
		_, _ = w.WriteString("</tr></thead>\n<tbody")
		_, _ = fmt.Fprintf(w, "%s>\n", r.elementStyle(ElemTBody))
	}
	return ast.WalkContinue, nil
}

func (r *inlineRenderer) renderTableRow(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = fmt.Fprintf(w, "<tr%s>", r.elementStyle(ElemTR))
	} else {
		_, _ = w.WriteString("</tr>\n")
		if node.NextSibling() == nil {
			_, _ = w.WriteString("</tbody>\n")
		}
	}
	return ast.WalkContinue, nil
}

func (r *inlineRenderer) renderTableCell(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	cell := node.(*extast.TableCell)
	tag := "td"
	elem := ElemTD
	if _, inHead := node.Parent().(*extast.TableHeader); inHead {
		tag = "th"
		elem = ElemTH
	}
	align := ""
	switch cell.Alignment {
	case extast.AlignLeft:
		align = " align=\"left\""
	case extast.AlignRight:
		align = " align=\"right\""
	case extast.AlignCenter:
		align = " align=\"center\""
	}
	if entering {
		_, _ = fmt.Fprintf(w, "<%s%s%s>", tag, align, r.elementStyle(elem))
	} else {
		_, _ = fmt.Fprintf(w, "</%s>", tag)
	}
	return ast.WalkContinue, nil
}

// escapeURL is a minimal URL escaper that preserves scheme and most path
// characters while neutralising quote characters.
func escapeURL(s string) string {
	s = strings.ReplaceAll(s, `"`, "%22")
	s = strings.ReplaceAll(s, "<", "%3C")
	s = strings.ReplaceAll(s, ">", "%3E")
	return s
}
