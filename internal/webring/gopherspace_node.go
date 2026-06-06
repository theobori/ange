package webring

type GopherspaceNode struct {
	Id       int
	Entry    GopherspaceEntry
	Token    string
	Approved bool
	Previous int
	Next     int
}
