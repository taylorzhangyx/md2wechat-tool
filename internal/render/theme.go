// Package render provides a local Markdown-to-HTML renderer that produces
// WeChat-compatible HTML with all CSS inlined onto every element.
//
// The WeChat public-account editor strips <style> tags and most external CSS
// rules, so every element must carry its own style attribute. This package
// wraps goldmark with a custom NodeRenderer that consults a Theme for the
// inline style string to emit on each element.
package render

// Element identifies a styled node kind. Using plain string keys keeps the
// theme map simple and YAML-friendly for future per-theme asset files.
type Element string

const (
	ElemContainer  Element = "container"
	ElemH1         Element = "h1"
	ElemH2         Element = "h2"
	ElemH3         Element = "h3"
	ElemH4         Element = "h4"
	ElemH5         Element = "h5"
	ElemH6         Element = "h6"
	ElemParagraph  Element = "p"
	ElemStrong     Element = "strong"
	ElemEmphasis   Element = "em"
	ElemLink       Element = "a"
	ElemBlockquote Element = "blockquote"
	ElemUL         Element = "ul"
	ElemOL         Element = "ol"
	ElemLI         Element = "li"
	ElemCodeInline Element = "code_inline"
	ElemCodeBlock  Element = "pre"
	ElemHR         Element = "hr"
	ElemImage      Element = "img"
	ElemTable      Element = "table"
	ElemTHead      Element = "thead"
	ElemTBody      Element = "tbody"
	ElemTR         Element = "tr"
	ElemTH         Element = "th"
	ElemTD         Element = "td"
)

// Theme returns inline CSS strings for each styled element.
//
// Implementations must return deterministic strings (no randomness, no time
// of day). Returning an empty string is allowed and means "no style" for
// that element.
type Theme interface {
	// Name returns the theme's canonical identifier.
	Name() string

	// Style returns the inline CSS for the given element. An unknown element
	// should return an empty string rather than an error.
	Style(e Element) string
}
