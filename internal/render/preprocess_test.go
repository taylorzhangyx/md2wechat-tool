package render_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/geekjourneyx/md2wechat-skill/internal/render"
)

func TestResolveObsidianEmbeds_Found(t *testing.T) {
	dir := t.TempDir()
	fname := "Pasted image 20260504170054.png"
	if err := os.WriteFile(filepath.Join(dir, fname), []byte("stub"), 0o644); err != nil {
		t.Fatal(err)
	}
	md := "before ![[" + fname + "]] after"
	got, warns := render.ResolveObsidianEmbeds(md, dir)
	if len(warns) != 0 {
		t.Fatalf("unexpected warnings: %+v", warns)
	}
	if !strings.Contains(got, "![Pasted image 20260504170054.png](") {
		t.Fatalf("did not rewrite embed: %s", got)
	}
	if !strings.Contains(got, fname) {
		t.Fatalf("absolute path should include filename: %s", got)
	}
}

func TestResolveObsidianEmbeds_FoundInAttachmentsDir(t *testing.T) {
	root := t.TempDir()
	att := filepath.Join(root, "attachments")
	if err := os.Mkdir(att, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(att, "x.png"), []byte("."), 0o644); err != nil {
		t.Fatal(err)
	}
	got, warns := render.ResolveObsidianEmbeds("![[x.png]]", root)
	if len(warns) != 0 {
		t.Fatalf("unexpected warnings: %+v", warns)
	}
	if !strings.Contains(got, "attachments/x.png") {
		t.Fatalf("want path under attachments/, got: %s", got)
	}
}

func TestResolveObsidianEmbeds_NotFound_Preserves(t *testing.T) {
	got, warns := render.ResolveObsidianEmbeds("![[missing.png]]", t.TempDir())
	if len(warns) != 1 {
		t.Fatalf("want 1 warning, got %d: %+v", len(warns), warns)
	}
	if warns[0].Filename != "missing.png" {
		t.Fatalf("warning filename: %q", warns[0].Filename)
	}
	if !strings.Contains(got, "![[missing.png]]") {
		t.Fatalf("should preserve original on miss: %s", got)
	}
}

func TestResolveObsidianEmbeds_NoEmbeds_Noop(t *testing.T) {
	in := "No embeds here, just ![std](x.png) links."
	got, warns := render.ResolveObsidianEmbeds(in, t.TempDir())
	if got != in {
		t.Fatalf("input mutated: %s", got)
	}
	if len(warns) != 0 {
		t.Fatalf("unexpected warnings: %+v", warns)
	}
}

func TestResolveObsidianEmbeds_WithAltText(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "pic.png"), []byte("."), 0o644); err != nil {
		t.Fatal(err)
	}
	got, _ := render.ResolveObsidianEmbeds("![[pic.png|my caption]]", dir)
	if !strings.Contains(got, "![my caption](") {
		t.Fatalf("want alt text preserved: %s", got)
	}
}
