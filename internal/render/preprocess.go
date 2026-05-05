package render

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// obsidianEmbedPattern matches Obsidian-style embeds such as ![[Pasted image
// 20260504170054.png]] or ![[Pasted image.png|alt text]]. The filename may
// contain spaces but not ']' or '|' or newlines.
var obsidianEmbedPattern = regexp.MustCompile(`!\[\[([^\]|\n]+?)(?:\|([^\]\n]*))?\]\]`)

// ObsidianWarning carries a single unresolved Obsidian embed so the caller
// can surface it (log or JSON field) without breaking the render.
type ObsidianWarning struct {
	Filename string
	Reason   string
}

// ResolveObsidianEmbeds rewrites ![[filename]] to standard ![alt](abspath)
// markdown. Lookup order, relative to baseDir:
//
//  1. baseDir/<name>
//  2. baseDir/attachments/<name>
//  3. baseDir/../attachments/<name>
//  4. walk up two more directories looking for "attachments/<name>"
//
// On lookup failure the original embed is preserved and a warning is
// appended. The caller decides whether to log or surface it.
func ResolveObsidianEmbeds(markdown, baseDir string) (string, []ObsidianWarning) {
	if !strings.Contains(markdown, "![[") {
		return markdown, nil
	}

	var warnings []ObsidianWarning
	out := obsidianEmbedPattern.ReplaceAllStringFunc(markdown, func(match string) string {
		sub := obsidianEmbedPattern.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		name := strings.TrimSpace(sub[1])
		alt := ""
		if len(sub) >= 3 {
			alt = strings.TrimSpace(sub[2])
		}
		if alt == "" {
			alt = name
		}

		resolved, ok := locateEmbeddedFile(baseDir, name)
		if !ok {
			warnings = append(warnings, ObsidianWarning{
				Filename: name,
				Reason:   "file not found near " + baseDir,
			})
			return match
		}
		// Always wrap in angle brackets so spaces and other URL-unsafe
		// characters in absolute paths don't break goldmark's image
		// parser. CommonMark allows <> around destinations.
		return "![" + alt + "](<" + resolved + ">)"
	})

	return out, warnings
}

func locateEmbeddedFile(baseDir, name string) (string, bool) {
	if name == "" {
		return "", false
	}

	// Try siblings of each ancestor directory, up to the vault root or an
	// arbitrary depth. Obsidian users commonly store attachments at the
	// vault root, the markdown's own directory, or a sibling
	// "attachments/" folder. We also bail out once we reach filesystem
	// root or an .obsidian vault marker.
	const maxAscend = 6

	dir := baseDir
	for i := 0; i <= maxAscend; i++ {
		tries := []string{
			filepath.Join(dir, name),
			filepath.Join(dir, "attachments", name),
			filepath.Join(dir, "Attachments", name),
		}
		for _, p := range tries {
			if _, err := os.Stat(p); err == nil {
				if abs, err := filepath.Abs(p); err == nil {
					return abs, true
				}
				return p, true
			}
		}
		// Bail once we've hit an Obsidian vault root — the file would
		// have been under dir or dir/attachments.
		if _, err := os.Stat(filepath.Join(dir, ".obsidian")); err == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", false
}
