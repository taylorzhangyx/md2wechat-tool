package render_test

import (
	"strings"
	"testing"

	"github.com/geekjourneyx/md2wechat-skill/internal/render"
)

func TestFlattenMarkdownLinks_Basic(t *testing.T) {
	in := "See [MMLU paper](https://arxiv.org/abs/2009.03300) for details."
	got := render.FlattenMarkdownLinks(in)
	want := "See MMLU paper（https://arxiv.org/abs/2009.03300） for details."
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestFlattenMarkdownLinks_MultipleOnOneLine(t *testing.T) {
	in := "[one](http://a) and [two](http://b)"
	got := render.FlattenMarkdownLinks(in)
	want := "one（http://a） and two（http://b）"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestFlattenMarkdownLinks_PreservesImages(t *testing.T) {
	// ![alt](url) is an image, not a link — must NOT be rewritten.
	in := "![hero image](/path/to/img.png) and [link](http://x)"
	got := render.FlattenMarkdownLinks(in)
	if !strings.Contains(got, "![hero image](/path/to/img.png)") {
		t.Fatalf("image embed mutated: %s", got)
	}
	if !strings.Contains(got, "link（http://x）") {
		t.Fatalf("link not rewritten: %s", got)
	}
}

func TestFlattenMarkdownLinks_PreservesObsidianEmbeds(t *testing.T) {
	// ![[x.png]] must not be eaten by the link regex.
	in := "before ![[Pasted image.png]] after [text](url)"
	got := render.FlattenMarkdownLinks(in)
	if !strings.Contains(got, "![[Pasted image.png]]") {
		t.Fatalf("obsidian embed mutated: %s", got)
	}
}

func TestFlattenMarkdownLinks_NoLinks_Noop(t *testing.T) {
	in := "Plain markdown with no links, just text and **bold**."
	got := render.FlattenMarkdownLinks(in)
	if got != in {
		t.Fatalf("noop changed input: %q -> %q", in, got)
	}
}

// ---- FootnoteMarkdownLinks ----

func TestFootnoteMarkdownLinks_SingleLink(t *testing.T) {
	in := "See [MMLU](https://arxiv.org/abs/2009.03300) for details.\n"
	got := render.FootnoteMarkdownLinks(in)
	if !strings.Contains(got, "See MMLU[1] for details.") {
		t.Fatalf("expected text[1] rewrite; got:\n%s", got)
	}
	if !strings.Contains(got, "### 参考链接") {
		t.Fatalf("expected footnote section header; got:\n%s", got)
	}
	if !strings.Contains(got, "[1] MMLU — https://arxiv.org/abs/2009.03300") {
		t.Fatalf("expected footnote list entry with [N] bracket style; got:\n%s", got)
	}
}

func TestFootnoteMarkdownLinks_DedupesSameURL(t *testing.T) {
	in := "First [A](https://ex.com/x) and second [B](https://ex.com/x) and third [C](https://ex.com/y).\n"
	got := render.FootnoteMarkdownLinks(in)
	if !strings.Contains(got, "A[1]") || !strings.Contains(got, "B[1]") {
		t.Fatalf("duplicate URL should share footnote number; got:\n%s", got)
	}
	if !strings.Contains(got, "C[2]") {
		t.Fatalf("unique URL should get next number; got:\n%s", got)
	}
	if strings.Count(got, "[1] A — https://ex.com/x") != 1 {
		t.Fatalf("want one footnote entry for x; got:\n%s", got)
	}
	if strings.Count(got, "[2] C — https://ex.com/y") != 1 {
		t.Fatalf("want one footnote entry for y; got:\n%s", got)
	}
}

func TestFootnoteMarkdownLinks_NoLinks_Noop(t *testing.T) {
	in := "No links in this text at all."
	got := render.FootnoteMarkdownLinks(in)
	if got != in {
		t.Fatalf("no-op should not append footnote section; got:\n%s", got)
	}
}

func TestFootnoteMarkdownLinks_PreservesImages(t *testing.T) {
	in := "![cover](/a.png) and [text](https://x)"
	got := render.FootnoteMarkdownLinks(in)
	if !strings.Contains(got, "![cover](/a.png)") {
		t.Fatalf("image embed should survive footnote pass: %s", got)
	}
	if !strings.Contains(got, "text[1]") {
		t.Fatalf("link should be rewritten: %s", got)
	}
}

func TestFootnoteMarkdownLinks_InList(t *testing.T) {
	// Reference-list shape: - name: [link](url)
	// Without a Reference heading, the footnote is appended at the end.
	in := "- MMLU paper: [Measuring Massive Multitask Language Understanding](https://arxiv.org/abs/2009.03300)\n- GPQA: [A Graduate-Level](https://arxiv.org/abs/2311.12022)\n"
	got := render.FootnoteMarkdownLinks(in)
	if !strings.Contains(got, "- MMLU paper: Measuring Massive Multitask Language Understanding[1]") {
		t.Fatalf("list item 1 should rewrite inline link; got:\n%s", got)
	}
	if !strings.Contains(got, "- GPQA: A Graduate-Level[2]") {
		t.Fatalf("list item 2 should get next number; got:\n%s", got)
	}
}

// ---- Code-span skipping (prevents eating illustrative link syntax) ----

func TestFlatten_SkipsInlineCode(t *testing.T) {
	// The `[text](URL)` inside backticks is documentation, not a real
	// link — it must pass through unchanged.
	in := "Illustration: `[text](URL)`. Real link: [Real](https://x).\n"
	got := render.FlattenMarkdownLinks(in)
	if !strings.Contains(got, "`[text](URL)`") {
		t.Fatalf("inline code span should survive verbatim; got:\n%s", got)
	}
	if !strings.Contains(got, "Real（https://x）") {
		t.Fatalf("real link outside code should still be flattened; got:\n%s", got)
	}
}

func TestFootnote_SkipsInlineCode(t *testing.T) {
	in := "See `[text](URL)` illustration. Real: [Real](https://y).\n"
	got := render.FootnoteMarkdownLinks(in)
	if !strings.Contains(got, "`[text](URL)`") {
		t.Fatalf("inline code should survive footnote pass; got:\n%s", got)
	}
	if !strings.Contains(got, "Real[1]") {
		t.Fatalf("real link should rewrite to [1]; got:\n%s", got)
	}
	// URL inside code span must NOT appear in the footnote list.
	if strings.Contains(got, "1. URL") {
		t.Fatalf("code-span URL leaked into footnote list; got:\n%s", got)
	}
}

func TestFootnote_SkipsFencedCodeBlock(t *testing.T) {
	in := "Real: [R](https://x)\n\n```\n[fake](nope)\n```\n\nAlso real: [S](https://y)\n"
	got := render.FootnoteMarkdownLinks(in)
	if !strings.Contains(got, "R[1]") {
		t.Fatalf("pre-fence link should be [1]; got:\n%s", got)
	}
	if !strings.Contains(got, "S[2]") {
		t.Fatalf("post-fence link should be [2]; got:\n%s", got)
	}
	// The fake link inside ``` must not be rewritten.
	if !strings.Contains(got, "[fake](nope)") {
		t.Fatalf("fenced code content must survive verbatim; got:\n%s", got)
	}
	if strings.Contains(got, "nope") && strings.Count(got, "nope") != 1 {
		t.Fatalf("URL inside fence leaked into footnote list; got:\n%s", got)
	}
}

func TestFootnote_EntryIncludesLinkText(t *testing.T) {
	in := "See [MMLU paper](https://arxiv.org/abs/2009.03300) for details.\n"
	got := render.FootnoteMarkdownLinks(in)
	if !strings.Contains(got, "[1] MMLU paper — https://arxiv.org/abs/2009.03300") {
		t.Fatalf("footnote entry should be `[N] text — url`; got:\n%s", got)
	}
}

func TestFootnote_ReplacesExistingReferenceSection(t *testing.T) {
	in := strings.Join([]string{
		"Body with [link one](https://a.com) and [link two](https://b.com).",
		"",
		"## Reference",
		"",
		"- old item 1",
		"- old item 2",
		"",
	}, "\n")
	got := render.FootnoteMarkdownLinks(in)
	if !strings.Contains(got, "## Reference") {
		t.Fatalf("Reference heading should be preserved; got:\n%s", got)
	}
	if strings.Contains(got, "old item 1") || strings.Contains(got, "old item 2") {
		t.Fatalf("old Reference body should be replaced; got:\n%s", got)
	}
	if !strings.Contains(got, "[1] link one — https://a.com") {
		t.Fatalf("new footnote list missing; got:\n%s", got)
	}
	if !strings.Contains(got, "[2] link two — https://b.com") {
		t.Fatalf("new footnote list item 2 missing; got:\n%s", got)
	}
	if strings.Count(got, "参考链接") > 0 && !strings.Contains(got, "## Reference") {
		t.Fatalf("should not add extra 参考链接 heading when Reference exists; got:\n%s", got)
	}
}

func TestFootnote_MatchesChineseReferenceHeading(t *testing.T) {
	in := strings.Join([]string{
		"Body with [A](https://a.com).",
		"",
		"## 参考链接",
		"",
		"某段老内容",
		"",
	}, "\n")
	got := render.FootnoteMarkdownLinks(in)
	if !strings.Contains(got, "## 参考链接") {
		t.Fatalf("Chinese heading should be preserved; got:\n%s", got)
	}
	if strings.Contains(got, "某段老内容") {
		t.Fatalf("old Chinese content should be replaced; got:\n%s", got)
	}
	if !strings.Contains(got, "[1] A — https://a.com") {
		t.Fatalf("new list should replace old body; got:\n%s", got)
	}
}

func TestFootnote_AppendsWhenNoReferenceHeading(t *testing.T) {
	in := "Body with [A](https://a.com). No reference section.\n"
	got := render.FootnoteMarkdownLinks(in)
	if !strings.Contains(got, "### 参考链接") {
		t.Fatalf("should append fresh 参考链接 heading when none exists; got:\n%s", got)
	}
	if !strings.Contains(got, "[1] A — https://a.com") {
		t.Fatalf("should append numbered list entry; got:\n%s", got)
	}
}

func TestFootnote_UsesHTMLPNotOLLi(t *testing.T) {
	// WeChat's editor inserts an empty <li> before every <li> in
	// subscription-account drafts. By emitting the footnote list as a
	// single <p> with <br/> separators, we sidestep the <ol>/<ul>
	// rendering path entirely.
	in := "See [A](https://a.com) and [B](https://b.com).\n"
	got := render.FootnoteMarkdownLinks(in)
	if !strings.Contains(got, "<p>\n[1] A — https://a.com<br/>") {
		t.Fatalf("footnote should emit raw <p> with <br/>; got:\n%s", got)
	}
	// Must not look like a markdown ordered list ("1. " + "2. " at line start).
	lines := strings.Split(got, "\n")
	orderedListCount := 0
	for _, ln := range lines {
		trimmed := strings.TrimSpace(ln)
		if strings.HasPrefix(trimmed, "1. ") || strings.HasPrefix(trimmed, "2. ") {
			orderedListCount++
		}
	}
	if orderedListCount > 0 {
		t.Fatalf("footnote must not use markdown ordered list syntax; got:\n%s", got)
	}
}
