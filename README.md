# Goroviz

A lightweight TUI tool that visualizes goroutine behavior from Go applications using pprof data. 

Just point it at your running Go app - Goroviz fetches the goroutine dump, groups identical stacks, and lets you explore them interactively.

## Walk Through 

<p align="center">
  <img
    src="https://github.com/user-attachments/assets/31d870c7-6ec9-4b44-965c-2227bdaad733"
    width="600"
  />
</p>

## Features

- **One-command attach** — `goroviz attach localhost:6060` and you're done
- **Smart Grouping** — Automatically groups identical goroutines by stack signature
- **Interactive TUI** — Navigate goroutine groups with keyboard shortcuts
- **Stack Trace Explorer** — Drill into any group to see full stack traces with file:line references
- **Editor Integration** — Jump to source code directly from the TUI (VS Code, GoLand, Vim, Neovim, and more)
- **Plain Text Mode** — Non-interactive output for piping and scripting

## Installation

```bash
go install github.com/santhosh-005/goroviz/cmd/goroviz@latest
```

## Usage

### 1. Enable pprof in your Go app (one-time setup)

```go
import _ "net/http/pprof"

func main() {
    go http.ListenAndServe("localhost:6060", nil)
    // ... your app code
}
```

### 2. Attach Goroviz to your running app

```bash
goroviz attach localhost:6060
```

That's it! Goroviz fetches the goroutine dump automatically and launches an interactive TUI.

### Other URL formats

```bash
goroviz attach http://myapp.example.com:8080
goroviz attach https://staging.myapp.com:6060
```

### Read from a file

```bash
goroviz dump goroutines.txt
```

### Plain text mode (no TUI)

```bash
goroviz attach localhost:6060 --text
goroviz dump goroutines.txt --text
```

## Keyboard Shortcuts

### List View

| Key | Action |
|-----|--------|
| `↑` / `k` | Move cursor up |
| `↓` / `j` | Move cursor down |
| `g` / `Home` | Jump to top |
| `b` / `End` | Jump to bottom |
| `Enter` | Open group details |
| `q` / `Ctrl+C` | Quit |

### Detail View

| Key | Action |
|-----|--------|
| `↑` / `k` | Previous stack frame |
| `↓` / `j` | Next stack frame |
| `e` / `o` | Open file in editor |
| `Esc` / `Backspace` | Back to list |
| `q` / `Ctrl+C` | Quit |

## Editor Support

Goroviz auto-detects your editor from `$EDITOR`, `$VISUAL`, or by searching your PATH:

| Editor | Detection |
|--------|-----------|
| VS Code | `code` |
| Cursor | `cursor` |
| GoLand | `goland` |
| IntelliJ | `idea` |
| Sublime Text | `subl` |
| Vim | `vim` |
| Neovim | `nvim` |
| Nano | `nano` |
| Emacs | `emacs` |

## Architecture

```
goroviz/
├── cmd/goroviz/      # CLI entry point (installable binary)
├── internal/
│   ├── parser/       # Goroutine dump parsing
│   ├── group/        # Grouping engine (pluggable strategies)
│   ├── fetcher/      # HTTP pprof endpoint fetcher
│   ├── tui/          # Bubble Tea interactive UI
│   └── editor/       # Editor integration
└── testdata/         # Sample goroutine dumps + demo app
```

## License

[MIT](LICENSE)
