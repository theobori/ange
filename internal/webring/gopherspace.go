package webring

import (
	"fmt"

	"github.com/theobori/ange/internal/random"
	"github.com/theobori/fleur/gophermap"
)

type Gopherspace struct {
	Id       int
	Title    string
	Domain   string
	Port     int
	Path     string
	Token    string
	Approved bool
}

func NewGopherspaceFromUserInput(title string, domain string, port int, path string) (*Gopherspace, error) {
	if port < 0 {
		return nil, fmt.Errorf("Port must be a positive integer.")
	}

	token := random.Generate(32)

	return &Gopherspace{
		Title:    title,
		Domain:   domain,
		Port:     port,
		Path:     path,
		Token:    token,
		Approved: false,
	}, nil
}

func (g *Gopherspace) URL() string {
	return fmt.Sprintf(
		"gopher://%s:%d/%s",
		g.Domain,
		g.Port,
		g.Path,
	)
}

func (g *Gopherspace) GophermapItem() *gophermap.Item {
	return &gophermap.Item{
		ItemType:    gophermap.ItemTypeGopherMenu,
		Description: g.Title,
		Selector:    g.Path,
		Domain:      g.Domain,
		Port:        g.Port,
	}
}
