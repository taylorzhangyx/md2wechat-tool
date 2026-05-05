package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// loadDotenv looks for a .env file in common locations and sets any variable
// it finds into the process environment — but only if that variable is not
// already set. Real environment variables always win.
//
// Lookup order (first hit wins; subsequent files are ignored to keep
// behaviour predictable):
//
//  1. ./.env           — project root / CWD
//  2. ./.env.local     — local override
//  3. ~/.config/md2wechat/.env
//
// Unparsable lines are silently skipped. Comment lines start with '#'.
// Keys must match [A-Za-z_][A-Za-z0-9_]*. Values may be wrapped in
// single or double quotes; quotes are stripped. No variable expansion.
func loadDotenv() {
	candidates := []string{".env", ".env.local"}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		candidates = append(candidates, filepath.Join(home, ".config", "md2wechat", ".env"))
	}

	for _, path := range candidates {
		if path == "" {
			continue
		}
		if info, err := os.Stat(path); err != nil || info.IsDir() {
			continue
		}
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		applyDotenvReader(f)
		_ = f.Close()
		return
	}
}

func applyDotenvReader(r interface {
	Read(p []byte) (n int, err error)
}) {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Strip leading "export " if present.
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}
		eq := strings.IndexByte(line, '=')
		if eq <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:eq])
		val := strings.TrimSpace(line[eq+1:])
		if key == "" || !isValidEnvKey(key) {
			continue
		}
		// Strip surrounding matching quotes.
		if len(val) >= 2 {
			first, last := val[0], val[len(val)-1]
			if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
				val = val[1 : len(val)-1]
			}
		}
		if _, already := os.LookupEnv(key); already {
			continue
		}
		_ = os.Setenv(key, val)
	}
}

func isValidEnvKey(k string) bool {
	for i, r := range k {
		switch {
		case r >= 'A' && r <= 'Z':
		case r >= 'a' && r <= 'z':
		case r == '_':
		case i > 0 && r >= '0' && r <= '9':
		default:
			return false
		}
	}
	return true
}

// maskSecret returns a redacted form of a secret suitable for display.
// Short values are fully masked; longer values show a 4-character prefix.
func maskSecret(v string) string {
	if v == "" {
		return ""
	}
	if len(v) <= 6 {
		return strings.Repeat("*", len(v))
	}
	return v[:4] + strings.Repeat("*", len(v)-4)
}
