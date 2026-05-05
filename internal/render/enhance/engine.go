// Package enhance applies rule-driven layout upgrades to already-rendered
// HTML. Rules operate deterministically on the HTML string produced by the
// inline-styled goldmark renderer — since we control that output, we know
// exactly what shape to match.
//
// Each rule is narrowly scoped and comes with a body of tests. Rules that
// do not match pass the HTML through unchanged.
package enhance

import (
	"github.com/geekjourneyx/md2wechat-skill/internal/render"
)

// Rule is a single enhancement pass. Apply must be a pure string transform
// — no I/O, no randomness — so rule ordering is the only source of
// indeterminism.
type Rule interface {
	Name() string
	Apply(html string, theme render.Theme) string
}

// Pipeline runs an ordered set of rules.
type Pipeline struct {
	rules []Rule
}

// Default returns the enhancement pipeline shipped with the tool today:
//
//  1. tldrCallout      — promotes "太长不看版/TL;DR/一句话总结" + next block to callout
//  2. takeawayQuote    — promotes chapter-end single-line blockquote to takeaway
//
// The order matters only if rules can overlap; these two don't.
func Default() *Pipeline {
	return &Pipeline{rules: []Rule{
		tldrCallout{},
		takeawayQuote{},
	}}
}

// Run applies all rules in order and returns the transformed HTML.
func (p *Pipeline) Run(html string, theme render.Theme) string {
	for _, r := range p.rules {
		html = r.Apply(html, theme)
	}
	return html
}
