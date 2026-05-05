package render

import (
	"fmt"
	"regexp"
	"strings"
)

// LinkStyle controls how markdown links are rendered in the final HTML.
type LinkStyle string

const (
	// LinkStyleInline is the default: [text](URL) is rewritten to
	// text（URL） in the markdown before goldmark sees it. The rendered
	// HTML therefore contains no <a> tags — suitable for unverified
	// WeChat accounts whose editor silently strips all external hrefs.
	LinkStyleInline LinkStyle = "inline"

	// LinkStyleFootnote rewrites [text](URL) to "text[N]" in body order
	// and appends a "### 参考链接" section with URLs numbered 1..N to the
	// end of the markdown. Duplicate URLs share a number. Produces clean
	// body text with no inline URL noise, at the cost of an extra
	// reference section — skip this style if the article already has a
	// hand-curated Reference list of the same URLs.
	LinkStyleFootnote LinkStyle = "footnote"

	// LinkStyleNative keeps [text](URL) intact so goldmark emits a normal
	// <a href=…> anchor. Use this when the target WeChat account is
	// verified and allowed to ship external links, or when the output is
	// meant for a browser preview rather than WeChat.
	LinkStyleNative LinkStyle = "native"
)

// markdownLinkPattern matches a standard markdown link [text](URL).
//
// Constraints:
//   - Must NOT be preceded by '!' (that's an image embed).
//   - Must NOT be part of an Obsidian ![[…]] or [[…]] wiki-link.
//
// We guard against '!' with a leading negative lookbehind emulated by
// asserting either start-of-string or a non-'!' character. Go's regexp
// has no lookbehind, so we capture the preceding rune and restore it.
var markdownLinkPattern = regexp.MustCompile(`(^|[^!\[])\[([^\]\n]+)\]\(([^)\s]+)\)`)

// codeSpanPattern matches a fenced code block ``` … ``` OR an inline code
// span `…`. We need to exclude these from link rewriting because they may
// contain literal "[text](URL)" as documentation rather than a real link.
//
// The alternation handles, in order:
//  1. Fenced blocks: ```…```  (multi-line, greedy to next closing fence)
//  2. Inline spans: `…`      (single line, non-greedy)
//
// We run this before markdownLinkPattern; the regex engine tries
// alternation left-to-right so fenced blocks are matched first.
var codeSpanPattern = regexp.MustCompile("(?s)```[\\s\\S]*?```|`[^`\\n]+`")

// rewriteOutsideCode applies fn to every part of md that is NOT inside
// a code span or fenced code block. Code spans are passed through
// unchanged. This is the guard that keeps our link-rewriting passes from
// mangling illustrative "[text](URL)" examples inside code.
func rewriteOutsideCode(md string, fn func(string) string) string {
	if md == "" {
		return md
	}
	var out strings.Builder
	cursor := 0
	for _, loc := range codeSpanPattern.FindAllStringIndex(md, -1) {
		out.WriteString(fn(md[cursor:loc[0]]))
		out.WriteString(md[loc[0]:loc[1]])
		cursor = loc[1]
	}
	out.WriteString(fn(md[cursor:]))
	return out.String()
}

// FlattenMarkdownLinks rewrites every markdown link [text](URL) as
// "text（URL）" using full-width parentheses. Image embeds (![alt](url)),
// Obsidian wiki-links ([[…]], ![[…]]), and links inside inline/fenced
// code spans are preserved verbatim.
func FlattenMarkdownLinks(md string) string {
	if md == "" {
		return md
	}
	return rewriteOutsideCode(md, func(segment string) string {
		return markdownLinkPattern.ReplaceAllString(segment, "${1}${2}（${3}）")
	})
}

// FootnoteMarkdownLinks rewrites every markdown link [text](URL) to
// "text[N]" in document order and, at the end of the markdown, emits a
// numbered reference list mapping N → "text — URL". URLs that appear
// multiple times share a single number. Image embeds, Obsidian
// wiki-links, and text inside inline or fenced code spans are preserved
// verbatim. If no links are found, the input is returned unchanged.
//
// Reference-section handling: if the markdown already contains a
// Reference-like heading near the bottom (matches titles such as
// "Reference", "References", "参考", "参考资料", "参考链接",
// "参考文献", "延伸阅读"), the content below that heading is REPLACED
// with the auto-generated footnote list (the heading itself is kept
// intact). This avoids the confusing case where a hand-curated
// Reference list and an auto-generated 参考链接 list coexist with
// mismatched ordering. Articles without such a heading get a fresh
// "### 参考链接" section appended at the end.
//
// Design note: we do NOT use markdown reference-link syntax ([text][1])
// because goldmark would resolve it back into an <a href> if a matching
// [1]: URL definition is present — and we must NOT emit any <a href>
// for unverified-account output. "text[N]" is plain text.
func FootnoteMarkdownLinks(md string) string {
	if md == "" {
		return md
	}

	// First pass: allocate numbers for each unique URL in document order,
	// remembering the first link text we see for each URL. Skip matches
	// inside code spans.
	type entry struct {
		url  string
		text string
		n    int
	}
	urlToIdx := map[string]int{}
	var ordered []entry

	codeSpans := codeSpanPattern.FindAllStringIndex(md, -1)
	inCodeSpan := func(start int) bool {
		for _, loc := range codeSpans {
			if start >= loc[0] && start < loc[1] {
				return true
			}
		}
		return false
	}

	for _, m := range markdownLinkPattern.FindAllStringSubmatchIndex(md, -1) {
		if inCodeSpan(m[0]) {
			continue
		}
		url := md[m[6]:m[7]]
		if _, seen := urlToIdx[url]; seen {
			continue
		}
		text := md[m[4]:m[5]]
		n := len(ordered) + 1
		urlToIdx[url] = n
		ordered = append(ordered, entry{url: url, text: text, n: n})
	}
	if len(ordered) == 0 {
		return md
	}

	// Second pass: rewrite links outside code spans to "text[N]".
	rewritten := rewriteOutsideCode(md, func(segment string) string {
		return markdownLinkPattern.ReplaceAllStringFunc(segment, func(match string) string {
			sub := markdownLinkPattern.FindStringSubmatch(match)
			if len(sub) < 4 {
				return match
			}
			prefix, text, url := sub[1], sub[2], sub[3]
			n, ok := urlToIdx[url]
			if !ok {
				return match
			}
			return fmt.Sprintf("%s%s[%d]", prefix, text, n)
		})
	})

	// Build the footnote list body as a raw HTML <p> with <br/> line
	// breaks — deliberately NOT a markdown ordered list. goldmark would
	// render "1. text" as <ol><li>...</li></ol>, and WeChat's editor
	// prepends an empty <li> sibling to every real <li> in unverified
	// accounts (visible as a blank bullet between each entry). By
	// emitting <p>…<br/>…<br/>…</p> we avoid the <ul>/<ol> rendering
	// path entirely; blank-line separation around the block ensures
	// goldmark treats it as an HTML block and passes it through intact.
	//
	// We use bracket-style labels "[N]" instead of "N." so that, even if
	// some reader later pastes this into a markdown-aware tool, the
	// lines won't be re-parsed as list items.
	var body strings.Builder
	body.WriteString("<p>\n")
	for i, e := range ordered {
		fmt.Fprintf(&body, "[%d] %s — %s", e.n, e.text, e.url)
		if i < len(ordered)-1 {
			body.WriteString("<br/>\n")
		} else {
			body.WriteString("\n")
		}
	}
	body.WriteString("</p>\n")

	// If a Reference-like heading exists near the end, replace its
	// content section (everything until the next heading or EOF) with
	// the generated list. Otherwise append a fresh "### 参考链接" section.
	if head, replaced, ok := replaceReferenceSection(rewritten, body.String()); ok {
		_ = head
		return replaced
	}
	var appended strings.Builder
	appended.WriteString(rewritten)
	if !strings.HasSuffix(rewritten, "\n") {
		appended.WriteString("\n")
	}
	appended.WriteString("\n---\n\n### 参考链接\n\n")
	appended.WriteString(body.String())
	return appended.String()
}

// referenceHeadingPattern matches an ATX heading (one or more #) whose
// trimmed text matches one of the reference-list titles we want to
// reuse. Case-insensitive for the English variants.
var referenceHeadingPattern = regexp.MustCompile(
	`(?im)^(#{1,6})\s+(Reference|References|参考|参考资料|参考链接|参考文献|延伸阅读)\s*$`,
)

// replaceReferenceSection finds the LAST reference-like heading and
// replaces all content between that heading and the next same-or-higher
// heading (or EOF) with body. Returns the heading text that was kept,
// the resulting markdown, and true if a replacement happened.
func replaceReferenceSection(md, body string) (string, string, bool) {
	// Find all reference-heading matches and pick the last one — users
	// sometimes mention "参考" in prose above the real section, and the
	// actual section heading is always the last occurrence.
	idxs := referenceHeadingPattern.FindAllStringSubmatchIndex(md, -1)
	if len(idxs) == 0 {
		return "", md, false
	}
	last := idxs[len(idxs)-1]
	headStart, headEnd := last[0], last[1]
	headLevel := last[3] - last[2] // "#" run length

	heading := md[headStart:headEnd]

	// Find the end of this section: next line beginning with up to
	// headLevel hashes, OR EOF.
	tail := md[headEnd:]
	// Scan for "\n#{1,headLevel} " at a line start.
	endOffset := len(tail)
	lineStart := 0
	for lineStart < len(tail) {
		nl := strings.IndexByte(tail[lineStart:], '\n')
		var line string
		if nl < 0 {
			line = tail[lineStart:]
		} else {
			line = tail[lineStart : lineStart+nl]
		}
		trimmed := strings.TrimLeft(line, " \t")
		if strings.HasPrefix(trimmed, "#") {
			// Count leading '#'.
			hashes := 0
			for hashes < len(trimmed) && trimmed[hashes] == '#' {
				hashes++
			}
			if hashes > 0 && hashes <= headLevel &&
				hashes < len(trimmed) && (trimmed[hashes] == ' ' || trimmed[hashes] == '\t') {
				endOffset = lineStart
				break
			}
		}
		if nl < 0 {
			break
		}
		lineStart += nl + 1
	}

	before := md[:headStart]
	after := tail[endOffset:]

	var b strings.Builder
	b.WriteString(before)
	b.WriteString(heading)
	b.WriteString("\n\n")
	b.WriteString(body)
	if !strings.HasSuffix(after, "\n") && after != "" {
		b.WriteString("\n")
	}
	if after != "" {
		if !strings.HasPrefix(after, "\n") {
			b.WriteString("\n")
		}
		b.WriteString(after)
	}
	return heading, b.String(), true
}
