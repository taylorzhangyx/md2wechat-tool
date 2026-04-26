package layoutcatalog

import (
	"strings"
	"testing"
)

func TestRenderHeroFields(t *testing.T) {
	c := NewCatalog()
	if err := c.Load(); err != nil {
		t.Fatal(err)
	}
	out, err := c.Render("hero", map[string]any{
		"eyebrow":  "深度观察",
		"title":    "公众号排版的真问题不是好不好看",
		"subtitle": "是读者愿不愿意读完",
	})
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	if !strings.HasPrefix(out, ":::hero") || !strings.HasSuffix(strings.TrimRight(out, "\n"), ":::") {
		t.Errorf("output missing :::hero block fence:\n%s", out)
	}
	for _, want := range []string{"eyebrow: 深度观察", "title: 公众号排版的真问题不是好不好看", "subtitle: 是读者愿不愿意读完"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n%s", want, out)
		}
	}
}

func TestRenderMissingRequiredFieldFails(t *testing.T) {
	c := NewCatalog()
	if err := c.Load(); err != nil {
		t.Fatal(err)
	}
	_, err := c.Render("hero", map[string]any{"eyebrow": "x"})
	if err == nil {
		t.Fatal("expected error for missing title")
	}
	if !strings.Contains(err.Error(), "title") {
		t.Errorf("error should mention missing field name, got: %v", err)
	}
}

func TestRenderUnknownModuleFails(t *testing.T) {
	c := NewCatalog()
	if err := c.Load(); err != nil {
		t.Fatal(err)
	}
	_, err := c.Render("nonexistent", nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

// rowsSpecWithEnum returns an in-memory LayoutSpec that has rows mode,
// a required field with an enum constraint, and an optional field with an enum constraint.
func rowsSpecWithEnum() *LayoutSpec {
	return &LayoutSpec{
		SchemaVersion: SchemaVersion,
		Name:          "test-rows-enum",
		Category:      "body",
		Serves:        []string{"readability"},
		Fields: &FieldsSpec{
			Required: []FieldSpec{
				{Name: "align", Enum: []string{"left", "center", "right"}},
			},
			Optional: []FieldSpec{
				{Name: "style", Enum: []string{"plain", "bordered"}},
			},
		},
		Rows: &RowsSpec{Delimiter: "|", MinColumns: 1},
		Metadata: LayoutMetadata{
			Author:     "test",
			Provenance: "test",
		},
	}
}

func TestRenderRowsEnumRequiredInvalidFails(t *testing.T) {
	c := NewCatalog()
	c.modules["test-rows-enum"] = rowsSpecWithEnum()

	_, err := c.Render("test-rows-enum", map[string]any{
		"align": "invalid-align",
		"rows":  []any{[]any{"cell1"}},
	})
	if err == nil {
		t.Fatal("expected error for invalid enum value on required field in rows mode")
	}
	if !strings.Contains(err.Error(), "align") {
		t.Errorf("error should mention field name 'align', got: %v", err)
	}
}

func TestRenderRowsEnumOptionalInvalidFails(t *testing.T) {
	c := NewCatalog()
	c.modules["test-rows-enum"] = rowsSpecWithEnum()

	_, err := c.Render("test-rows-enum", map[string]any{
		"align": "left",
		"style": "fancy", // not in enum
		"rows":  []any{[]any{"cell1"}},
	})
	if err == nil {
		t.Fatal("expected error for invalid enum value on optional field in rows mode")
	}
	if !strings.Contains(err.Error(), "style") {
		t.Errorf("error should mention field name 'style', got: %v", err)
	}
}

func TestRenderRowsEnumValidSucceeds(t *testing.T) {
	c := NewCatalog()
	c.modules["test-rows-enum"] = rowsSpecWithEnum()

	out, err := c.Render("test-rows-enum", map[string]any{
		"align": "center",
		"style": "bordered",
		"rows":  []any{[]any{"a", "b"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "align: center") || !strings.Contains(out, "style: bordered") {
		t.Errorf("output missing expected fields:\n%s", out)
	}
}
