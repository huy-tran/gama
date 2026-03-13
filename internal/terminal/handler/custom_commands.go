package handler

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	tea "github.com/charmbracelet/bubbletea"
)

// execCustomCommandMsg is sent from a tab-option callback (goroutine) back into the
// Bubble Tea Update loop so that tea.ExecProcess can be returned as a Cmd.
type execCustomCommandMsg struct {
	cmdTemplate string
	repoName    string
	repoPath    string
}

// customCommandFinishedMsg is delivered to the Update loop after the subprocess exits.
type customCommandFinishedMsg struct {
	err error
}

// repoPathLookup resolves the local filesystem path for repoName ("owner/repo")
// by looking it up in the repo_paths config map.
// Returns an empty string if the repo has no mapping.
// Supports ~ expansion in the stored path value.
func repoPathLookup(repoPaths map[string]string, repoName string) string {
	if len(repoPaths) == 0 {
		return ""
	}
	raw, ok := repoPaths[repoName]
	if !ok || raw == "" {
		return ""
	}
	if strings.HasPrefix(raw, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, raw[2:])
		}
	}
	return raw
}

// runCustomCommand expands the command template and suspends the TUI via
// tea.ExecProcess so the command runs attached to the real terminal.
// On Windows it uses "cmd /C"; everywhere else it uses "sh -c".
func runCustomCommand(cmdTemplate, repoName, repoPath string) tea.Cmd {
	tmpl, err := template.New("cmd").Parse(cmdTemplate)
	if err != nil {
		return func() tea.Msg {
			return customCommandFinishedMsg{err: fmt.Errorf("invalid command template: %w", err)}
		}
	}

	var buf bytes.Buffer
	data := struct {
		RepoName string
		RepoPath string
	}{RepoName: repoName, RepoPath: repoPath}

	if err := tmpl.Execute(&buf, data); err != nil {
		return func() tea.Msg {
			return customCommandFinishedMsg{err: fmt.Errorf("template expansion failed: %w", err)}
		}
	}

	expanded := buf.String()

	var execCmd *exec.Cmd
	if runtime.GOOS == "windows" {
		execCmd = exec.Command("cmd", "/C", expanded) // #nosec G204
	} else {
		execCmd = exec.Command("sh", "-c", expanded) // #nosec G204
	}

	return tea.ExecProcess(execCmd, func(err error) tea.Msg {
		return customCommandFinishedMsg{err: err}
	})
}
