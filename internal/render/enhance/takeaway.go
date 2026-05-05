package enhance

import (
	"regexp"
	"strings"

	"github.com/geekjourneyx/md2wechat-skill/internal/render"
)

// takeawayQuote promotes a single-line blockquote that sits at the end of
// a section (immediately followed by <hr>, <h2>, or the closing </section>)
// into a visually stronger "takeaway".
//
// Implementation note — why blockquote+strong, not a styled <div>:
// WeChat's subscription-account editor silently strips inline `style`
// attributes from <div> (and usually unwraps <div> into <p>), which used
// to destroy our upgrade entirely on publish. Emitting a real
// <blockquote> gives us the default left-border styling even with all
// inline styles stripped, and the inner <strong> stays bold. Both tags
// sit firmly inside WeChat's HTML whitelist.
//
// We match the canonical shape our own renderer emits:
//
//	<blockquote style="...">
//	<p style="...">CONTENT</p>
//	</blockquote>
//
// followed by optional whitespace, then one of:
//
//	<hr   // chapter divider
//	<h2   // next heading
//	</section>   // end of document
//
// Multi-paragraph blockquotes are skipped — they are a discussion, not a
// takeaway.
type takeawayQuote struct{}

func (takeawayQuote) Name() string { return "takeaway-quote" }

// single-line blockquote with exactly one inner <p>
var takeawaySingleBlockquoteRe = regexp.MustCompile(
	`(?s)<blockquote[^>]*>\s*<p[^>]*>(.*?)</p>\s*</blockquote>\s*`,
)

var takeawayTrailRe = regexp.MustCompile(`^(<hr |<h2 |</section>)`)

func (takeawayQuote) Apply(html string, _ render.Theme) string {
	matches := takeawaySingleBlockquoteRe.FindAllStringSubmatchIndex(html, -1)
	if len(matches) == 0 {
		return html
	}

	type promote struct {
		start, end int
		content    string
	}
	var promotions []promote

	for _, m := range matches {
		startFull, endFull := m[0], m[1]
		startG, endG := m[2], m[3]
		content := html[startG:endG]

		// Skip if the blockquote actually contains more than one <p>.
		inner := html[startFull:endFull]
		if strings.Count(inner, "<p ") > 1 || strings.Count(inner, "<p>") > 1 {
			continue
		}

		trail := html[endFull:]
		if !takeawayTrailRe.MatchString(trail) {
			continue
		}

		promotions = append(promotions, promote{
			start:   startFull,
			end:     endFull,
			content: content,
		})
	}

	if len(promotions) == 0 {
		return html
	}

	// Walk from end to start, splicing in replacements so byte offsets
	// stay valid.
	result := html
	for i := len(promotions) - 1; i >= 0; i-- {
		p := promotions[i]
		replacement := "<!-- md2wechat: enhanced by rule=takeaway-quote -->\n" +
			"<blockquote>\n<p><strong>" + p.content + "</strong></p>\n</blockquote>\n"
		result = result[:p.start] + replacement + result[p.end:]
	}

	return result
}
