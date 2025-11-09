package selector

type TrySelector struct {
	SearchTerm      string
	CursorPos       int
	ScrollOffset    int
	InputBuffer     string
	Selected        map[string]interface{}
	AllTries        []map[string]interface{}
	BasePath        string
	DeleteStatus    string
	TestRenderOnce  bool
	TestNoCls       bool
	TestKeys        []string
	TestConfirm     string
}

func NewTrySelector(searchTerm, basePath string, options map[string]interface{}) *TrySelector {
	panic("not implemented")
}

func (ts *TrySelector) Run() map[string]interface{} {
	panic("not implemented")
}

func (ts *TrySelector) LoadAllTries() []map[string]interface{} {
	panic("not implemented")
}

func (ts *TrySelector) GetTries() []map[string]interface{} {
	panic("not implemented")
}

func (ts *TrySelector) CalculateScore(text, query string, ctime, mtime int64) float64 {
	panic("not implemented")
}

func (ts *TrySelector) FormatRelativeTime(t int64) string {
	panic("not implemented")
}

func (ts *TrySelector) HighlightMatches(text, query string) string {
	panic("not implemented")
}

func (ts *TrySelector) TruncateWithANSI(text string, maxLength int) string {
	panic("not implemented")
}
