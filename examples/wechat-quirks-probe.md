---
title: 微信排版探针：对 md2wechat 做一次全链路压力测试
author: md2wechat
digest: 这是一篇用来测试 md2wechat 渲染边界的文章——每一个段落都在测一个已知或未知的微信排版坑。如果你看到了各种奇怪符号和编号，那是实验的副作用，不是错别字。
---

# 微信排版探针：对 md2wechat 做一次全链路压力测试

这篇文章不是给读者看的，是给工具看的。它故意把每一个已知会出问题的 markdown 结构都放一遍，用来验证 `md2wechat` 的本地渲染和微信最终呈现是否一致。

太长不看版：

| 实验类型 | 触发点 | 期望现象 |
|---|---|---|
| A. footnote 外链 | `[text](URL)` → `text[1]` + 末尾列表 | 正文无 `（URL）`，脚注段在末尾 |
| B. 代码块缩进 | `&nbsp;` + `<br>` | 缩进不塌，换行保留 |
| C. 反 auto-linkify | 三种 URL 包裹对照 | 至少一种消除空 `<li>` |

---

## 第一章：shell 代码块 + 章末 takeaway

典型的 shell 脚本示例：

```bash
#!/usr/bin/env bash
set -euo pipefail

for name in alice bob charlie; do
    echo "hello, ${name}"
    if [[ "${name}" == "bob" ]]; then
        echo "  bob is special"
    fi
done
```

这个代码块里有 4 空格的嵌套缩进、空行（上面没有，但下面那段 python 有）、字符串和控制符。如果微信把它挤成一坨，说明 B 没做到位。

> 代码块的结构稳定性，是写技术文章时最不能妥协的一环。

---

## 第二章：多语言代码混合

Python 示例，带 tab 缩进、空行、注释：

```python
def summarize(items):
    """Return counts grouped by status."""
    buckets = {}
    for it in items:
        key = it.get("status", "unknown")

        # bump the bucket, defaulting to zero
        buckets[key] = buckets.get(key, 0) + 1
    return buckets


if __name__ == "__main__":
    print(summarize([{"status": "ok"}, {"status": "fail"}]))
```

Go 版本，带制表符 + 空行：

```go
package main

import (
	"fmt"
	"strings"
)

func greet(name string) string {
	if name == "" {
		return "hello, stranger"
	}
	return fmt.Sprintf("hello, %s", strings.ToLower(name))
}

func main() {
	fmt.Println(greet("Taylor"))
}
```

> Go 的 gofmt 强制用 tab 缩进——这对微信渲染器是双重挑战：既要保留 tab 又要显示出"层级感"。

---

## 第三章：外链密度测试（A 和 C 的主战场）

下面这段故意把 10 个外链塞进相邻几段，测 `footnote` 模式下 `text[1]` 在中英文之间的贴合。

研究 LLM benchmark 的入口文献有不少：最权威的综述是 [Measuring Massive Multitask Language Understanding](https://arxiv.org/abs/2009.03300)，后来被 [GPQA](https://arxiv.org/abs/2311.12022) 推到了研究生水平，再进一步到 [Humanity's Last Exam](https://arxiv.org/abs/2501.14249) 就彻底放弃和人类拉平的念头了。

代码能力那一路从 [SWE-bench](https://arxiv.org/abs/2310.06770) 起步，延伸到 [Terminal-Bench](https://www.tbench.ai/) 和它的 [git-multibranch 样题](https://www.tbench.ai/registry/terminal-bench-core/head/git-multibranch)。Agent 操作网页相关的有 [WebArena](https://webarena.dev/og/) 和 [OSWorld](https://os-world.github.io/)。

再往后走就是综合评测，比如 [BrowseComp](https://openai.com/index/browsecomp/) 和 [GAIA 数据集](https://huggingface.co/datasets/gaia-benchmark/GAIA)，以及用户主观偏好导向的 [Chatbot Arena 论文](https://arxiv.org/abs/2403.04132)。

链接一旦密度高，`（URL）` inline 模式会非常冗余；这段就是用来直观感受 footnote 模式读起来是不是更顺。

---

## 第四章：URL 三候选对照（C 的核心实验）

下面这个列表的三条中故意用**不同的 URL 包裹格式**。我们本地渲染后不再变换，直接丢给微信渲染器；发到草稿再抓 DOM，看哪一条的 bullet 前不再出现空 `<li>`。

- `[CONTROL]` 对照组：GAIA dataset: gaia-benchmark/GAIA（https://huggingface.co/datasets/gaia-benchmark/GAIA）
- `[ZWSP]` 零宽空格候选：GPQA example: research-paper（https:​//arxiv.org/abs/2311.12022）
- `[DASH]` 破折号候选：HLE dataset — https://arxiv.org/abs/2501.14249

（你看到的 `[CONTROL]` / `[ZWSP]` / `[DASH]` marker 是故意留的 DOM grep 锚点，文章验证完会删。）

> 这一章的每一行都是实验样本，不是写给读者看的。

---

## 第五章：blockquote 与嵌套

一个**中段单行**引用（不应该触发 takeaway，因为后面不是 hr/h2/EOF）：

> 模型评测的噪声来源之一是评分员分歧本身。

讨论还没结束。

一个**多行**引用（同样不应该触发 takeaway）：

> 首先，任何 benchmark 的数值都只是一个压缩投影。
>
> 其次，能在 benchmark 高分 ≠ 能在用户场景高分。
>
> 最后，用户场景高分 ≠ 商业上赚钱。

上面三段只是一个讨论块，不是结论。

一个**嵌套**引用：

> 原作者说：
>
> >> 我把 benchmark 设计成这样，就是要让模型在没见过的任务类型上暴露短板。

嵌套引用是 markdown 里相对罕见的语法，放这里看 goldmark + 我们的 renderer 有没有正确处理。

---

## 第六章：表格 + 分割线

这是一个四列表格：

| Benchmark | 测什么 | 评分方式 | 外部工具 |
|---|---|---|---|
| MMLU | 学科知识 | 选择题 | 无 |
| SWE-bench | 真实 bug 修复 | 测试通过率 | git + pytest |
| WebArena | 浏览器任务 | 最终状态 | 完整浏览器 |
| Chatbot Arena | 对话偏好 | 两两比较 | 真人 |

---

表格上下各一个 `---` 分隔线，用来确认 hr 样式稳定。

---

## 第七章：收尾

> md2wechat 的每一次发版，都是和微信渲染器打一场新的仗——对方的过滤规则不公开、偶尔变动，唯一可靠的是 DOM 证据。

这句是结论，要被升级成 takeaway。

---

## Reference

下面是 5 条参考链接，结构**故意和 Taylor 那篇文章完全一样**，用来对比两次发布之间 Reference 段的 DOM 行为。

- MMLU 论文: [Measuring Massive Multitask Language Understanding](https://arxiv.org/abs/2009.03300)
- GPQA 论文: [A Graduate-Level Google-Proof Q&A Benchmark](https://arxiv.org/abs/2311.12022)
- SWE-bench 论文: [Can Language Models Resolve Real-World GitHub Issues?](https://arxiv.org/abs/2310.06770)
- Chatbot Arena 论文: [An Open Platform for Evaluating LLMs by Human Preference](https://arxiv.org/abs/2403.04132)
- GAIA 数据集: [gaia-benchmark/GAIA](https://huggingface.co/datasets/gaia-benchmark/GAIA)
