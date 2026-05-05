# WeChat 渲染 knowhow

> 本文档记录 `--mode local` 的设计依据。写给维护者、二次开发的 agent、想定制主题的人看。

## 为什么有 `--mode local`

微信公众号草稿接口 (`add_draft`) 的 HTML sanitizer 强度远超官方文档承诺。我们基于真实 DOM 对比（用 `opencli browser` 抓已发布文章的 `<div id="js_content">`）总结了一份**白名单清单**，发现绝大部分装饰性样式会被剥光。

`api` 模式（服务端渲染）我们不能控制，`ai` 模式（Claude 生成）输出不确定。为了让写作者有一条可预期、可离线、可回溯 bug 的路径，我们引入 `local` 模式——自己 goldmark → 自己内联样式 → 自己处理微信 quirks。

## 微信 HTML 白名单（经验数据）

从实际发布后 DOM 反推（未认证订阅号环境）：

| 标签 | 本身保留？ | 内联 `style` 保留？ | 备注 |
|---|---|---|---|
| `h1`–`h6` | ✅ | ❌ | 结构在，样式全剥 |
| `p` | ✅ | ❌ | 同上 |
| `span` | ✅ | ❌ | 同上 |
| `strong` / `em` | ✅ | ❌ | 加粗/斜体样式由浏览器默认实现 |
| `blockquote` | ✅ | ❌ | 浏览器默认左边缘 + 缩进存活 |
| `ul` / `ol` / `li` | ✅ | ❌ | **每个 `<li>` 前微信自动插一个空 `<li>` 兄弟**，非 URL 相关 |
| `table` / `thead` / `tbody` / `tr` / `th` / `td` | ✅ | ❌ | 最稳的结构化容器 |
| `hr` | ✅ | ❌ | 显示为默认细线 |
| `img` | ✅ | 部分 | src 必须来自 `mmbiz.qpic.cn` 上传 |
| `br` | ✅ | — | 白名单 |
| `section` | ✅ | ❌ | 标签保留，所有样式剥 |
| `div` | ❌ | ❌ | **整个标签被 unwrap，内容包到 `<p>` 里** |
| `a href=` | **依赖账号** | — | 未开通微信支付的订阅号会直接剥掉 href，`<a>` 外层也没了 |
| `style` (元素) / `link` | ❌ | — | 剥除 |
| `script` | ❌ | — | 剥除 |
| `class` / `id` 属性 | ❌ | — | 全部剥除 |

被吃掉的 CSS 属性（即使在白名单标签的 inline style 里）：`color`、`font-size`、`font-weight`、`background`、`border-*`、`padding`、`margin`、`line-height`、`letter-spacing`、`position`、`transform`、`animation`、百分比单位 → 全部吃掉。

**一句话结论**：我们的 `minimal-green` 主题在本地浏览器是绿色精致的，在微信里是"素颜但结构正确"。本地样式只服务**校对预览**，不指望保留到微信。

## 本仓库在 v3/v4 的具体设计决策

### 1. TL;DR callout → 加粗段落而不是 callout 盒子

**尝试过**：`<section style="TLDR_WRAPPER"><div style="BADGE">太长不看版</div> + <table>`
**失败原因**：`<section>` 样式剥、`<div>` unwrap 成 `<p>`，视觉完全失效
**最终设计**：`<p><strong>太长不看版</strong></p>` + 后续块原样保留，不套外壳

代码位置：`internal/render/enhance/tldr.go`

### 2. 章末 takeaway → `<blockquote><p><strong>` 而不是样式化 div

**尝试过**：`<div style="takeaway-heavy">text</div>`
**失败原因**：`<div>` 被 unwrap 成 `<p>`，样式剥光
**最终设计**：把原生 `<blockquote>` 保留，内部套一层 `<strong>`。微信对 `<blockquote>` 的默认样式（左边缘）至少还在。

代码位置：`internal/render/enhance/takeaway.go`

### 3. 代码块缩进 → `&nbsp;` + `<br/>`

**尝试过**：标准 `<pre style="white-space: pre">...<code>` 输出
**失败原因**：微信剥 `white-space` 并把连续空格合并、`\n` 丢弃
**最终设计**：
- 普通空格 `U+0020` → `&nbsp;`（`U+00A0` 不可断空格，作为字符存在，不会被合并）
- Tab → 4 个 `&nbsp;`
- `\n` → `<br/>`

`&nbsp;` 是**字符级**的保留，不依赖 `style`，微信无法合并。`<br/>` 是白名单标签。

代码位置：`internal/render/goldmark.go:renderFencedCodeBlock` / `wechatSafeCode`

### 4. 外链 → `--link-style` 三挡

**约束**：未认证（或未开通微信支付）账号，`<a href>` 会被剥只剩锚文本；URL 彻底丢。

| `--link-style` | 策略 | 适用 |
|---|---|---|
| `inline`（默认）| `[text](URL)` → `text（URL）` | 链接稀疏的文章 |
| `footnote` | 正文 `text[N]`，文末统一 `[N] text — URL` | 链接密集 / 有手写 Reference 段 |
| `native` | 保留 `<a href="URL">` | 已认证 / 开通支付的公众号 |

`footnote` 模式的关键设计：
- 自动识别 `## Reference` / `## References` / `## 参考` / `## 参考链接` / `## 参考资料` / `## 参考文献` / `## 延伸阅读` 这些 heading，**替换它们的正文内容**（保留 heading 本身），避免"手写 Reference + 自动脚注"两份清单并存。
- 脚注列表用 `<p>` + `<br/>` 而不是 markdown 的 `1. item`（会渲染成 `<ol><li>`）。因为 `<ol><li>` 在微信里每个 `<li>` 前都会多一个空 `<li>`，用 `<p>` 直接规避。
- 跳过 inline/fenced code span 内的伪链接（例如 `` `[text](URL)` `` 这种文档示例），不会被误改。

代码位置：`internal/render/links.go`

### 5. 空 `<li>` 的真相（C 假设证伪记录）

**原假设**：微信 auto-linkify 裸 URL → 生成 `<a>` → 策略剥 `<a>` → 留空兄弟元素。

**用 probe 文章做三候选对照实验**（见 `examples/wechat-quirks-probe.md` 第四章）：
- `text（URL）` 对照组
- `text（https:<U+200B>//URL）` ZWSP 候选（破坏 URL 识别）
- `text — URL` 破折号候选（破坏紧邻包围）

**结果**：三者的 bullet 前都出现空 `<li>`。

**结论**：空 `<li>` 是微信 `<ul>` / `<ol>` 默认渲染行为（大概率是其 CSS/后处理脚本加的视觉间距），与 URL、`<a>` 无关。

**应对**：只能绕开 `<ul>/<ol>` 结构。footnote 模式下用 `<p>+<br/>` 代替，消除空 li；其他正常 markdown 列表（用户手写 `- item`）仍会触发，这部分可视为微信的视觉约定。

### 6. 首行 h1 剥离

微信文章页顶部已经有 `<h1 id="activity-name">` 渲染自 draft API 的 title 字段。如果 markdown body 第一行又是 `# 同样的标题`，读者看到标题出现两次。

**修法**：`render.Render(opts)` 接收 `opts.Title`，如果 body 首行 `# xxx` 的 `xxx` 等于 `Title`（`strings.TrimSpace` 比较），则剥掉那行再送进 goldmark。

代码位置：`internal/render/render.go:stripLeadingTitleHeading`

## 验证流程（遇到新 bug 时按这套查）

1. 本地：`./md2wechat convert article.md --output /tmp/x.html`，浏览器打开对着 markdown 源看是否逻辑正确。
2. 上传到草稿：`./md2wechat convert article.md --upload --draft --cover <图> --digest <≤128>`，拿 `draft_id`。
3. 在微信公众号后台点"预览"获取临时链接，用 `opencli browser open <预览链接>` + `opencli browser state > /tmp/dom.txt` 抓 DOM。
4. 对比 DOM 与本地 HTML：grep 关键片段，看哪些标签/style 被剥、哪里多了空元素。
5. 写最小复现到 `examples/wechat-quirks-probe.md`（或扩展现有探针），加回归测试。

## 参考资料 (外部 knowhow)

- [bm.md 技术原理（CSS 内联化、微信限制、渲染架构）](https://www.verysmallwoods.com/blog/20260119-wechat-markdown-copy-paste)
- [doocs/md 微信 Markdown 编辑器](https://md.doocs.org/)（"外链转脚注" 功能的参考实现）
- [微信开放社区官方帖：content HTML 样式乱的原因](https://developers.weixin.qq.com/community/develop/doc/00046c6febcd282048a0492116b800)（官方透露：未开通微信支付无法插入外链；双引号要改单引号才能通过 API 校验）
- [jimmysong - Markdown 一键发布到微信的经验](https://jimmysong.io/zh/blog/markdown-wechat-publication/)（SVG→JPG、图床必须用微信 material、hugo shortcode 过滤）
- [axiaoxin - 外链转脚注](https://blog.axiaoxin.com/post/md2wechat/)
- [MD2WeChat - 代码块缩进方案](https://github.com/Mapoet/MD2WeChat/blob/main/docs/USAGE.md)

## 本仓库的视觉源头

`themes/minimal-green/minimal-green-theme-detail.mhtml` 是 md2wechat.app 的 `minimal-green` 主题页的 Chrome 归档。Theme 的 inline style 全部从这份 MHTML 提取，提取后硬编码进 `internal/render/themes/minimal_green.go`。想改配色/字号，先看这份归档，再改 style map。

## 本次 session 没做的事（供以后接力的人）

- **SVG 图片自动转 JPG**：微信 material API 不吃 SVG。当前 CLI 会失败，用户需要手动转。bm.md 的 knowhow 提到了这条。
- **Checkbox 任务列表**：`- [ ]` / `- [x]` 在微信里不支持 `<input type=checkbox>`，要换成 `☐` / `☑` Unicode。
- **URL 双引号编码问题**：微信官方帖提到 content 里双引号要改单引号。我们用 `silenceper/wechat` SDK 时没遇到，看起来 SDK 自己做了 JSON escape；保持监控。
- **其他主题本地化**：除了 `minimal-green`，其他 30+ v2.0 主题还是要走 `--mode api`。没有迁移的 backlog。
- **`<ul>/<ol>` 空 `<li>` 彻底解法**：目前只有 footnote 模式绕开。普通用户 `- item` 列表仍有这个视觉副作用，没有根治方案。
