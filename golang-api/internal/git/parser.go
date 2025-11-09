package git

import (
	"regexp"
	"strings"
	"time"
)

type ParsedURI struct {
	User string
	Repo string
	Host string
}

func ParseGitURI(uri string) *ParsedURI {
	uri = strings.TrimSuffix(uri, ".git")

	httpsGithub := regexp.MustCompile(`^https?://github\.com/([^/]+)/([^/]+)`)
	if m := httpsGithub.FindStringSubmatch(uri); m != nil {
		return &ParsedURI{User: m[1], Repo: m[2], Host: "github.com"}
	}

	sshGithub := regexp.MustCompile(`^git@github\.com:([^/]+)/([^/]+)`)
	if m := sshGithub.FindStringSubmatch(uri); m != nil {
		return &ParsedURI{User: m[1], Repo: m[2], Host: "github.com"}
	}

	httpsOther := regexp.MustCompile(`^https?://([^/]+)/([^/]+)/([^/]+)`)
	if m := httpsOther.FindStringSubmatch(uri); m != nil {
		return &ParsedURI{User: m[2], Repo: m[3], Host: m[1]}
	}

	sshOther := regexp.MustCompile(`^git@([^:]+):([^/]+)/([^/]+)`)
	if m := sshOther.FindStringSubmatch(uri); m != nil {
		return &ParsedURI{User: m[2], Repo: m[3], Host: m[1]}
	}

	return nil
}

func GenerateCloneDirectoryName(gitURI, customName string) string {
	if customName != "" {
		return customName
	}

	parsed := ParseGitURI(gitURI)
	if parsed == nil {
		return ""
	}

	datePrefix := time.Now().Format("2006-01-02")
	return datePrefix + "-" + parsed.User + "-" + parsed.Repo
}

func IsGitURI(arg string) bool {
	if arg == "" {
		return false
	}
	return strings.HasPrefix(arg, "https://") ||
		strings.HasPrefix(arg, "http://") ||
		strings.HasPrefix(arg, "git@") ||
		strings.Contains(arg, "github.com") ||
		strings.Contains(arg, "gitlab.com") ||
		strings.HasSuffix(arg, ".git")
}
