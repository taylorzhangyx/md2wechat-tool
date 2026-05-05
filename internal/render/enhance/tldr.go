package enhance

import (
	"regexp"
	"strings"

	"github.com/geekjourneyx/md2wechat-skill/internal/render"
)

// tldrCallout promotes a line matching the TL;DR trigger into a minimal
// "<p><strong>太长不看版</strong></p>" label. Unlike the original design,
// it does NOT wrap the following block in a <section>/<div> container —
// WeChat strips inline styles on those tags (and often unwraps <div>
// entirely), so any callout framing dies on publish.
//
// The new design is purposely cosmetic-light. It guarantees the label
// survives as bold text, and lets the following table/paragraph/list
// render as-is. The gain over plain markdown is:
//   - The trigger text becomes a bold marker readers recognise.
//   - The HTML enhancement comment makes the change auditable.
//
// Rule preconditions: the trigger line appears as a standalone paragraph
// whose only content is 太长不看版 / TL;DR / 一句话总结 with an optional
// trailing colon. If the paragraph has any other content the rule does
// not fire.
type tldrCallout struct{}

func (tldrCallout) Name() string { return "tldr-callout" }

// Trigger anchored at a <p ...>-only paragraph whose text exactly matches
// one of the TL;DR phrases. We accept half-width ":" and full-width "：".
var tldrTriggerRe = regexp.MustCompile(`(?m)<p[^>]*>\s*(太长不看版|TL;DR|tl;dr|一句话总结)\s*[：:]?\s*</p>\s*`)

func (tldrCallout) Apply(html string, _ render.Theme) string {
	return tldrTriggerRe.ReplaceAllStringFunc(html, func(match string) string {
		sub := tldrTriggerRe.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		label := normalizeTLDRLabel(sub[1])
		return "<!-- md2wechat: enhanced by rule=tldr-callout -->\n" +
			"<p><strong>" + label + "</strong></p>\n"
	})
}

func normalizeTLDRLabel(raw string) string {
	r := strings.ToLower(raw)
	switch {
	case strings.Contains(raw, "太长不看版"):
		return "太长不看版"
	case strings.Contains(raw, "一句话总结"):
		return "一句话总结"
	case strings.Contains(r, "tl;dr"):
		return "TL;DR"
	}
	return raw
}
