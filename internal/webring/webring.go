package webring

import (
	"sync"

	"github.com/theobori/ange/internal/hash"
)

const DatabasePerm = 0600

type Webring struct {
	database   *Database
	secretHash *hash.Hash
	mu         sync.Mutex
}

func NewWebring(database *Database, secretHash *hash.Hash) (*Webring, error) {
	return &Webring{
		database:   database,
		secretHash: secretHash,
	}, nil
}

func (w *Webring) AddForReview(gopherspace *Gopherspace) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	id, err := w.database.AddGopherspace(gopherspace)
	if err != nil {
		return -1, err
	}

	return id, nil
}

func (w *Webring) Delete(token string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.database.DeleteByTokenFromWebring(token)
}

func (w *Webring) ListGopherspaces(approved bool) ([]Gopherspace, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.database.Gopherspaces(approved)
}

func (w *Webring) IsAdmin(secret string) bool {
	return w.secretHash.Verify(secret)
}

func (w *Webring) Approve(id int) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	err := w.database.AddGopherspaceToWebring(id)
	if err != nil {
		return err
	}

	return w.database.Approve(id)
}

func (w *Webring) Deny(id int) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.database.Deny(id)
}

func (w *Webring) RandomGopherspace() (*Gopherspace, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.database.RandomGopherspace()
}

func (w *Webring) NeighborGopherspace(id int, isNext bool) (*Gopherspace, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.database.NeighborGopherspace(id, isNext)
}
