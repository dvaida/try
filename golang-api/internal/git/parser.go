package git

type ParsedURI struct {
	User string
	Repo string
	Host string
}

func ParseGitURI(uri string) *ParsedURI {
	panic("not implemented")
}

func GenerateCloneDirectoryName(gitURI, customName string) string {
	panic("not implemented")
}

func IsGitURI(arg string) bool {
	panic("not implemented")
}
