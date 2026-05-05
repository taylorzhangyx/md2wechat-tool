package render

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// Image holds one image reference discovered during rendering in document
// order. Index matches the "<!-- IMG:N -->" placeholder emitted in the HTML.
type Image struct {
	Index int
	Src   string
}

// Result bundles the rendered HTML, images in document order, and any
// non-fatal warnings (e.g. unresolved Obsidian embeds).
type Result struct {
	HTML     string
	Images   []Image
	Warnings []string
}

// Options tunes a single render call.
type Options struct {
	// BaseDir is the directory markdown was loaded from. Used to resolve
	// Obsidian-style ![[x.png]] embeds to real files.
	BaseDir string

	// Title, when non-empty, causes the renderer to strip the first line
	// of markdown if it is a level-1 heading whose text matches Title.
	// This avoids the article title appearing twice on a WeChat page
	// (once in the WeChat article header, once as the body's first h1).
	Title string

	// LinkStyle selects how markdown links are serialised. Empty defaults
	// to LinkStyleInline, which pre-flattens links to "text（URL）" before
	// goldmark runs — safe for unverified WeChat accounts. See LinkStyle
	// for the full set.
	LinkStyle LinkStyle
}

// Render converts markdown to inline-styled HTML using the given theme.
// The output wraps the rendered body in a themed container <section>.
func Render(markdown string, theme Theme, opts Options) (*Result, error) {
	if theme == nil {
		return nil, fmt.Errorf("render: theme must not be nil")
	}

	preprocessed, obsWarnings := ResolveObsidianEmbeds(markdown, opts.BaseDir)
	preprocessed = stripLeadingTitleHeading(preprocessed, opts.Title)

	switch opts.LinkStyle {
	case LinkStyleNative:
		// keep [text](URL) intact
	case LinkStyleFootnote:
		preprocessed = FootnoteMarkdownLinks(preprocessed)
	default: // "" or LinkStyleInline
		preprocessed = FlattenMarkdownLinks(preprocessed)
	}

	imgs := newImageState()
	inline := newInlineRenderer(theme, imgs)

	md := goldmark.New(
		goldmark.WithExtensions(extension.Table, extension.Strikethrough),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(
			renderer.WithNodeRenderers(
				util.Prioritized(inline, 100),
			),
		),
	)

	var buf bytes.Buffer
	if err := md.Convert([]byte(preprocessed), &buf); err != nil {
		return nil, fmt.Errorf("render: goldmark convert: %w", err)
	}

	body := buf.String()
	container := theme.Style(ElemContainer)
	var out bytes.Buffer
	if container != "" {
		_, _ = fmt.Fprintf(&out, `<section style="%s">`+"\n", container)
	} else {
		out.WriteString("<section>\n")
	}
	out.WriteString(body)
	out.WriteString("</section>\n")

	images := make([]Image, len(imgs.Sources()))
	for i, src := range imgs.Sources() {
		images[i] = Image{Index: i, Src: src}
	}

	warnings := make([]string, 0, len(obsWarnings))
	for _, w := range obsWarnings {
		warnings = append(warnings, fmt.Sprintf("obsidian embed %q: %s", w.Filename, w.Reason))
	}

	return &Result{
		HTML:     out.String(),
		Images:   images,
		Warnings: warnings,
	}, nil
}

// stripLeadingTitleHeading removes a leading "# Title" line from markdown
// body when Title matches the heading text exactly. This prevents WeChat
// from showing the article title twice (once in its own header, once as
// the body's opening h1).
func stripLeadingTitleHeading(body, title string) string {
	if title == "" || body == "" {
		return body
	}
	// Skip any leading blank lines.
	trimmed := strings.TrimLeft(body, "\n")
	newlineIdx := strings.IndexByte(trimmed, '\n')
	var firstLine, rest string
	if newlineIdx < 0 {
		firstLine = trimmed
		rest = ""
	} else {
		firstLine = trimmed[:newlineIdx]
		rest = trimmed[newlineIdx+1:]
	}
	firstLine = strings.TrimSpace(firstLine)
	if !strings.HasPrefix(firstLine, "# ") {
		return body
	}
	if strings.TrimSpace(firstLine[2:]) != strings.TrimSpace(title) {
		return body
	}
	return strings.TrimLeft(rest, "\n")
}
