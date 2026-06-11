// Package editor provides functionality to open source files
// in the user's preferred text editor at a specific line number.
package editor

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// editorConfig holds the command template for opening a file at a line.
type editorConfig struct {
	name string
	// args is a function that returns the command args for opening file:line.
	args func(file string, line int) []string
}

// knownEditors maps editor binary names to their file:line argument format.
var knownEditors = []editorConfig{
	{
		name: "code", // VS Code
		args: func(file string, line int) []string {
			return []string{"--goto", fmt.Sprintf("%s:%d", file, line)}
		},
	},
	{
		name: "cursor", // Cursor editor
		args: func(file string, line int) []string {
			return []string{"--goto", fmt.Sprintf("%s:%d", file, line)}
		},
	},
	{
		name: "goland", // GoLand
		args: func(file string, line int) []string {
			return []string{"--line", fmt.Sprintf("%d", line), file}
		},
	},
	{
		name: "idea", // IntelliJ IDEA
		args: func(file string, line int) []string {
			return []string{"--line", fmt.Sprintf("%d", line), file}
		},
	},
	{
		name: "subl", // Sublime Text
		args: func(file string, line int) []string {
			return []string{fmt.Sprintf("%s:%d", file, line)}
		},
	},
	{
		name: "vim",
		args: func(file string, line int) []string {
			return []string{fmt.Sprintf("+%d", line), file}
		},
	},
	{
		name: "nvim", // Neovim
		args: func(file string, line int) []string {
			return []string{fmt.Sprintf("+%d", line), file}
		},
	},
	{
		name: "nano",
		args: func(file string, line int) []string {
			return []string{fmt.Sprintf("+%d", line), file}
		},
	},
	{
		name: "emacs",
		args: func(file string, line int) []string {
			return []string{fmt.Sprintf("+%d", line), file}
		},
	},
}

// OpenInEditor opens the given file at the specified line number in the
// user's preferred editor.
//
// Editor detection order:
//  1. $EDITOR environment variable
//  2. $VISUAL environment variable
//  3. Auto-detect from known editors on PATH
//
// Returns an error if no editor could be found or the command fails to start.
func OpenInEditor(file string, line int) error {
	editorCmd, args := DetectEditor(file, line)
	if editorCmd == "" {
		return fmt.Errorf("no editor found: set $EDITOR or $VISUAL environment variable")
	}

	cmd := exec.Command(editorCmd, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Start()
}

// DetectEditor determines which editor to use and returns the command
// name and arguments for opening the file at the given line.
func DetectEditor(file string, line int) (string, []string) {
	// 1. Check $EDITOR
	if editor := os.Getenv("EDITOR"); editor != "" {
		return resolveEditor(editor, file, line)
	}

	// 2. Check $VISUAL
	if editor := os.Getenv("VISUAL"); editor != "" {
		return resolveEditor(editor, file, line)
	}

	// 3. Auto-detect from known editors on PATH
	for _, ec := range knownEditors {
		if path, err := exec.LookPath(ec.name); err == nil {
			return path, ec.args(file, line)
		}
	}

	return "", nil
}

// resolveEditor takes an editor name (possibly from $EDITOR) and returns
// the command and appropriate arguments for opening file:line.
func resolveEditor(editor string, file string, line int) (string, []string) {
	// The editor env var might contain flags (e.g. "code --wait")
	parts := strings.Fields(editor)
	if len(parts) == 0 {
		return "", nil
	}

	editorName := parts[0]
	baseName := basename(editorName)

	// Check if it's a known editor so we can use the right file:line syntax
	for _, ec := range knownEditors {
		if baseName == ec.name {
			args := append(parts[1:], ec.args(file, line)...)
			return editorName, args
		}
	}

	// Unknown editor — use generic "editor file +line" approach
	// Most editors accept the filename as the last argument
	args := append(parts[1:], fmt.Sprintf("+%d", line), file)
	return editorName, args
}

// basename extracts the base name from a path (e.g. "/usr/bin/vim" → "vim").
func basename(path string) string {
	if idx := strings.LastIndex(path, "/"); idx >= 0 {
		return path[idx+1:]
	}
	return path
}

// EditorName returns the name of the editor that would be used,
// for display purposes in the UI.
func EditorName() string {
	if editor := os.Getenv("EDITOR"); editor != "" {
		parts := strings.Fields(editor)
		if len(parts) > 0 {
			return basename(parts[0])
		}
	}
	if editor := os.Getenv("VISUAL"); editor != "" {
		parts := strings.Fields(editor)
		if len(parts) > 0 {
			return basename(parts[0])
		}
	}
	for _, ec := range knownEditors {
		if _, err := exec.LookPath(ec.name); err == nil {
			return ec.name
		}
	}
	return "none"
}
