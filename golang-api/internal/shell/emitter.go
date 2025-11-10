package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/tobi/try/golang-api/internal/ui"
)

type Task struct {
	Type string
	Path string
	URI  string
	Repo string
	Msg  string
}

func EmitTasksScript(tasks []Task) {
	var targetPath string
	for _, t := range tasks {
		if t.Type == "target" {
			targetPath = t.Path
			break
		}
	}

	if targetPath == "" {
		panic("emit_tasks_script requires a target path")
	}

	parts := []string{}
	quotedPath := shellQuote(targetPath)

	for _, t := range tasks {
		switch t.Type {
		case "echo":
			if t.Msg != "" {
				expanded := ui.ExpandTokens(t.Msg)
				quotedMsg := shellQuote(expanded)
				parts = append(parts, fmt.Sprintf("echo %s", quotedMsg))
			}
		case "mkdir":
			parts = append(parts, fmt.Sprintf("mkdir -p %s", quotedPath))
		case "git-clone":
			parts = append(parts, fmt.Sprintf("git clone '%s' %s", t.URI, quotedPath))
		case "git-worktree":
			if t.Repo != "" {
				quotedRepo := shellQuote(t.Repo)
				parts = append(parts, fmt.Sprintf(
					"/usr/bin/env sh -c 'if git -C %s rev-parse --is-inside-work-tree >/dev/null 2>&1; then repo=$(git -C %s rev-parse --show-toplevel); git -C \"$repo\" worktree add --detach %s >/dev/null 2>&1 || true; fi; exit 0'",
					quotedRepo, quotedRepo, quotedPath))
			} else {
				parts = append(parts, fmt.Sprintf(
					"/usr/bin/env sh -c 'if git rev-parse --is-inside-work-tree >/dev/null 2>&1; then repo=$(git rev-parse --show-toplevel); git -C \"$repo\" worktree add --detach %s >/dev/null 2>&1 || true; fi; exit 0'",
					quotedPath))
			}
		case "touch":
			parts = append(parts, fmt.Sprintf("touch %s", quotedPath))
		case "cd":
			parts = append(parts, fmt.Sprintf("cd %s", quotedPath))
		}
	}

	fmt.Print(JoinCommands(parts))
}

func JoinCommands(parts []string) string {
	return strings.Join(parts, " \\\n  && ")
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
}

func UniqueDirName(basePath, dirName string) string {
	candidate := dirName
	i := 2
	for {
		fullPath := filepath.Join(basePath, candidate)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			return candidate
		}
		candidate = fmt.Sprintf("%s-%d", dirName, i)
		i++
	}
}

func ResolveUniqueNameWithVersioning(basePath, datePrefix, base string) string {
	initial := datePrefix + "-" + base
	initialPath := filepath.Join(basePath, initial)

	if _, err := os.Stat(initialPath); os.IsNotExist(err) {
		return base
	}

	if matches := regexp.MustCompile(`^(.*?)(\d+)$`).FindStringSubmatch(base); matches != nil {
		stem := matches[1]
		num, _ := strconv.Atoi(matches[2])
		candidateNum := num + 1

		for {
			candidateBase := fmt.Sprintf("%s%d", stem, candidateNum)
			candidateFull := filepath.Join(basePath, datePrefix+"-"+candidateBase)
			if _, err := os.Stat(candidateFull); os.IsNotExist(err) {
				return candidateBase
			}
			candidateNum++
		}
	}

	fullName := datePrefix + "-" + base
	unique := UniqueDirName(basePath, fullName)
	return strings.TrimPrefix(unique, datePrefix+"-")
}
