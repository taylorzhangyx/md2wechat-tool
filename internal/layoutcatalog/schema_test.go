package layoutcatalog

import "testing"

func TestValidServesContainsFourValues(t *testing.T) {
	want := []string{"attention", "readability", "memorability", "conversion"}
	for _, v := range want {
		if !ValidServes[v] {
			t.Errorf("expected %q to be a valid serve, missing", v)
		}
	}
	if len(ValidServes) != 4 {
		t.Errorf("ValidServes should contain exactly 4 values, got %d", len(ValidServes))
	}
}

func TestSchemaVersionConstant(t *testing.T) {
	if SchemaVersion != "1" {
		t.Errorf("SchemaVersion = %q, want %q", SchemaVersion, "1")
	}
}
