package webring

import (
	"sync"
)

const DatabasePerm = 0600

type Webring struct {
	database *Database
	secret   string
	mu       sync.Mutex
}

func NewWebring(database *Database, secret string) (*Webring, error) {
	return &Webring{
		database: database,
		secret:   secret,
	}, nil
}

func (w *Webring) Add(entry *GopherspaceEntry) (*GopherspaceNode, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	node, err := w.database.AddGopherspaceEntry(entry)
	if err != nil {
		return nil, err
	}

	return node, nil
}

func (w *Webring) Delete(token string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.database.DeleteByToken(token)
}

func (w *Webring) List() ([]*GopherspaceNode, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.database.ListAsGopherspaceNode()
}

func (w *Webring) IsAdmin(secret string) bool {
	return secret == w.secret
}
