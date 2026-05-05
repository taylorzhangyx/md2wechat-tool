package enhance_test

import (
	"strings"
	"testing"

	"github.com/geekjourneyx/md2wechat-skill/internal/render"
	"github.com/geekjourneyx/md2wechat-skill/internal/render/enhance"
	"github.com/geekjourneyx/md2wechat-skill/internal/render/themes"
)

func renderWith(t *testing.T, md string, withEnhance bool) string {
	t.Helper()
	theme := themes.MinimalGreen{}
	got, err := render.Render(md, theme, render.Options{})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if withEnhance {
		return enhance.Default().Run(got.HTML, theme)
	}
	return got.HTML
}

// stripOuterContainer discards the outermost <section style="...">…</section>
// so tests checking for the presence of inner <section>/<div> tags don't
// accidentally match the ever-present outer container wrapper.
func innerBody(t *testing.T, html string) string {
	t.Helper()
	open := strings.Index(html, ">")
	if open < 0 {
		t.Fatalf("no opening tag in %q", html)
	}
	close := strings.LastIndex(html, "</section>")
	if close < 0 {
		t.Fatalf("no closing section in %q", html)
	}
	return html[open+1 : close]
}

// ---- TLDR rule ----

func TestTLDR_PromotesFollowingParagraph(t *testing.T) {
	md := "太长不看版：\n\nA summary paragraph.\n"
	out := renderWith(t, md, true)
	inner := innerBody(t, out)
	if !strings.Contains(out, "rule=tldr-callout") {
		t.Errorf("expected tldr-callout marker; got:\n%s", out)
	}
	if !strings.Contains(out, "<strong>太长不看版</strong>") {
		t.Errorf("expected <strong>太长不看版</strong> label; got:\n%s", out)
	}
	// WeChat strips <section> and <div> styles; avoid them entirely.
	if strings.Contains(inner, "<section") {
		t.Errorf("TLDR must NOT emit nested <section>; got:\n%s", inner)
	}
	if strings.Contains(inner, "<div") {
		t.Errorf("TLDR must NOT emit <div>; got:\n%s", inner)
	}
}

func TestTLDR_PromotesFollowingTable(t *testing.T) {
	md := "太长不看版：\n\n| A | B |\n|---|---|\n| 1 | 2 |\n"
	out := renderWith(t, md, true)
	inner := innerBody(t, out)
	if !strings.Contains(out, "rule=tldr-callout") {
		t.Errorf("expected marker; got:\n%s", out)
	}
	if !strings.Contains(out, "<table") {
		t.Errorf("table should survive; got:\n%s", out)
	}
	if strings.Contains(inner, "<section") {
		t.Errorf("TLDR must NOT wrap table in <section>; got:\n%s", inner)
	}
	if strings.Contains(inner, "<div") {
		t.Errorf("TLDR must NOT emit <div>; got:\n%s", inner)
	}
	if !strings.Contains(out, "<strong>太长不看版</strong>") {
		t.Errorf("expected <strong>太长不看版</strong>; got:\n%s", out)
	}
}

func TestTLDR_NoTriggerIsNoop(t *testing.T) {
	md := "Just a paragraph.\n\nAnother one.\n"
	withoutEnh := renderWith(t, md, false)
	withEnh := renderWith(t, md, true)
	if withoutEnh != withEnh {
		t.Errorf("rule should be noop without trigger; diff:\n--without--\n%s\n--with--\n%s", withoutEnh, withEnh)
	}
}

func TestTLDR_EnglishTrigger(t *testing.T) {
	md := "TL;DR:\n\nA summary.\n"
	out := renderWith(t, md, true)
	if !strings.Contains(out, "rule=tldr-callout") {
		t.Errorf("English trigger should match; got:\n%s", out)
	}
	if !strings.Contains(out, "<strong>TL;DR</strong>") {
		t.Errorf("expected <strong>TL;DR</strong>; got:\n%s", out)
	}
}

// ---- takeaway rule ----

func TestTakeaway_PromotesChapterEndQuote(t *testing.T) {
	md := "Leading para.\n\n> MMLU 高，说明模型基础知识面不错。\n\n---\n\nNext chapter.\n"
	out := renderWith(t, md, true)
	inner := innerBody(t, out)
	if !strings.Contains(out, "rule=takeaway-quote") {
		t.Errorf("expected takeaway marker; got:\n%s", out)
	}
	if strings.Contains(inner, "<div") {
		t.Errorf("takeaway must NOT emit <div> (WeChat unwraps it); got:\n%s", inner)
	}
	if !strings.Contains(out, "<blockquote") {
		t.Errorf("takeaway must use <blockquote>; got:\n%s", out)
	}
	if !strings.Contains(out, "<strong>MMLU 高") {
		t.Errorf("takeaway should wrap content in <strong>; got:\n%s", out)
	}
}

func TestTakeaway_PromotesEndOfDocument(t *testing.T) {
	md := "Leading.\n\n> The final takeaway.\n"
	out := renderWith(t, md, true)
	if !strings.Contains(out, "rule=takeaway-quote") {
		t.Errorf("end-of-doc quote should promote; got:\n%s", out)
	}
}

func TestTakeaway_SkipsMultiLineQuote(t *testing.T) {
	md := "Leading.\n\n> line one\n>\n> line two\n\n---\n"
	out := renderWith(t, md, true)
	if strings.Contains(out, "rule=takeaway-quote") {
		t.Errorf("multi-para quote should NOT promote; got:\n%s", out)
	}
}

func TestTakeaway_SkipsMidSectionQuote(t *testing.T) {
	md := "Leading.\n\n> A quote mid-section.\n\nMore prose after.\n"
	out := renderWith(t, md, true)
	if strings.Contains(out, "rule=takeaway-quote") {
		t.Errorf("mid-section quote should NOT promote; got:\n%s", out)
	}
}
