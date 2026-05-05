package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadDotenv_PopulatesMissingEnv(t *testing.T) {
	// Isolate: save & restore CWD, create temp dir with a .env.
	tmp := t.TempDir()
	_ = os.Unsetenv("TEST_DOTENV_KEY_1")
	_ = os.Unsetenv("TEST_DOTENV_KEY_2")

	path := filepath.Join(tmp, ".env")
	body := strings.Join([]string{
		"# a comment",
		"TEST_DOTENV_KEY_1=value1",
		`TEST_DOTENV_KEY_2="value with spaces"`,
		"export TEST_DOTENV_KEY_EXPORTED='abc'",
		"malformed line without equals",
	}, "\n")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	old, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(old) })
	_ = os.Chdir(tmp)

	loadDotenv()

	if got := os.Getenv("TEST_DOTENV_KEY_1"); got != "value1" {
		t.Errorf("TEST_DOTENV_KEY_1 = %q", got)
	}
	if got := os.Getenv("TEST_DOTENV_KEY_2"); got != "value with spaces" {
		t.Errorf("TEST_DOTENV_KEY_2 = %q", got)
	}
	if got := os.Getenv("TEST_DOTENV_KEY_EXPORTED"); got != "abc" {
		t.Errorf("TEST_DOTENV_KEY_EXPORTED = %q", got)
	}

	// Cleanup test env
	for _, k := range []string{"TEST_DOTENV_KEY_1", "TEST_DOTENV_KEY_2", "TEST_DOTENV_KEY_EXPORTED"} {
		_ = os.Unsetenv(k)
	}
}

func TestLoadDotenv_DoesNotOverrideExisting(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("TEST_DOTENV_EXISTING", "from-env")

	path := filepath.Join(tmp, ".env")
	if err := os.WriteFile(path, []byte("TEST_DOTENV_EXISTING=from-file\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	old, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(old) })
	_ = os.Chdir(tmp)

	loadDotenv()

	if got := os.Getenv("TEST_DOTENV_EXISTING"); got != "from-env" {
		t.Errorf("real env should win; got %q", got)
	}
}

func TestMaskSecret(t *testing.T) {
	cases := map[string]string{
		"":          "",
		"a":         "*",
		"abcdef":    "******",
		"abcdefg":   "abcd***",
		"wx1234567": "wx12*****",
	}
	for in, want := range cases {
		if got := maskSecret(in); got != want {
			t.Errorf("maskSecret(%q) = %q, want %q", in, got, want)
		}
	}
}
