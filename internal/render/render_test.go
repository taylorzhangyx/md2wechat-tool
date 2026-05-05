package render_test

import (
	"strings"
	"testing"

	"github.com/geekjourneyx/md2wechat-skill/internal/render"
	"github.com/geekjourneyx/md2wechat-skill/internal/render/themes"
)

func TestRender_BasicMarkdown(t *testing.T) {
	md := "# Title\n\nA paragraph with **bold** and *em* and `code` and [link](https://example.com).\n\n> a quote\n\n---\n\n- item 1\n- item 2\n"
	got, err := render.Render(md, themes.MinimalGreen{}, render.Options{LinkStyle: render.LinkStyleNative})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	mustContain(t, got.HTML, "<section style=\"")
	mustContain(t, got.HTML, "<h1 style=\"")
	mustContain(t, got.HTML, "#19644d") // h1 dark green
	mustContain(t, got.HTML, "<strong style=\"")
	mustContain(t, got.HTML, "<em style=\"")
	mustContain(t, got.HTML, "<code style=\"")
	mustContain(t, got.HTML, `<a href="https://example.com"`)
	mustContain(t, got.HTML, "<blockquote style=\"")
	mustContain(t, got.HTML, "<hr style=\"")
	mustContain(t, got.HTML, "<ul style=\"")
	mustContain(t, got.HTML, "<li style=\"")
	mustNotContain(t, got.HTML, "<style")
}

func TestRender_Table(t *testing.T) {
	md := "| A | B |\n|---|---|\n| 1 | 2 |\n"
	got, err := render.Render(md, themes.MinimalGreen{}, render.Options{})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	mustContain(t, got.HTML, "<table style=\"")
	mustContain(t, got.HTML, "<thead style=\"")
	mustContain(t, got.HTML, "<th")
	mustContain(t, got.HTML, "<tbody")
	mustContain(t, got.HTML, "<td")
}

func TestRender_ImagesEmitPlaceholders(t *testing.T) {
	md := "First ![a](./a.png) then ![b](https://example.com/b.png)\n"
	got, err := render.Render(md, themes.MinimalGreen{}, render.Options{})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	mustContain(t, got.HTML, "<!-- IMG:0 -->")
	mustContain(t, got.HTML, "<!-- IMG:1 -->")
	if len(got.Images) != 2 {
		t.Fatalf("want 2 images, got %d", len(got.Images))
	}
	if got.Images[0].Src != "./a.png" {
		t.Fatalf("image[0].Src = %q", got.Images[0].Src)
	}
	if got.Images[1].Src != "https://example.com/b.png" {
		t.Fatalf("image[1].Src = %q", got.Images[1].Src)
	}
}

func TestRender_NoStyleTag(t *testing.T) {
	md := "# H\n\nA paragraph.\n"
	got, err := render.Render(md, themes.MinimalGreen{}, render.Options{})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if strings.Contains(got.HTML, "<style") {
		t.Fatalf("must not emit <style>: %s", got.HTML)
	}
}

// ---- link-style behaviour (WeChat strips <a href> for unverified accounts) ----

func TestRender_LinkStyleInline_DefaultRewritesLinks(t *testing.T) {
	// Default is "inline": [text](URL) → text（URL） before markdown parsing,
	// so goldmark never emits an <a> tag.
	md := "See [MMLU](https://arxiv.org/abs/2009.03300) for details.\n"
	got, err := render.Render(md, themes.MinimalGreen{}, render.Options{})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if strings.Contains(got.HTML, "<a href") {
		t.Fatalf("default (inline) link style must not emit <a href>; got:\n%s", got.HTML)
	}
	if !strings.Contains(got.HTML, "MMLU（https://arxiv.org/abs/2009.03300）") {
		t.Fatalf("expected flattened link text; got:\n%s", got.HTML)
	}
}

func TestRender_LinkStyleNative_KeepsAnchorTags(t *testing.T) {
	md := "See [MMLU](https://arxiv.org/abs/2009.03300) for details.\n"
	got, err := render.Render(md, themes.MinimalGreen{}, render.Options{LinkStyle: render.LinkStyleNative})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(got.HTML, `<a href="https://arxiv.org/abs/2009.03300"`) {
		t.Fatalf("native link style must emit <a href>; got:\n%s", got.HTML)
	}
}

// ---- duplicate h1 stripping (WeChat article page already shows title) ----

func TestRender_StripsLeadingTitleMatchingMetadata(t *testing.T) {
	md := "# My Title\n\nBody paragraph.\n"
	got, err := render.Render(md, themes.MinimalGreen{}, render.Options{Title: "My Title"})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if strings.Contains(got.HTML, "<h1") {
		t.Fatalf("leading h1 matching Title should be stripped; got:\n%s", got.HTML)
	}
	if !strings.Contains(got.HTML, "Body paragraph.") {
		t.Fatalf("body should be preserved; got:\n%s", got.HTML)
	}
}

func TestRender_KeepsLeadingTitleWhenNoMatch(t *testing.T) {
	md := "# Section Heading\n\nBody.\n"
	got, err := render.Render(md, themes.MinimalGreen{}, render.Options{Title: "Different Article Title"})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(got.HTML, "<h1") {
		t.Fatalf("leading h1 should survive when Title doesn't match; got:\n%s", got.HTML)
	}
}

// ---- Code block WeChat hardening: spaces→&nbsp;, \n→<br />, tabs→4×&nbsp; ----

func TestRender_CodeBlock_PreservesIndentation(t *testing.T) {
	md := "```bash\nif true; then\n    echo hi\nfi\n```\n"
	got, err := render.Render(md, themes.MinimalGreen{}, render.Options{})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	// Must emit &nbsp; for runs of spaces so WeChat can't collapse them.
	if !strings.Contains(got.HTML, "&amp;nbsp;&amp;nbsp;&amp;nbsp;&amp;nbsp;echo") && !strings.Contains(got.HTML, "&nbsp;&nbsp;&nbsp;&nbsp;echo") {
		t.Fatalf("indentation should be preserved via &nbsp; in code block; got:\n%s", got.HTML)
	}
	// Must emit <br /> for each newline.
	if !strings.Contains(got.HTML, "<br />") {
		t.Fatalf("newlines should be <br /> inside code block; got:\n%s", got.HTML)
	}
	// Must NOT contain raw runs of 4 plain spaces inside <pre><code>.
	codeStart := strings.Index(got.HTML, "<pre")
	codeEnd := strings.Index(got.HTML, "</code></pre>")
	if codeStart < 0 || codeEnd < 0 {
		t.Fatalf("missing <pre><code> block; got:\n%s", got.HTML)
	}
	codeSection := got.HTML[codeStart:codeEnd]
	if strings.Contains(codeSection, "    ") {
		t.Fatalf("raw 4-space run survived in code block (WeChat will collapse it):\n%s", codeSection)
	}
}

func TestRender_CodeBlock_ExpandsTabs(t *testing.T) {
	md := "```go\nfunc f() {\n\treturn 1\n}\n```\n"
	got, err := render.Render(md, themes.MinimalGreen{}, render.Options{})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	// Tab should become 4 × &nbsp;
	if !strings.Contains(got.HTML, "&nbsp;&nbsp;&nbsp;&nbsp;return") {
		t.Fatalf("tab should expand to 4 &nbsp; before 'return'; got:\n%s", got.HTML)
	}
}

func TestRender_CodeBlock_HandlesBlankLines(t *testing.T) {
	md := "```python\nline1\n\nline2\n```\n"
	got, err := render.Render(md, themes.MinimalGreen{}, render.Options{})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	// Blank line should become two <br /> in succession
	if !strings.Contains(got.HTML, "line1<br />\n<br />\nline2") {
		t.Fatalf("blank line inside code block should emit consecutive <br />; got:\n%s", got.HTML)
	}
}

func TestRender_InlineCode_NotAffectedByCodeBlockHardening(t *testing.T) {
	// Inline `code` uses renderCodeSpan, not the fenced-block path.
	md := "Some text with `   spaces   ` inside.\n"
	got, err := render.Render(md, themes.MinimalGreen{}, render.Options{})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if strings.Contains(got.HTML, "&nbsp;") {
		t.Fatalf("inline code should NOT get the &nbsp; treatment; got:\n%s", got.HTML)
	}
}

func mustContain(t *testing.T, s, sub string) {
	t.Helper()
	if !strings.Contains(s, sub) {
		t.Fatalf("missing %q in output:\n%s", sub, s)
	}
}

func mustNotContain(t *testing.T, s, sub string) {
	t.Helper()
	if strings.Contains(s, sub) {
		t.Fatalf("unexpected %q in output:\n%s", sub, s)
	}
}
