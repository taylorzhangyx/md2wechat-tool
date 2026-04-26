# Layout Module Catalog 设计文档

- 日期：2026-04-26
- 主题：在 md2wechat-skill CLI 中暴露 `/api/convert` 的 38+ 个高级排版模块，使 Agent 能够发现、选型、渲染并校验 `:::block` 语法
- 范围：单一 PR 可交付的能力面；为未来行业模块包扩展预留架构
- 终态：`md2wechat layout {list,show,render,validate}` 四个子命令 + `internal/assets/builtin/layout/` 模块字典 + 两套 SKILL.md 增加"排版 4 步决策流"

---

## 1. 问题陈述

### 1.1 商业背景

`/api/convert` 的真实差异点是 38+ 个 `:::模块` 排版 DSL（hero / verdict / infographic / cta 等），不是基础 Markdown→HTML 转换。这些模块是商业护城河：任何开源渲染器都能做基础转换，但只有 md2wechat.cn 能产出服务"读者四件事"的结构化模块。

### 1.2 当前 Agent 路径上的缺口

通过代码审计确认（`internal/`、`cmd/`、`themes/` 目录全量 grep）：md2wechat-skill CLI 当前对 `:::模块` 语法**完全无感知**：

- `discovery` 命令暴露 themes / providers / prompts，**没有 layout module 目录**
- `convert` 把 Markdown 直接透传给 API
- 唯一的模块手册（1370 行 `advanced-layout-modules-guide.md`）住在另一个仓库 `wechat-markdown-editor` 中，CLI 不知道它的存在

**直接后果**：Agent 在 API 模式下默认产出**纯 Markdown**。客户付了 API 费用，但拿到的是被竞品轻易复制的通用排版。**护城河存在于产品里，但不存在于 Agent 能调用的 CLI 范围里。**

### 1.3 设计目标

让 Agent 能像顶级排版大师一样工作：先判断内容，再选最少但最准确的模块，每个模块服务"读者四件事"之一：

1. **attention** — 让读者知道值不值得读
2. **readability** — 让手机阅读不累
3. **memorability** — 让读者记住一个判断、一个人、一个品牌
4. **conversion** — 让读者愿意收藏、关注、咨询、转发或购买

### 1.4 非目标（明确不做）

- **不做 `suggest` 命令**。Agent 自己是顶级判断器；CLI 写规则只会用劣质判断替代优质判断。判断力的归宿是模块元数据 + SKILL.md 决策流提示词，不是 Go 算法。详见 [附录 A：suggest 反向论证](#附录-asuggest-的反向论证)
- **不做行业模块包（pack）的命令面**。v1 只在元数据里把 `industry` / `tags` / `pairs_well_with` 字段就位，pack 作为 v2 packaging 增强
- **不在 convert 中默认开启 validate**。validate 是独立子命令，让 Agent 显式控制反馈环

---

## 2. 架构

### 2.1 命令面

```
md2wechat layout list    [--serves <p>] [--position <pos>] [--content-type <t>]
                         [--industry <i>] [--tag <t>] [--pairs-with <name>] [--json]
md2wechat layout show    <name> [--json]
md2wechat layout render  <name> --var KEY=VALUE [...] [--json]
md2wechat layout validate <file.md> [--json]
```

设计语义：

- **list**：返回模块清单（默认人类可读表格；`--json` 返回 JSON envelope）。所有过滤是 AND 关系
- **show**：返回单个模块的完整 schema（字段、示例、决策提示）
- **render**：根据模块名 + 变量，输出合规的 `:::block` 文本到 stdout（可被管道注入文章）
- **validate**：扫描 Markdown 文件中的所有 `:::block`，按字典里的 `fields.required` / `fields.optional` 校验；未知模块按 S1 策略 **warn**（exit code 0，但 stderr/JSON 中带 warning 列表）

### 2.2 目录结构（Mirror prompt catalog 模式）

```
internal/assets/builtin/layout/
  opening/
    hero.yaml
    cards.yaml
    audience-fit.yaml
    part.yaml
    toc.yaml
    label-title.yaml
  judgment/
    verdict.yaml
    manifesto.yaml
    myth-fact.yaml
    bridge.yaml
  infographic/
    infographic.yaml
  evidence/
    metrics.yaml
    compare.yaml
    steps.yaml
    timeline.yaml
    quote.yaml
  image/
    image-text.yaml
    image-compare.yaml
    image-annotate.yaml
    image-steps.yaml
  brand/
    author-card.yaml
    series.yaml
    subscribe.yaml
  conversion/
    cta.yaml
    faq.yaml
    pricing.yaml
    cases.yaml
    checklist.yaml
    notice.yaml
    summary.yaml
    people.yaml
    logos.yaml
    toolbox.yaml
    specs.yaml
  layout-extra/
    split.yaml
    flow.yaml
    matrix.yaml
    dialogue-pair.yaml
    gallery.yaml
    dialogue.yaml
    longimage.yaml
    resource-list.yaml
    changelog.yaml
    comparison-table.yaml

internal/layoutcatalog/
  catalog.go      # 索引、过滤
  loader.go       # override 链加载
  schema.go       # YAML schema 类型
  renderer.go     # render 子命令实现
  validator.go    # validate 子命令实现
  catalog_test.go
  loader_test.go
  renderer_test.go
  validator_test.go

cmd/md2wechat/
  layout.go       # cobra 命令树
  layout_test.go
```

最终模块清单以翻译实施时的 1370 行手册为准；上面的归类是骨架（不要求与最终精确一致，分组允许在实施时按手册重新切分）。

### 2.3 Loader Override 链（与 prompt catalog 同构）

优先级从高到低：

1. `MD2WECHAT_LAYOUT_DIR`（环境变量）
2. `./layout/`（项目级覆盖）
3. `~/.config/md2wechat/layout/`（用户级覆盖）
4. 内置 assets（embed.FS）

同名模块按优先级覆盖（不合并字段，整份替换），保留 `metadata.provenance` 标记来源（builtin / user / project / env）。

### 2.4 模块 YAML Schema

```yaml
schema_version: 1
name: hero
version: 1.0.0
since: api/v1
deprecated_in: ""              # 可选；非空时 list 默认隐藏
replaced_by: ""                # 可选；与 deprecated_in 配套

category: opening              # opening / judgment / infographic / evidence /
                               #   image / brand / conversion / layout-extra
serves: [attention, memorability]      # 1..N，必填，枚举：
                                       # attention / readability /
                                       # memorability / conversion
content_types: [opinion, launch, system]   # 自由标签数组
industry: [general]                        # 默认 general，垂直 pack 写行业 tag
tags: [eye-catching, brand]                # 自由标签
position: opening              # opening / body / closing / any
when_to_use: 文章开头需要快速建立判断和品牌识别
when_not_to_use: 短消息、纯教程、不需要建立立场的资讯
pairs_well_with: [audience-fit, verdict, cta]
avoid_combining_with: [label-title]
anti_pattern: 一篇文章用多个 hero

# 字段定义驱动 render 与 validate
fields:
  required:
    - name: title
      type: string
      hint: 主标题
  optional:
    - name: eyebrow
      type: string
      hint: 标签/眼眉
    - name: subtitle
      type: string
    - name: cta_label
      type: string

# 行格式模块（cards/metrics/toc 等）使用 rows 而非 fields
rows:
  delimiter: "|"
  min_columns: 3
  schema:
    - name: label
      required: true
    - name: title
      required: true
    - name: description
      required: true
    - name: style
      required: false
      enum: [accent, default]

example: |
  :::hero
  eyebrow: 系统升级
  title: 让普通人也有自己的内容系统
  :::

metadata:
  author: md2wechat-team
  provenance: builtin
  inspired_by: wechat-markdown-editor/docs/advanced-layout-modules-guide.md
```

**Schema 约束**：

- `name` 必填且全字典唯一；`schema_version: 1` 必填
- `serves` 至少 1 项；枚举值固定为四件事英文标准名
- `fields` 与 `rows` 互斥（每个模块只能是字段型或行格式型）
- `infographic` 因有 `type` 子类型回退保护，单独在 schema 中允许 `variants` 数组（实施时细化）

### 2.5 Render 实现策略

- **不引入 Go template**。Render 是声明式拼装：
  - `fields` 型模块：拼 `key: value` 行
  - `rows` 型模块：按 `delimiter` 拼列
  - 标题型可选 `[标题]` 后缀按模块声明
- 缺必填字段时返回非零 exit code 并在 stderr 给出明确错误；JSON 模式下 `success: false` + `error.code: missing_required_field`
- 渲染输出始终是 UTF-8、LF 换行、不带尾随空白

### 2.6 Validate 实现策略

- 用一个简单的 `:::name ... :::` 扫描器（不调用完整 Markdown parser）解析所有顶层 directive 块
- 对每个块查字典：
  - **未知模块**：warn，不阻塞（S1 策略）。warning 列表带 `module_name` / `line_number` / `reason: not_in_catalog`
  - **已知模块**：按 `fields` / `rows` 校验，缺字段/列数不足/枚举越界都报 error
- exit code：有 error → 1；只有 warning → 0
- JSON envelope：
  ```json
  {
    "success": true,
    "schema_version": 1,
    "data": {
      "blocks_total": 12,
      "blocks_valid": 10,
      "blocks_unknown": 1,
      "errors": [...],
      "warnings": [...]
    }
  }
  ```

### 2.7 JSON Envelope 契约

所有 `--json` 输出严格遵循现有 envelope schema_version 1：

```json
{
  "success": true,
  "code": "ok",
  "message": "...",
  "schema_version": 1,
  "status": "...",
  "retryable": false,
  "data": { ... },
  "error": null
}
```

错误码命名（layout 子树）：
- `layout_module_not_found`
- `layout_invalid_filter`
- `layout_missing_required_field`
- `layout_invalid_field_value`
- `layout_validate_has_errors`

---

## 3. SKILL.md 增量

两套 SKILL.md（`skills/md2wechat/SKILL.md` 与 `platforms/openclaw/md2wechat/SKILL.md`）顶部加入"**排版选型 4 步决策流**"段落。这是顶级排版师判断力真正的容器：

```
排版选型 4 步：

1. 判断内容类型：观点 / 复盘 / 教程 / 发布稿 / 商业稿 / 资讯
2. 按读者四件事，各选 0-1 个模块（少即多，不是必选）：
   a. attention   — 让读者知道值不值得读
   b. readability — 让手机阅读不累
   c. memorability — 让读者记住一个判断/品牌
   d. conversion  — 让读者愿意收藏/关注/咨询/转发/购买
3. 调 `md2wechat layout list --serves <p> --content-type <t> --json`
   过滤候选；调 `md2wechat layout show <name> --json` 看字段约束
4. 调 `md2wechat layout render <name> --var ...` 生成合规 :::block；
   写完整篇后用 `md2wechat layout validate file.md` 离线校验
```

不写"必须使用某个模块"的硬规则，把决策权留给 Agent。

---

## 4. 数据迁移：1370 行手册 → 38+ 份 YAML

- **来源**：`/Users/geekjourney/Workspace/cursor/wechat-markdown-editor/docs/advanced-layout-modules-guide.md`
- **方式**：人工/半自动翻译，每个模块一份 YAML
- **进度跟踪**：用会话 SQL `todos` 表逐模块跟踪（done / in_progress / blocked）
- **质量门**：每份 YAML 必须通过 schema 测试（required 字段齐全、`serves` 枚举有效、example 可被 render 反向解析等价）
- **drift 防御**：在 `Makefile` 或 release checklist 中加一条人工审校项："对照 advanced-layout-modules-guide.md 当前版本，检查 layout 字典是否需要补充新模块"。v1 不做自动 drift CI（避免引入跨仓库依赖）

---

## 5. 测试矩阵（按测试纪律分级，三层闭环）

测试分为三层：**单元测试**（默认 CI 跑）、**集成测试**（默认 CI 跑，全本地）、**E2E**（默认 skip，需要本地或 CI 启动 API 服务才跑）。三层都必须有；任一缺失都不视为完成。

### 5.1 单元测试（Unit）

完全无外部依赖，`go test ./...` 默认运行。

#### 5.1.1 CLI 契约（`cmd/md2wechat/layout_test.go`）

- `list`：合法过滤组合返回结果；非法 `--serves` 值报错；`--json` envelope 严格符合 schema_version 1
- `show`：存在/不存在模块；`--json` 与人类格式 data 一致
- `render`：必填齐全成功；缺必填失败；行格式列数不足失败；JSON 错误码符合 §2.7 契约
- `validate`：error / warning / 全合规三种 fixture；exit code 与 JSON 一致

#### 5.1.2 Loader override（`internal/layoutcatalog/loader_test.go`）

- 4 级 override 链按优先级生效
- 同名模块完全覆盖（不合并字段）
- `metadata.provenance` 正确标注来源

#### 5.1.3 Schema 完整性（`internal/layoutcatalog/catalog_test.go`）

- 所有内置 YAML 都能加载
- `name` 全字典唯一
- `serves` 枚举严格校验
- `fields` 与 `rows` 互斥
- `pairs_well_with` / `avoid_combining_with` 引用的模块名都在字典中（防引用悬挂）

#### 5.1.4 Render 反向不变量（`internal/layoutcatalog/renderer_test.go`）

- 对每个内置模块的 `example`，用 example 的字段值跑 render，输出与 example 等价（normalize 空白后）
- 这条不变量是 catalog drift 的最低防线

### 5.2 集成测试（Integration，全本地）

跨包串通、无外部 HTTP 依赖。`go test ./...` 默认运行。

#### 5.2.1 catalog → renderer → validator 闭环（`internal/layoutcatalog/integration_test.go`）

对每个内置模块：
1. 从 example 提取字段值
2. 调 renderer 生成 `:::block` 文本
3. 把生成的文本喂给 validator
4. 必须 0 error / 0 warning

这是单一 catalog 数据驱动三个能力的"自洽性"证明。

#### 5.2.2 多模块组合 fixture（`internal/layoutcatalog/integration_test.go`）

构造 3 份代表性 markdown fixture（`testdata/articles/`）：
- `opinion-piece.md`：hero + audience-fit + verdict + manifesto + cta
- `data-report.md`：cards + metrics + compare + summary + cta
- `mixed-with-unknown.md`：合规模块 + 一个故意写错的 + 一个未知模块

各自跑 validate，断言 errors / warnings / blocks_total 符合预期。

#### 5.2.3 SKILL.md 决策流引用一致性（`cmd/md2wechat/skill_consistency_test.go`）

简单 grep 两套 SKILL.md，断言：
- "排版选型 4 步" 段落存在
- 引用的所有命令都能在 cobra 命令树中解析

防止 SKILL.md 与 CLI 实际能力漂移。

### 5.3 E2E 测试（默认 skip，需要 API 服务）

通过 `MD2WECHAT_E2E=1` 环境变量门控。本地有 `http://localhost:3000` 的 API convert 服务时启用；CI 默认 skip（除非 CI 也起了 API 服务）。

放在 `cmd/md2wechat/e2e_layout_test.go`，每个 test 顶部 `if os.Getenv("MD2WECHAT_E2E") != "1" { t.Skip(...) }`。

API 端点通过 `MD2WECHAT_BASE_URL` 配置（默认 `http://localhost:3000`），`MD2WECHAT_API_KEY` 通过本地 `.env` 或 shell 注入。

#### 5.3.1 单模块 E2E

对每个内置模块：
1. 用 renderer 生成最小合规 `:::block`
2. 通过 `md2wechat convert --mode api --base-url $MD2WECHAT_BASE_URL` 真调本地 API
3. 断言：HTTP 200、返回 HTML 非空、HTML 中包含模块预期的 hook（如 hero 模块的 HTML 中包含 title 文本）

E2E 失败模式说明 API 端字段约束跟 CLI 字典 drift；按 §7.1 流程修字典或开 issue。

#### 5.3.2 完整文章 E2E

用 §5.2.2 的 `opinion-piece.md` 与 `data-report.md` 两份 fixture：
1. 先跑 `md2wechat layout validate` 通过
2. 再跑 `md2wechat convert --mode api` 通过
3. 比较 HTML 输出与 baseline snapshot（首次运行写入；后续比对，差异需人工 review）

baseline snapshot 文件提交到 `cmd/md2wechat/testdata/e2e_snapshots/`，drift 时手动重新生成。

#### 5.3.3 Validator 与 API 错误的一致性 E2E

构造 3 个故意错误的 fixture（缺必填 / 字段越界 / 未知模块），断言：
- 本地 validator 报错的，API 端也拒绝（或对未知模块都接受）
- 本地 validator 通过的，API 端必须接受

这条是 S1 策略（CLI 是 API 子集快照）的真实闭环验证。失败说明 CLI 字典对 API 校验语义有误读。

### 5.4 测试运行命令（实施时固化进 Makefile）

```bash
# 单元 + 集成（CI 默认）
GOCACHE=/tmp/md2wechat-go-build go test ./...

# 含 E2E（本地 API 服务起来后）
MD2WECHAT_E2E=1 \
MD2WECHAT_BASE_URL=http://localhost:3000 \
MD2WECHAT_API_KEY=$LOCAL_KEY \
GOCACHE=/tmp/md2wechat-go-build go test ./cmd/md2wechat -run E2E -v

# 完整 gate
make quality-gates
```

E2E 不进 `make quality-gates` 默认链路（CI 不一定有 API 服务），但作为 release 前的 manual gate 列入 §8 acceptance。

### 5.5 不做的测试

- **不为 coverage 数字加测试**
- **不测 Markdown parser 边界**（validate 用的是简单扫描器，已在 5.1.1 / 5.2.2 用 fixture 覆盖）
- **不测 SKILL.md 全文内容**（仅做决策流段落存在性 + 命令引用有效性的 5.2.3 断言）
- **不在 CI 默认链路启动 API 服务**（避免环境耦合）

---

## 6. 文档同步清单（CLI 改动必须同步）

按 AGENTS.md 规定，本次涉及 CLI command + JSON 输出 + 新概念，必须同步审校以下文件：

- `README.md` — 顶部 features 段加 layout 能力 + Quick start 一条命令示例
- `docs/DISCOVERY.md` — 新增 "Layout Modules" 章节，覆盖四个子命令
- `docs/FAQ.md` — 加入"如何让 Agent 用上高级排版模块"FAQ
- `skills/md2wechat/SKILL.md` — "排版 4 步决策流"段落
- `platforms/openclaw/md2wechat/SKILL.md` — 同上（保持两端一致）
- `CHANGELOG.md` — 新版本章节

---

## 7. 风险与未来扩展

### 7.1 已知风险

| 风险 | 缓解 |
|---|---|
| 38+ YAML 翻译工作量大 | 用 SQL todos 分批跟踪；先做开场+收尾共 ~12 个验证架构，再批量补全 |
| API 端新增模块时 CLI 字典滞后 | S1 策略：validate warn 不阻塞；release checklist 中人工审校 drift |
| 模块字段约束与 API 真实校验有出入 | example 反向不变量测试 + 真实 smoke 一条最小闭环；drift 时以 API 行为为准 |
| Agent 误以为 CLI 字典是 API 全集 | SKILL.md 与 `--json` 输出明确标注"字典覆盖至 vX.Y"，未知模块 warn 时给出说明 |

### 7.2 v2+ 扩展点（明确不在本 spec 范围）

- `md2wechat layout packs {list,install,show}` — 行业模块包（金融/教育/电商等）
- 远程 pack registry（类似 npm scope）
- `md2wechat layout templates {list,show,apply}` — 整篇组合模板（如"金融观点文标准结构"）
- 自动 drift CI：跨仓库对照 `advanced-layout-modules-guide.md` 与 layout 字典
- `suggest` 命令（仅当出现真实数据表明需要时再考虑，参见附录 A）

---

## 8. Acceptance Criteria

完整 Done 的判定：

1. `md2wechat layout {list,show,render,validate}` 四个子命令可运行，`--help` 文档完整
2. `internal/assets/builtin/layout/` 包含 38+ 份 YAML，全部通过 schema 测试
3. **单元测试**（§5.1）全部通过
4. **集成测试**（§5.2）全部通过
5. **E2E 测试**（§5.3）在本地 `MD2WECHAT_BASE_URL=http://localhost:3000` 环境下全部通过；E2E 失败即视为 release blocker
6. `make quality-gates` 全绿（不含 E2E）
7. release 前手动跑过一次 E2E 全套（§5.4 命令）并记录在 PR 描述
8. 文档同步清单中所有文件审校并更新
9. 两套 SKILL.md 都包含"排版 4 步决策流"段落
10. CHANGELOG.md 记录本次新增能力

---

## 附录 A：suggest 的反向论证

记录此设计**为什么不做** `suggest` 命令，避免未来重复讨论：

1. **Agent 是顶级判断器**：调用 CLI 的 Agent 是 GPT/Claude 大模型；在 Go 里写规则建议本质是用劣质判断替代优质判断；用 LLM 实现 suggest 则等于多走一层中介
2. **建议会变成 Agent 的天花板**：Agent 大概率照建议执行，反而压缩了"顶级排版师"该有的发挥空间
3. **判断力的正确归宿是 prompt**：模块元数据 + SKILL.md 决策流就是判断力的容器；写进 Go 算法等于把活的判断力固化成死规则
4. **维护成本指数级**：模块从 38 增长到 100+ 时，规则爆炸；A+B 路线只需新增 YAML
5. **Agent-friendly ≠ Agent-dependent**：CLI 是工具箱不是半成品大脑

**重新评估 suggest 的触发条件**（任一满足才考虑）：

- 真实数据显示 Agent 选错/漏用模块
- 出现非 LLM 调用方（脚本/工具集成）需要离线决策
- 模块数量爆炸到人类/Agent 都难以全量阅读
