# Syntex

[![Go Report Card](https://goreportcard.com/badge/github.com/jbwfu/syntex)](https://goreportcard.com/report/github.com/jbwfu/syntex)
[![Build Status](https://github.com/jbwfu/syntex/actions/workflows/go.yml/badge.svg)](https://github.com/jbwfu/syntex/actions/workflows/go.yml)
[![Latest release](https://img.shields.io/github/v/release/jbwfu/syntex)](https://github.com/jbwfu/syntex/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

[[English](./README.md)] | 简体中文

Syntex 是一个命令行工具，用于将多个源代码文件打包成一个单一的、上下文丰富的文本文件。它专为大型语言模型（LLM）设计，在非交互式环境中为其提供项目代码的完整快照。

### 功能特性

-   **专注打包:** 专注于将代码文件内容整合为 Markdown 或 Org-Mode 格式，并附加上下文信息。
-   **智能过滤:** 默认遵守 `.gitignore` 规则，并能自动跳过二进制文件，确保输出内容纯净。
-   **管道友好:** 深度集成 Unix 管道理念，可与 `find`, `fd` 等工具无缝协作，实现复杂的查询与打包工作流。
-   **开箱即用:** 基于 Go 语言构建的单一原生二进制文件，无需任何运行时依赖。

---

## 安装

推荐的方式是从 [**Releases**](https://github.com/jbwfu/syntex/releases/latest) 页面下载最新的预编译二进制文件。

<details>
<summary>其它安装方式</summary>

**使用 `go install`：**
```sh
go install github.com/jbwfu/syntex@latest
```

**从源码构建：**
```sh
git clone https://github.com/jbwfu/syntex.git
cd syntex
make build
sudo make install
```
</details>

---

## 使用方法

`syntex` 遵循 Unix “组合优于内建”的设计哲学。我们强烈建议您将其与 `fd` 或 `find` 等专业文件查找工具结合使用，通过管道传递文件列表。

```sh
# 推荐：与 fd 结合使用，查找所有 .go 和 .md 文件并打包
fd . -e go -e md --print0 | syntex -0 -o context.md

# 经典：与 find 结合使用
find . -name "*.go" -o -name "*.md" -print0 | syntex -0 --clipboard

# 简单场景：使用内建 glob
syntex 'src/**/*.go'
```

-   `-0` / `--print0` 选项组合可以安全地处理包含特殊字符的文件名。
-   使用 `-o <file>` 将结果写入文件，或使用 `-c` / `--clipboard` 复制到剪贴板。

您可以使用 `--dry-run` 标志来预览哪些文件将被打包，而不会实际生成任何输出。这对于验证您的文件查找模式和忽略规则非常有用。

```sh
$ syntex cmd --dry-run
```

输出:
```
[Dry Run] Planning to process files using the 'markdown' format:
go  cmd/syntex/main.go
go  cmd/syntex/options/options.go

[Dry Run] Total: 2
```

---

## 贡献

欢迎任何形式的贡献！如果您发现 Bug 或有功能请求，请随时创建 Issue 或提交 Pull Request。

## 许可证

本项目采用 **MIT 许可证** 授权。
