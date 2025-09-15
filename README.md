# Syntex

[![Go Report Card](https://goreportcard.com/badge/github.com/jbwfu/syntex)](https://goreportcard.com/report/github.com/jbwfu/syntex)
[![Build Status](https://github.com/jbwfu/syntex/actions/workflows/go.yml/badge.svg)](https://github.com/jbwfu/syntex/actions/workflows/go.yml)
[![Latest release](https://img.shields.io/github/v/release/jbwfu/syntex)](https://github.com/jbwfu/syntex/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

English | [[简体中文](./README_zh-CN.md)]

Syntex is a command-line tool for packing multiple source code files into a single, context-rich text file. It is designed specifically for Large Language Models (LLMs) to provide them with a complete snapshot of a project's codebase in non-interactive environments.

### Features

-   **Focused Packing:** Specializes in consolidating file contents into Markdown or Org-Mode formats with contextual metadata.
-   **Intelligent Filtering:** Respects `.gitignore` rules by default and automatically skips binary files to ensure clean output.
-   **Pipeline-Friendly:** Built on the Unix philosophy, it works seamlessly with tools like `find` and `fd` to create powerful packing workflows.
-   **Works Out-of-the-Box:** A single, native binary built with Go, requiring no runtime dependencies.

---

## Installation

The recommended way is to download the latest pre-compiled binary from the [**Releases**](https://github.com/jbwfu/syntex/releases/latest) page.

<details>
<summary>Other Installation Methods</summary>

**With `go install`:**
```sh
go install github.com/jbwfu/syntex@latest
```

**From Source:**
```sh
git clone https://github.com/jbwfu/syntex.git
cd syntex
make build
sudo make install
```
</details>

---

## Usage

`syntex` follows the Unix philosophy of "composition over built-in features." We strongly recommend using it in combination with professional file-finding tools like `fd` or `find`, piping the file list to `syntex`.

```sh
# Recommended: Use with fd to find all .go and .md files and pack them
fd . -e go -e md --print0 | syntex -0 -o context.md

# Classic: Use with find
find . -name "*.go" -o -name "*.md" -print0 | syntex -0 --clipboard

# Simple cases: Use the built-in glob
syntex 'src/**/*.go'
```

-   The `-0` / `--print0` combination safely handles filenames with special characters.
-   Use `-o <file>` to write the result to a file, or `-c` / `--clipboard` to copy it to the clipboard.

You can use the `--dry-run` flag to preview which files will be packed without actually generating any output. This is very useful for verifying your find patterns and ignore rules.

```sh
syntex cmd --dry-run
```

output:
```
[Dry Run] Planning to process files using the 'markdown' format:
go  cmd/syntex/main.go
go  cmd/syntex/options/options.go

[Dry Run] Total: 2
```

---

## Contributing

Contributions are welcome! Please feel free to open an issue or submit a pull request if you find a bug or have a feature request.

## License

This project is licensed under the **MIT License**.
