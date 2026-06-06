package webring

type Node struct {
	GopherspaceId int
	Gopherspace   Gopherspace
	Previous      int
	Next          int
}
