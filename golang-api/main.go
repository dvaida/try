package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tobi/try/golang-api/internal/git"
	"github.com/tobi/try/golang-api/internal/selector"
	"github.com/tobi/try/golang-api/internal/shell"
)

const version = "0.1.0-golang"

const TRY_PATH_DEFAULT = "~/src/tries"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		printGlobalHelp()
		os.Exit(0)
	}

	if len(os.Args) == 1 {
		printGlobalHelp()
		os.Exit(2)
	}

	args := os.Args[1:]
	triesPath := extractOptionWithValue(&args, "--path")
	if triesPath == "" {
		if envPath := os.Getenv("TRY_PATH"); envPath != "" {
			triesPath = envPath
		} else {
			triesPath = expandPath(TRY_PATH_DEFAULT)
		}
	}
	triesPath = expandPath(triesPath)

	andType := extractOptionWithValue(&args, "--and-type")
	andExit := hasFlag(&args, "--and-exit")
	andKeysRaw := extractOptionWithValue(&args, "--and-keys")
	andConfirm := extractOptionWithValue(&args, "--and-confirm")

	var andKeys []string
	if andKeysRaw != "" {
		andKeys = parseTestKeys(andKeysRaw)
	}

	command := ""
	if len(args) > 0 {
		command = args[0]
		args = args[1:]
	}

	switch command {
	case "clone":
		tasks := cmdClone(args, triesPath)
		shell.EmitTasksScript(tasks)
		os.Exit(0)
	case "init":
		cmdInit(args, triesPath)
		os.Exit(0)
	case "worktree":
		tasks := cmdWorktree(args, triesPath)
		shell.EmitTasksScript(tasks)
		os.Exit(0)
	case "cd":
		tasks := cmdCd(args, triesPath, andType, andConfirm, andExit, andKeys)
		if tasks != nil {
			shell.EmitTasksScript(tasks)
		}
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printGlobalHelp()
		os.Exit(2)
	}
}

func printGlobalHelp() {
	tryPath := os.Getenv("TRY_PATH")
	if tryPath == "" {
		tryPath = TRY_PATH_DEFAULT
	}

	help := `try something!

Lightweight experiments for people with ADHD

this tool is not meant to be used directly,
but added to your ~/.zshrc or ~/.bashrc:

  eval "$(try init ~/src/tries)"

for fish shell, add to ~/.config/fish/config.fish:

  eval (try init ~/src/tries | string collect)

Usage:

  init [--path PATH]  # Initialize shell function for aliasing
  cd [QUERY] [name?]  # Interactive selector; Git URL shorthand supported
  clone <git-uri> [name]  # Clone git repo into date-prefixed directory
  worktree dir [name]  # Create date-prefixed dir; add worktree from CWD if git repo
  worktree <repo-path> [name]  # Same as above, but source repo is <repo-path>

Clone Examples:

  try clone https://github.com/tobi/try.git
  # Creates: 2025-08-27-tobi-try

  try clone https://github.com/tobi/try.git my-fork
  # Creates: my-fork

  try https://github.com/tobi/try.git
  # Shorthand for clone (same as first example)

Worktree Examples:

  try worktree dir
  # From current git repo, creates: 2025-08-27-repo-name and adds detached worktree

  try worktree ~/src/github.com/tobi/try my-branch
  # From given repo path, creates: 2025-08-27-my-branch and adds detached worktree

Defaults:
  Default path: ` + TRY_PATH_DEFAULT + ` (override with --path on commands)
  Current default: ` + tryPath + `
`
	fmt.Print(help)
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	absPath, _ := filepath.Abs(path)
	return absPath
}

func extractOptionWithValue(args *[]string, optName string) string {
	for i := len(*args) - 1; i >= 0; i-- {
		arg := (*args)[i]
		if arg == optName {
			if i+1 < len(*args) {
				value := (*args)[i+1]
				*args = append((*args)[:i], (*args)[i+2:]...)
				return value
			}
			*args = append((*args)[:i], (*args)[i+1:]...)
			return ""
		}
		if strings.HasPrefix(arg, optName+"=") {
			value := strings.TrimPrefix(arg, optName+"=")
			*args = append((*args)[:i], (*args)[i+1:]...)
			return value
		}
	}
	return ""
}

func hasFlag(args *[]string, flag string) bool {
	for i, arg := range *args {
		if arg == flag {
			*args = append((*args)[:i], (*args)[i+1:]...)
			return true
		}
	}
	return false
}

func cmdClone(args []string, triesPath string) []shell.Task {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: git URI required for clone command")
		fmt.Fprintln(os.Stderr, "Usage: try clone <git-uri> [name]")
		os.Exit(1)
	}

	gitURI := args[0]
	var customName string
	if len(args) > 1 {
		customName = args[1]
	}

	dirName := git.GenerateCloneDirectoryName(gitURI, customName)
	if dirName == "" {
		fmt.Fprintf(os.Stderr, "Error: Unable to parse git URI: %s\n", gitURI)
		os.Exit(1)
	}

	fullPath := filepath.Join(triesPath, dirName)
	return []shell.Task{
		{Type: "target", Path: fullPath},
		{Type: "mkdir"},
		{Type: "echo", Msg: fmt.Sprintf("Using git clone to create this trial from %s.", gitURI)},
		{Type: "git-clone", URI: gitURI},
		{Type: "touch"},
		{Type: "cd"},
	}
}

func cmdInit(args []string, triesPath string) {
	scriptPath, _ := filepath.Abs(os.Args[0])

	if len(args) > 0 && strings.HasPrefix(args[0], "/") {
		triesPath = filepath.Clean(args[0])
		args = args[1:]
	}

	pathArg := ""
	if triesPath != "" {
		pathArg = fmt.Sprintf(` --path "%s"`, triesPath)
	}

	if isFish() {
		fishScript := fmt.Sprintf(`function try
  set -l script_path "%s"
  # Check if first argument is a known command
  switch $argv[1]
    case clone worktree init
      set -l cmd (/usr/bin/env %s%s $argv 2>/dev/tty | string collect)
    case '*'
      set -l cmd (/usr/bin/env %s cd%s $argv 2>/dev/tty | string collect)
  end
  set -l rc $status
  if test $rc -eq 0
    if string match -r ' && ' -- $cmd
      eval $cmd
    else
      printf %%s $cmd
    end
  else
    printf %%s $cmd
  end
end
`, scriptPath, scriptPath, pathArg, scriptPath, pathArg)
		fmt.Print(fishScript)
	} else {
		bashScript := fmt.Sprintf(`try() {
  script_path='%s'
  # Check if first argument is a known command
  case "$1" in
    clone|worktree|init)
      cmd=$(/usr/bin/env "$script_path"%s "$@" 2>/dev/tty)
      ;;
    *)
      cmd=$(/usr/bin/env "$script_path" cd%s "$@" 2>/dev/tty)
      ;;
  esac
  rc=$?
  if [ $rc -eq 0 ]; then
    case "$cmd" in
      *" && "*) eval "$cmd" ;;
      *) printf %%s "$cmd" ;;
    esac
  else
    printf %%s "$cmd"
  fi
}
`, scriptPath, pathArg, pathArg)
		fmt.Print(bashScript)
	}
}

func cmdWorktree(args []string, triesPath string) []shell.Task {
	sub := ""
	if len(args) > 0 {
		sub = args[0]
		args = args[1:]
	}

	var base, repoDir string
	cwd, _ := os.Getwd()

	if sub == "" || sub == "dir" {
		custom := strings.Join(args, " ")
		if custom != "" {
			base = strings.ReplaceAll(custom, " ", "-")
		} else {
			base = filepath.Base(cwd)
		}
		repoDir = cwd
	} else {
		repoDir = filepath.Clean(sub)
		custom := strings.Join(args, " ")
		if custom != "" {
			base = strings.ReplaceAll(custom, " ", "-")
		} else {
			base = filepath.Base(repoDir)
		}
	}

	datePrefix := time.Now().Format("2006-01-02")
	base = shell.ResolveUniqueNameWithVersioning(triesPath, datePrefix, base)
	dirName := datePrefix + "-" + base
	fullPath := filepath.Join(triesPath, dirName)

	tasks := []shell.Task{
		{Type: "target", Path: fullPath},
		{Type: "mkdir"},
	}

	gitDir := filepath.Join(repoDir, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		tasks = append(tasks, shell.Task{Type: "echo", Msg: fmt.Sprintf("Using git worktree to create this trial from %s.", repoDir)})
		if sub == "" || sub == "dir" {
			tasks = append(tasks, shell.Task{Type: "git-worktree"})
		} else {
			tasks = append(tasks, shell.Task{Type: "git-worktree", Repo: repoDir})
		}
	}

	tasks = append(tasks, shell.Task{Type: "touch"})
	tasks = append(tasks, shell.Task{Type: "cd"})
	return tasks
}

func cmdCd(args []string, triesPath string, andType, andConfirm string, andExit bool, andKeys []string) []shell.Task {
	if len(args) > 0 && args[0] == "clone" {
		return cmdClone(args[1:], triesPath)
	}

	if len(args) > 0 && (args[0] == "." || args[0] == "./") {
		pathArg := args[0]
		custom := ""
		if len(args) > 1 {
			custom = strings.Join(args[1:], " ")
		}

		cwd, _ := os.Getwd()
		repoDir := filepath.Clean(filepath.Join(cwd, pathArg))

		base := ""
		if custom != "" {
			base = strings.ReplaceAll(custom, " ", "-")
		} else {
			base = filepath.Base(repoDir)
		}

		datePrefix := time.Now().Format("2006-01-02")
		base = shell.ResolveUniqueNameWithVersioning(triesPath, datePrefix, base)
		dirName := datePrefix + "-" + base
		fullPath := filepath.Join(triesPath, dirName)

		tasks := []shell.Task{
			{Type: "target", Path: fullPath},
			{Type: "mkdir"},
		}

		gitDir := filepath.Join(repoDir, ".git")
		if _, err := os.Stat(gitDir); err == nil {
			tasks = append(tasks, shell.Task{Type: "echo", Msg: fmt.Sprintf("Using git worktree to create this trial from %s.", repoDir)})
			tasks = append(tasks, shell.Task{Type: "git-worktree", Repo: repoDir})
		}

		tasks = append(tasks, shell.Task{Type: "touch"})
		tasks = append(tasks, shell.Task{Type: "cd"})
		return tasks
	}

	if len(args) > 0 && git.IsGitURI(args[0]) {
		gitURI := args[0]
		var customName string
		if len(args) > 1 {
			customName = strings.Join(args[1:], " ")
		}

		dirName := git.GenerateCloneDirectoryName(gitURI, customName)
		if dirName == "" {
			fmt.Fprintf(os.Stderr, "Error: Unable to parse git URI: %s\n", gitURI)
			os.Exit(1)
		}

		fullPath := filepath.Join(triesPath, dirName)
		return []shell.Task{
			{Type: "target", Path: fullPath},
			{Type: "mkdir"},
			{Type: "echo", Msg: fmt.Sprintf("Using git clone to create this trial from %s.", gitURI)},
			{Type: "git-clone", URI: gitURI},
			{Type: "touch"},
			{Type: "cd"},
		}
	}

	searchTerm := strings.Join(args, " ")
	options := map[string]interface{}{
		"test_render_once": andExit,
		"test_no_cls":      andExit || len(andKeys) > 0,
		"test_keys":        andKeys,
		"test_confirm":     andConfirm,
	}
	if andType != "" {
		options["initial_input"] = andType
	}

	selector := selector.NewTrySelector(searchTerm, triesPath, options)

	if andExit {
		selector.Run()
		return nil
	}

	result := selector.Run()
	if result == nil {
		return nil
	}

	// Check if user cancelled (e.g., pressed Enter on "Create new" without typing)
	if result["type"] == "cancel" {
		return nil
	}

	tasks := []shell.Task{{Type: "target", Path: result["path"].(string)}}

	if result["type"] == "mkdir" {
		tasks = append(tasks, shell.Task{Type: "mkdir"})
	}

	tasks = append(tasks, shell.Task{Type: "touch"})
	tasks = append(tasks, shell.Task{Type: "cd"})

	return tasks
}

func parseTestKeys(spec string) []string {
	if spec == "" {
		return nil
	}

	tokens := strings.Split(spec, ",")
	var keys []string

	for _, tok := range tokens {
		tok = strings.TrimSpace(tok)
		tokUpper := strings.ToUpper(tok)

		switch tokUpper {
		case "UP":
			keys = append(keys, "\x1b[A")
		case "DOWN":
			keys = append(keys, "\x1b[B")
		case "LEFT":
			keys = append(keys, "\x1b[D")
		case "RIGHT":
			keys = append(keys, "\x1b[C")
		case "ENTER":
			keys = append(keys, "\r")
		case "ESC":
			keys = append(keys, "\x1b")
		case "BACKSPACE":
			keys = append(keys, "\x7F")
		case "CTRL-D", "CTRLD":
			keys = append(keys, "\x04")
		case "CTRL-P", "CTRLP":
			keys = append(keys, "\x10")
		case "CTRL-N", "CTRLN":
			keys = append(keys, "\x0E")
		case "CTRL-J", "CTRLJ":
			keys = append(keys, "\n")
		case "CTRL-K", "CTRLK":
			keys = append(keys, "\x0B")
		default:
			if strings.HasPrefix(tokUpper, "TYPE=") {
				text := tok[5:]
				for _, ch := range text {
					keys = append(keys, string(ch))
				}
			} else if len(tok) == 1 {
				keys = append(keys, tok)
			}
		}
	}

	return keys
}

func isFish() bool {
	shell := os.Getenv("SHELL")
	return strings.Contains(shell, "fish")
}
