package webring

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/theobori/ange/internal/common"
)

type Database struct {
	db *sql.DB
}

func NewDatabase(databasePath string) (*Database, error) {
	db, err := sql.Open("sqlite3", fmt.Sprintf("%s?cache=shared", databasePath))
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)

	sqlStmt := `CREATE TABLE IF NOT EXISTS gopherspaces (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		domain TEXT NOT NULL,
		port INTEGER NOT NULL,
		path TEXT NOT NULL,
		token TEXT NOT NULL,
		approved BOOLEAN DEFAULT 0,
		previous INTEGER NOT NULL DEFAULT -1,
		next INTEGER NOT NULL DEFAULT -1,
		UNIQUE(token),
		UNIQUE(domain, port, path)
	)`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	return &Database{
		db: db,
	}, nil
}

func (d *Database) addFirstGopherspace(entry *GopherspaceEntry, node *GopherspaceNode) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	result, err := tx.Exec(
		`INSERT INTO gopherspaces(title, domain, port, path, token) VALUES(?, ?, ?, ?, ?)`,
		entry.Title,
		entry.Domain,
		entry.Port,
		entry.Path,
		node.Token,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	lastInsertedId, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(`UPDATE gopherspaces SET previous = ?, next = ? WHERE id = ?`, &lastInsertedId, &lastInsertedId, &lastInsertedId)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}

	node.Id = int(lastInsertedId)
	node.Previous = node.Id
	node.Next = node.Id

	return nil
}

func (d *Database) addNonFirstGopherspace(entry *GopherspaceEntry, node *GopherspaceNode, headId int, tailId int) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	result, err := tx.Exec(
		`INSERT INTO gopherspaces(title, domain, port, path, token, previous, next) VALUES(?, ?, ?, ?, ?, ?, ?)`,
		entry.Title,
		entry.Domain,
		entry.Port,
		entry.Path,
		node.Token,
		tailId,
		headId,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	lastInsertedId64, err := result.LastInsertId()
	if err != nil {
		return err
	}
	lastInsertedId := int(lastInsertedId64)

	_, err = tx.Exec(`UPDATE gopherspaces SET previous = ? WHERE id = ?`, &lastInsertedId, &headId)
	if err != nil {
		tx.Rollback()
		return err
	}
	_, err = tx.Exec(`UPDATE gopherspaces SET next = ? WHERE id = ?`, &lastInsertedId, &tailId)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}

	node.Previous = tailId
	node.Next = headId
	node.Id = int(lastInsertedId)

	return nil
}

// expliquer chaine schame excalidraw
func (d *Database) AddGopherspaceEntry(entry *GopherspaceEntry) (*GopherspaceNode, error) {
	var (
		err    error
		headId int
		tailId int
	)

	node := GopherspaceNode{
		Entry: *entry,
		Token: common.GenerateToken(32),
	}

	err = d.db.QueryRow("SELECT id FROM gopherspaces ORDER BY id ASC LIMIT 1").Scan(&headId)

	switch err {
	case sql.ErrNoRows:
		err := d.addFirstGopherspace(entry, &node)
		if err != nil {
			return nil, err
		}

		return &node, nil
	case nil:
	default:
		return nil, err
	}

	err = d.db.QueryRow("SELECT max(id) FROM gopherspaces").Scan(&tailId)
	if err != nil {
		return nil, err
	}

	err = d.addNonFirstGopherspace(entry, &node, headId, tailId)
	if err != nil {
		return nil, err
	}

	return &node, nil
}

func (d *Database) DeleteByToken(token string) error {
	var (
		err      error
		previous int
		next     int
		id       int
	)

	err = d.db.QueryRow(
		`SELECT id, previous, next FROM gopherspaces WHERE token = ?`, &token,
	).Scan(&id, &previous, &next)
	if err != nil {
		return err
	}

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(`UPDATE gopherspaces SET next = ? WHERE next = ?`, &next, &id)
	if err != nil {
		tx.Rollback()
		return err
	}
	_, err = tx.Exec(`UPDATE gopherspaces SET previous = ? WHERE previous = ?`, &previous, &id)
	if err != nil {
		tx.Rollback()
		return err
	}
	_, err = tx.Exec(`DELETE FROM gopherspaces WHERE token = ?`, &token)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (d *Database) ListAsGopherspaceNode() ([]*GopherspaceNode, error) {
	rows, err := d.db.Query("SELECT id, title, domain, port, path FROM gopherspaces WHERE approved = 1")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var (
		id     int
		title  string
		domain string
		port   int
		path   string
	)

	nodes := []*GopherspaceNode{}
	for rows.Next() {
		err = rows.Scan(&id, &title, &domain, &port, &path)
		if err != nil {
			return nil, err
		}

		node := GopherspaceNode{
			Id: id,
			Entry: GopherspaceEntry{
				Title:  title,
				Domain: domain,
				Port:   port,
				Path:   path,
			},
		}
		nodes = append(nodes, &node)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return nodes, nil
}
