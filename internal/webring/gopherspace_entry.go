package webring

var GopherspaceEntryKeywords = []string{
	"title",
	"domain",
	"port",
	"path",
}

type GopherspaceEntry struct {
	Title  string
	Domain string
	Port   int
	Path   string
}
