// Package themes holds the built-in Theme implementations for the local
// renderer. Each theme is an ordinary Go value so that the renderer can
// consult it without filesystem I/O.
package themes

import "github.com/geekjourneyx/md2wechat-skill/internal/render"

// MinimalGreen is the local re-implementation of the "minimal-green" theme
// from md2wechat.app. Inline styles are transcribed verbatim from the
// archived MHTML at themes/minimal-green/minimal-green-theme-detail.mhtml,
// so the output should match the reference rendering pixel-for-pixel in
// most clients (WeChat's editor sometimes rounds colors).
type MinimalGreen struct{}

const (
	mgPrimaryGreen = "#2BAE85"
	mgDarkGreen    = "#19644d"
	mgTintBG       = "#e4f0ed"
	mgBorderGreen  = "#b8d8ce"
	mgCodeBG       = "#f7f8fa"
	mgBodyText     = "#2c2c2c"
	mgMutedText    = "#555555"
	mgFontStack    = `-apple-system, BlinkMacSystemFont, 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei UI', 'Microsoft YaHei', Helvetica, Arial, sans-serif`
	mgMonoStack    = `'SF Mono', 'JetBrains Mono', Consolas, Monaco, Menlo, monospace`
)

func (MinimalGreen) Name() string { return "minimal-green" }

func (MinimalGreen) Style(e render.Element) string {
	switch e {
	case render.ElemContainer:
		return "max-width: 680px; margin: 0 auto; padding: 16px; background-color: #ffffff; color: " + mgBodyText +
			"; font-family: " + mgFontStack + "; font-size: 16px; line-height: 1.75; box-sizing: border-box;"

	case render.ElemH1:
		return "font-size: 28px; font-weight: 900; color: " + mgDarkGreen +
			"; margin-top: 36px; margin-bottom: 24px; line-height: 1.35;"
	case render.ElemH2:
		return "font-size: 24px; font-weight: 700; color: " + mgPrimaryGreen +
			"; margin-top: 36px; margin-bottom: 24px; line-height: 1.4; letter-spacing: -0.02em;"
	case render.ElemH3:
		return "font-size: 20px; font-weight: 700; color: " + mgPrimaryGreen +
			"; margin-top: 28px; margin-bottom: 16px; line-height: 1.45;"
	case render.ElemH4:
		return "font-size: 18px; font-weight: 900; color: " + mgDarkGreen +
			"; margin-top: 24px; margin-bottom: 12px;"
	case render.ElemH5, render.ElemH6:
		return "font-size: 16px; font-weight: 700; color: " + mgDarkGreen +
			"; margin-top: 20px; margin-bottom: 8px;"

	case render.ElemParagraph:
		return "margin: 5px 0 20px; line-height: 1.75; color: " + mgBodyText + "; word-break: break-word;"

	case render.ElemStrong:
		return "font-weight: 700; color: " + mgDarkGreen + ";"
	case render.ElemEmphasis:
		return "font-style: italic;"

	case render.ElemLink:
		return "color: " + mgDarkGreen + "; text-decoration: none; border-bottom: 1px solid " + mgBorderGreen + ";"

	case render.ElemBlockquote:
		return "margin: 16px 0; padding: 12px 16px; border-left: 3px solid " + mgPrimaryGreen +
			"; color: " + mgMutedText + "; font-style: italic; background: " + mgTintBG + "; border-radius: 0 6px 6px 0;"

	case render.ElemUL, render.ElemOL:
		return "margin: 8px 0 20px; padding-left: 1.2em; font-size: 16px; color: " + mgBodyText + "; line-height: 1.75;"
	case render.ElemLI:
		return "margin: 4px 0; line-height: 1.75;"

	case render.ElemCodeInline:
		return "display: inline-block; font-family: " + mgMonoStack +
			"; font-size: 13px; line-height: 1.45; padding: 2px 7px; color: " + mgPrimaryGreen +
			"; background: rgba(43, 174, 133, 0.1); border: 1px solid rgba(43, 174, 133, 0.18); border-radius: 5px;"
	case render.ElemCodeBlock:
		return "display: block; box-sizing: border-box; margin: 24px 0; padding: 1.15em 1.2em 1.2em; overflow-x: auto; " +
			"font-size: 14px; line-height: 1.65; font-family: " + mgMonoStack + "; color: " + mgBodyText +
			"; background: linear-gradient(180deg, rgba(43,174,133,0.16) 0px, rgba(43,174,133,0.16) 12px, " + mgCodeBG + " 12px, " + mgCodeBG + " 100%);" +
			" border: 1px solid rgba(43, 174, 133, 0.18); border-radius: 10px; box-shadow: 0 6px 18px rgba(43, 174, 133, 0.07), inset 0 1px 0 rgba(255,255,255,0.92);"

	case render.ElemHR:
		return "margin: 3rem 0; border: none; height: 1px; background-color: rgba(43, 174, 133, 0.2);"

	case render.ElemImage:
		return "width: 100%; max-width: 100%; border-radius: 12px; display: block; margin: 20px auto;"

	case render.ElemTable:
		return "width: 100%; margin: 24px 0; border-collapse: collapse; font-size: 15px; border: 1px solid " + mgBorderGreen + ";"
	case render.ElemTHead:
		return "background: " + mgTintBG + ";"
	case render.ElemTBody:
		return ""
	case render.ElemTR:
		return "border-bottom: 1px solid " + mgBorderGreen + ";"
	case render.ElemTH:
		return "padding: 10px 12px; text-align: left; font-weight: 700; color: " + mgDarkGreen +
			"; border-right: 1px solid " + mgBorderGreen + ";"
	case render.ElemTD:
		return "padding: 10px 12px; text-align: left; color: " + mgBodyText +
			"; border-right: 1px solid " + mgBorderGreen + "; vertical-align: top;"
	}
	return ""
}
