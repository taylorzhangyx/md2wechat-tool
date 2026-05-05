package converter

import (
	"fmt"

	"github.com/geekjourneyx/md2wechat-skill/internal/action"
	"github.com/geekjourneyx/md2wechat-skill/internal/render"
	"github.com/geekjourneyx/md2wechat-skill/internal/render/enhance"
	"github.com/geekjourneyx/md2wechat-skill/internal/render/themes"
	"go.uber.org/zap"
)

// localThemes maps theme names to their local Go implementations. Only
// themes with a local renderer should be listed here. Callers must treat a
// missing entry as "this theme has no local implementation".
var localThemes = map[string]render.Theme{
	"minimal-green": themes.MinimalGreen{},
}

// ErrThemeNotLocallyRendered indicates the requested theme has no local
// implementation. The caller may choose to fall back to API/AI mode.
var ErrThemeNotLocallyRendered = &ConvertError{
	Code:    "THEME_NOT_LOCAL",
	Message: "theme has no local implementation; try --mode api",
}

// localTheme returns the local Theme implementation for name, or nil if
// none is registered.
func localTheme(name string) render.Theme {
	if name == "" {
		return localThemes["minimal-green"]
	}
	return localThemes[name]
}

// IsLocalTheme reports whether the given theme name is locally renderable.
func IsLocalTheme(name string) bool { return localTheme(name) != nil }

// LocalThemeNames returns the list of locally renderable theme names in
// lexicographic order. Exposed for capabilities/themes list commands.
func LocalThemeNames() []string {
	names := make([]string, 0, len(localThemes))
	for n := range localThemes {
		names = append(names, n)
	}
	// simple sort without importing sort here to avoid pulling in the
	// package for two call sites — callers can sort if they need.
	for i := 1; i < len(names); i++ {
		for j := i; j > 0 && names[j-1] > names[j]; j-- {
			names[j-1], names[j] = names[j], names[j-1]
		}
	}
	return names
}

// convertViaLocal renders markdown entirely in-process via the render
// package, with optional layout enhancements applied to the resulting HTML.
func (c *converter) convertViaLocal(req *ConvertRequest) *ConvertResult {
	result := &ConvertResult{
		Mode:      ModeLocal,
		Theme:     req.Theme,
		Status:    action.StatusFailed,
		Action:    action.ActionConvert,
		Retryable: false,
		Success:   false,
	}

	theme := localTheme(req.Theme)
	if theme == nil {
		result.Error = fmt.Sprintf("theme %q has no local implementation; available: %v", req.Theme, LocalThemeNames())
		c.log.Warn("local convert: unknown theme", zap.String("theme", req.Theme))
		return result
	}

	rendered, err := render.Render(req.Markdown, theme, render.Options{
		BaseDir:   req.BaseDir,
		Title:     req.Metadata.Title,
		LinkStyle: render.LinkStyle(req.LinkStyle),
	})
	if err != nil {
		result.Error = fmt.Sprintf("local render failed: %s", err.Error())
		c.log.Error("local render failed", zap.Error(err))
		return result
	}

	html := rendered.HTML
	if !req.NoEnhance {
		html = enhance.Default().Run(html, theme)
	}

	for _, w := range rendered.Warnings {
		c.log.Warn("local render warning", zap.String("warning", w))
	}

	// Map render.Image → converter.ImageRef, preserving placeholder ordering.
	images := make([]ImageRef, len(rendered.Images))
	for i, img := range rendered.Images {
		images[i] = ImageRef{
			Index:       img.Index,
			Original:    img.Src,
			Placeholder: fmt.Sprintf("<!-- IMG:%d -->", img.Index),
			Type:        classifyImageSrc(img.Src),
		}
	}

	result.HTML = html
	result.Images = images
	result.Status = action.StatusCompleted
	result.Success = true

	c.log.Info("local conversion succeeded",
		zap.String("theme", req.Theme),
		zap.Int("image_count", len(images)),
		zap.Bool("enhance", !req.NoEnhance))
	return result
}

func classifyImageSrc(src string) ImageType {
	if len(src) >= 7 && (src[:7] == "http://" || (len(src) >= 8 && src[:8] == "https://")) {
		return ImageTypeOnline
	}
	if len(src) >= 11 && src[:11] == "__generate:" {
		return ImageTypeAI
	}
	return ImageTypeLocal
}
