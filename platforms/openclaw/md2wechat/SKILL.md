---
name: md2wechat
description: Convert Markdown to WeChat Official Account HTML, inspect supported providers/themes/prompts, generate article images, create drafts, write with creator styles, and remove AI writing traces.
homepage: https://github.com/geekjourneyx/md2wechat-skill
metadata: {"clawdbot":{"emoji":"📝","requires":{"bins":["md2wechat"],"env":["WECHAT_APPID","WECHAT_SECRET"]},"install":[{"id":"brew","kind":"brew","formula":"geekjourneyx/tap/md2wechat","bins":["md2wechat"],"label":"Install md2wechat (brew)"},{"id":"go","kind":"go","module":"github.com/geekjourneyx/md2wechat-skill/cmd/md2wechat@latest","bins":["md2wechat"],"label":"Install md2wechat (go)"}]}}
---

# md2wechat

Use `md2wechat` when the user wants to:

- convert Markdown into WeChat Official Account HTML
- inspect resolved article metadata, readiness, and publish risks before conversion
- generate a local preview artifact or upload drafts
- inspect live capabilities, providers, themes, and prompts
- generate covers, infographics, or other article images
- create image posts
- write in creator styles or remove AI writing traces

## Intent Routing

Choose the command family before doing any publish action:

- Use `convert` / `inspect` / `preview` when the user wants a standard WeChat article draft (`news`), HTML conversion, article metadata, article preview, or a draft that needs `--cover`.
- Use `create_image_post` when the user says `小绿书`, `图文笔记`, `图片消息`, `newspic`, `多图帖子`, or asks to publish an image-first post rather than an article HTML draft.
- Do not route `小绿书` / `图文笔记` requests to `convert --draft` just because the user also has a Markdown article. A Markdown file can still be the image source for `create_image_post -m article.md`.
- Treat `convert --draft` and `create_image_post` as different publish targets, not interchangeable command variants.

## Defaults And Config

- Use this skill only when `md2wechat` is already available on `PATH`.
- Draft upload and publish-related actions require `WECHAT_APPID` and `WECHAT_SECRET`. The CLI auto-loads `./.env`, `./.env.local`, and `~/.config/md2wechat/.env` at startup; real shell env vars always win. Draft/upload calls also need the machine's public IP to be added to the WeChat MP 开发 → 基本配置 → IP 白名单; otherwise `errcode=40164`.
- Image generation may require additional image-service configuration in `~/.config/md2wechat/config.yaml`.
- `convert` defaults to `local` mode (offline goldmark renderer, theme `minimal-green`). Output goes to stdout unless `--output <file>` is passed. Local mode applies two layout enhancements by default (`太长不看版 / TL;DR` → bold callout, chapter-end single-line `>` quote → `<blockquote><strong>` takeaway); pass `--no-enhance` to disable. Local mode also auto-resolves Obsidian `![[file.png]]` embeds (walks up to 6 levels or a `.obsidian/` marker).
- Link handling in local mode is controlled by `--link-style`:
  - `inline` (default): rewrites `[text](URL)` → `text（URL）`.
  - `footnote`: rewrites to `text[N]` + reference list at the end as `<p>+<br/>`; reuses `## Reference` / `## 参考` / `## 参考链接` / `## 参考资料` / `## 参考文献` / `## 延伸阅读` heading if present.
  - `native`: keeps `<a href>` (verified / 微信支付-enabled accounts only).
- For 40+ server-rendered themes use `--mode api` (needs `MD2WECHAT_API_KEY`); for LLM-generated HTML use `--mode ai`. Non-local modes ignore `--link-style` and enhancement flags.
- Check configuration in this order:
  1. `~/.config/md2wechat/config.yaml`
  2. environment variables such as `MD2WECHAT_BASE_URL`
  3. project-local `md2wechat.yaml`, `md2wechat.yml`, or `md2wechat.json`
- If the user asks to switch API domain, update `api.md2wechat_base_url` or `MD2WECHAT_BASE_URL`.
- Treat live CLI discovery output as the source of truth. Do not guess provider names, theme names, or prompt names from repository files alone.

## Discovery First

Run these before selecting a provider, theme, or prompt:

- `md2wechat version --json`
- `md2wechat capabilities --json`
- `md2wechat providers list --json`
- `md2wechat themes list --json`
- `md2wechat prompts list --json`
- `md2wechat prompts list --kind image --json`
- `md2wechat prompts list --kind image --archetype cover --json`

Inspect a specific resource before using it:

- `md2wechat providers show openrouter --json`
- `md2wechat providers show volcengine --json`
- `md2wechat themes show autumn-warm --json`
- `md2wechat prompts show cover-default --kind image --json`
- `md2wechat prompts show cover-hero --kind image --archetype cover --tag hero --json`
- `md2wechat prompts show infographic-victorian-engraving-banner --kind image --archetype infographic --tag victorian --json`
- `md2wechat prompts render cover-default --kind image --var article_title='Example' --json`

When choosing image presets, prefer the prompt metadata returned by `prompts show --json`, especially `primary_use_case`, `compatible_use_cases`, `recommended_aspect_ratios`, and `default_aspect_ratio`.
When choosing an image model, prefer `providers show <name> --json` and read `supported_models` before hard-coding `--model`.

## Core Commands

Configuration:

- `md2wechat config init`
- `md2wechat config show --format json`
- `md2wechat config validate`

Conversion:

- `md2wechat inspect article.md` — resolve metadata, readiness, and publish risks
- `md2wechat preview article.md` — standalone HTML preview file
- `md2wechat convert article.md` — print rendered HTML to stdout (local mode, minimal-green)
- `md2wechat convert article.md -o output.html` — write HTML to file
- `md2wechat convert article.md --no-enhance` — disable TL;DR / takeaway upgrades
- `md2wechat convert article.md --link-style=footnote` — rewrite links to `text[N]` + reference list
- `md2wechat convert article.md --link-style=native` — keep `<a href>` (verified accounts only)
- `md2wechat convert article.md --upload` — upload images and replace src URLs
- `md2wechat convert article.md --upload --draft --cover cover.jpg` — full publish-to-draft
- `md2wechat convert article.md --upload --draft --cover-media-id <id>` — reuse existing cover asset
- `md2wechat convert article.md --mode api --theme minimal-blue --preview` — server-rendered, 40+ themes
- `md2wechat convert article.md --mode ai --theme autumn-warm --preview` — AI-generated HTML path
- `md2wechat convert article.md --title "新标题" --author "作者名" --digest "摘要"` — override metadata

Image handling:

- `md2wechat upload_image photo.jpg`
- `md2wechat download_and_upload https://example.com/image.jpg`
- `md2wechat generate_image "A cute cat sitting on a windowsill"`
- `md2wechat generate_image --preset cover-hero --article article.md --size 2560x1440`
- `md2wechat generate_cover --article article.md`
- `md2wechat generate_infographic --article article.md --preset infographic-comparison`
- `md2wechat generate_infographic --article article.md --preset infographic-dark-ticket-cn --aspect 21:9`
- `md2wechat generate_infographic --article article.md --preset infographic-handdrawn-sketchnote`

Drafts and image posts:

- `md2wechat create_draft draft.json`
- `md2wechat test-draft article.html cover.jpg`
- `md2wechat create_image_post --help`
- `md2wechat create_image_post -t "Weekend Trip" --images photo1.jpg,photo2.jpg`
- `md2wechat create_image_post -t "Travel Diary" -m article.md`
- `echo "Daily check-in" | md2wechat create_image_post -t "Daily" --images pic.jpg`
- `md2wechat create_image_post -t "Test" --images a.jpg,b.jpg --dry-run`

Writing and humanizing:

- `md2wechat write --list`
- `md2wechat write --style dan-koe`
- `md2wechat write --style dan-koe --input-type fragment article.md`
- `md2wechat write --style dan-koe --cover-only`
- `md2wechat write --style dan-koe --cover`
- `md2wechat write --style dan-koe --humanize --humanize-intensity aggressive`
- `md2wechat humanize article.md`
- `md2wechat humanize article.md --intensity gentle`
- `md2wechat humanize article.md --intensity aggressive`
- `md2wechat humanize article.md --intensity authentic`
- `md2wechat humanize article.md --show-changes`
- `md2wechat humanize article.md -o output.md`

Intensity levels: `gentle` / `medium` (default) / `aggressive` / `authentic`

`authentic` uses a standalone six-dimension writing-quality prompt and bypasses the 24-pattern AI-trace detection used by the other three levels. Use it when the goal is writing that reads like a skilled human — concrete expression, stable tone, no performative depth — rather than just removing AI traces.

## Article Metadata Rules

For `convert`, metadata resolution is:

- Title: `--title` -> `frontmatter.title` -> first Markdown heading -> `未命名文章`
- Author: `--author` -> `frontmatter.author`
- Digest: `--digest` -> `frontmatter.digest` -> `frontmatter.summary` -> `frontmatter.description`

Limits enforced by the CLI:

- `--title`: max 32 characters
- `--author`: max 16 characters
- `--digest`: max 128 characters

Draft behavior:

- If digest is still empty when creating a draft, the draft layer generates one from article HTML content with a 120-character fallback.
- Creating a draft requires either `--cover` or `--cover-media-id`.
- `--cover` is a local image path contract for article drafts. `--cover-media-id` is for an existing WeChat permanent cover asset. Do not assume a WeChat URL or `mmbiz.qpic.cn` URL can be reused as `thumb_media_id`.
- `inspect` is the source-of-truth command for resolved metadata, readiness, and checks.
- `preview` v1 writes a standalone local HTML preview file. It does not start a workbench, write back to Markdown, upload images, or create drafts.
- `convert --preview` is still the convert-path preview flag; it is not the same thing as the standalone `preview` command.
- `preview --mode ai` is degraded confirmation only; it must not be treated as final AI-generated layout.
- `--title` / `--author` / `--digest` affect draft metadata, not necessarily visible body HTML.
- Markdown images are only uploaded/replaced during `--upload` or `--draft`, not during plain `convert --preview`.

## Agent Rules

- Start with discovery commands before committing to a provider, theme, or prompt.
- Route by publish target first: article draft => `convert`; image post / 小绿书 / newspic => `create_image_post`.
- Prefer the confirm-first flow for article work: `inspect` -> `preview` -> `convert` / `--draft`.
- If the user says `小绿书`, `图文笔记`, `图片消息`, `newspic`, or asks for a multi-image post, prefer `create_image_post` even when the source content lives in Markdown.
- Prefer `generate_cover` or `generate_infographic` over a raw `generate_image "prompt"` call when a bundled preset fits the task.
- Validate config before any draft, publish, or image-post action.
- If draft creation returns `45004`, check digest/summary/description before assuming the body content is too long.
- If draft/upload returns `40164 invalid ip ... not in whitelist`, the user's current public IP is not in the MP backend IP whitelist. Ask them to add it at 开发 → 基本配置 → IP 白名单 and retry. Do NOT retry blindly.
- If the user asks for AI conversion or style writing, be explicit that the CLI may return an AI request/prompt rather than final HTML or prose unless the workflow completes the external model step.
- Do not perform draft creation, publishing, or remote image generation unless the user asked for it.
- 高级排版模块（`layout` 命令系列，43 个 `:::block` 模块）**仅在 API 模式下**渲染。本地模式（默认）和 AI 模式都不解析 `:::block` 语法。要用这些模块需显式加 `--mode api`。
- 未开通微信支付的公众号：外部链接 `<a href>` 会被平台剥成纯文本。`--link-style=inline`（默认）保留 URL 作为可复制文本，`--link-style=footnote` 搬到文末统一列表。

## 高级排版决策流（需要 `--mode api`）

> **高级排版模块只在 API 模式下渲染**。当前默认模式是 `local`（本地渲染），**必须显式传 `--mode api`** 才会解析 `:::block` 语法。
> 本地模式（默认）和 AI 模式（`--mode ai`）会把 `:::block` 原样输出或忽略，变成普通段落。
> 如需 API 访问或购买 API Key，请联系作者咨询。

每篇文章按 4 步选模块：

1. **判断内容**：是观点 / 数据 / 教程 / 发布 / 综合？
2. **挑选最少模块**：每个模块只服务这 4 件事之一：
   - **attention**：让读者知道值不值得读（hero / cards / verdict / audience-fit）
   - **readability**：让手机阅读不累（part / toc / label-title / steps）
   - **memorability**：让读者记住一个判断/品牌（verdict / manifesto / author-card）
   - **conversion**：让读者收藏/关注/咨询/转发/购买（cta / subscribe / faq / cases）
3. **发现 + 渲染**：
   ```bash
   md2wechat layout list --serves attention --json
   md2wechat layout show hero --json
   md2wechat layout render hero --var eyebrow=深度观察 --var title="真问题" --json
   ```
4. **校验后发布**：
   ```bash
   md2wechat layout validate --file article.md --json
   ```

**原则**：不要堆模块。一篇文章 hero 只有一个，verdict 只有一个，cta 只有一个。

**API 访问**：高级排版模块是付费 API 功能。如需开通，请联系作者咨询。

## Safety And Transparency

- Reads local Markdown files and local images.
- May download remote images when asked.
- May call external image-generation services when configured.
- May upload HTML, images, drafts, and image posts to WeChat when the user explicitly requests those actions.
