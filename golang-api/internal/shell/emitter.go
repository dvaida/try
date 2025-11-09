package shell

type Task struct {
	Type   string
	Path   string
	URI    string
	Repo   string
	Msg    string
}

func EmitTasksScript(tasks []Task) string {
	panic("not implemented")
}

func JoinCommands(parts []string) string {
	panic("not implemented")
}

func UniqueDirName(basePath, dirName string) string {
	panic("not implemented")
}

func ResolveUniqueNameWithVersioning(basePath, datePrefix, base string) string {
	panic("not implemented")
}
