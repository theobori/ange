package webring

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
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

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(
		`CREATE TABLE IF NOT EXISTS gopherspaces (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			domain TEXT NOT NULL,
			port INTEGER NOT NULL,
			path TEXT NOT NULL,
			token TEXT NOT NULL,
			approved BOOLEAN DEFAULT 0,
			UNIQUE(domain, port, path),
			UNIQUE(token)
		);
			
		CREATE TABLE IF NOT EXISTS webring (
			gopherspace_id INTEGER PRIMARY KEY,
			previous INTEGER NOT NULL DEFAULT -1,
			next INTEGER NOT NULL DEFAULT -1,
			FOREIGN KEY (gopherspace_id) REFERENCES gopherspaces(id) ON DELETE CASCADE
		)`,
	)

	if err != nil {
		return nil, err
	}

	return &Database{
		db: db,
	}, nil
}

func (d *Database) AddGopherspace(gopherspace *Gopherspace) (int, error) {
	result, err := d.db.Exec(
		`INSERT INTO gopherspaces(title, domain, port, path, token) VALUES(?, ?, ?, ?, ?)`,
		gopherspace.Title,
		gopherspace.Domain,
		gopherspace.Port,
		gopherspace.Path,
		gopherspace.Token,
	)
	if err != nil {
		return -1, err
	}

	id64, err := result.LastInsertId()
	if err != nil {
		return -1, err
	}

	id := int(id64)

	return id, nil
}

func (d *Database) addGopherspaceToWebring(gopherspaceId int, headId int, tailId int) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(
		`INSERT INTO webring(gopherspace_id, previous, next) VALUES(?, ?, ?)`,
		gopherspaceId,
		tailId,
		headId,
	)
	if err != nil {
		tx.Rollback()
		return err
	}
	if gopherspaceId != headId && gopherspaceId != tailId {
		_, err = tx.Exec(`UPDATE webring SET previous = ? WHERE gopherspace_id = ?`, gopherspaceId, headId)
		if err != nil {
			tx.Rollback()
			return err
		}
		_, err = tx.Exec(`UPDATE webring SET next = ? WHERE gopherspace_id = ?`, gopherspaceId, tailId)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (d *Database) AddGopherspaceToWebring(gopherspaceId int) error {
	var (
		err    error
		headId int
		tailId int
	)

	err = d.db.QueryRow("SELECT gopherspace_id, previous FROM webring LIMIT 1").Scan(&headId, &tailId)

	switch err {
	case sql.ErrNoRows:
		err := d.addGopherspaceToWebring(gopherspaceId, gopherspaceId, gopherspaceId)
		if err != nil {
			return err
		}

		return nil
	case nil:
	default:
		return err
	}

	err = d.addGopherspaceToWebring(gopherspaceId, headId, tailId)
	if err != nil {
		return err
	}

	return nil
}

func (d *Database) deleteFromWebring(id int, previous int, next int) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(`DELETE FROM gopherspaces WHERE id = ?`, id)
	if err != nil {
		tx.Rollback()
		return err
	}
	if previous != id && next != id {
		_, err = tx.Exec(`UPDATE webring SET next = ? WHERE gopherspace_id = ?`, next, previous)
		if err != nil {
			tx.Rollback()
			return err
		}
		_, err = tx.Exec(`UPDATE webring SET previous = ? WHERE gopherspace_id = ?`, previous, next)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (d *Database) DeleteByIdFromWebring(id int) error {
	var (
		err      error
		previous int
		next     int
	)

	err = d.db.QueryRow(
		`SELECT gopherspace_id, previous, next FROM webring WHERE gopherspace_id = ?`, id,
	).Scan(&id, &previous, &next)
	if err != nil {
		return err
	}

	return d.deleteFromWebring(id, previous, next)
}

func (d *Database) DeleteByTokenFromWebring(token string) error {
	var (
		err      error
		previous int
		next     int
		id       int
	)

	err = d.db.QueryRow(
		`SELECT gopherspace_id, previous, next
		FROM webring
		INNER JOIN gopherspaces on gopherspaces.id = webring.gopherspace_id
		WHERE gopherspaces.token = ?`,
		token,
	).Scan(&id, &previous, &next)
	if err != nil {
		return err
	}

	return d.deleteFromWebring(id, previous, next)
}

func (d *Database) GopherspacesInReview() ([]Gopherspace, error) {
	rows, err := d.db.Query(`SELECT id, title, domain, port, path FROM gopherspaces WHERE approved = 0`)
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

	gopherspaces := []Gopherspace{}

	for rows.Next() {
		err = rows.Scan(&id, &title, &domain, &port, &path)
		if err != nil {
			return nil, err
		}

		gopherspace := Gopherspace{
			Id:     id,
			Title:  title,
			Domain: domain,
			Port:   port,
			Path:   path,
		}

		gopherspaces = append(gopherspaces, gopherspace)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return gopherspaces, nil
}

func (d *Database) GopherspacesInWebring() ([]Gopherspace, error) {
	rows, err := d.db.Query(
		`SELECT id, title, domain, port, path, next
		FROM gopherspaces
		INNER JOIN webring on gopherspaces.id = webring.gopherspace_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var (
		id     int
		headId int
		title  string
		domain string
		port   int
		path   string
		next   int
	)

	graph := map[int]*Node{}

	if rows.Next() {
		err = rows.Scan(&headId, &title, &domain, &port, &path, &next)
		if err != nil {
			return nil, err
		}

		graph[headId] = &Node{
			Gopherspace: Gopherspace{
				Id:     headId,
				Title:  title,
				Domain: domain,
				Port:   port,
				Path:   path,
			},
			Next: next,
		}
	}

	for rows.Next() {
		err = rows.Scan(&id, &title, &domain, &port, &path, &next)
		if err != nil {
			return nil, err
		}

		graph[id] = &Node{
			Gopherspace: Gopherspace{
				Id:     id,
				Title:  title,
				Domain: domain,
				Port:   port,
				Path:   path,
			},
			Next: next,
		}
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	gopherspaces := []Gopherspace{}

	curr := headId
	for {
		node, ok := graph[curr]
		if !ok {
			break
		}

		gopherspaces = append(gopherspaces, node.Gopherspace)
		prev := curr
		curr = node.Next
		delete(graph, prev)
	}

	return gopherspaces, nil
}

func (d *Database) Gopherspaces(approved bool) ([]Gopherspace, error) {
	if !approved {
		return d.GopherspacesInReview()
	}

	return d.GopherspacesInWebring()
}

func (d *Database) SetApprove(id int, value bool) error {
	_, err := d.db.Exec(`UPDATE gopherspaces SET approved = ? WHERE id = ?`, value, id)
	if err != nil {
		return err
	}

	return nil
}

func (d *Database) Approve(id int) error {
	return d.SetApprove(id, true)
}

func (d *Database) Deny(id int) error {
	return d.SetApprove(id, false)
}

func (d *Database) RandomGopherspace() (*Gopherspace, error) {
	var (
		title  string
		domain string
		port   int
		path   string
	)

	err := d.db.QueryRow(
		`SELECT title, domain, port, path FROM gopherspaces ORDER BY RANDOM() LIMIT 1`,
	).Scan(&title, &domain, &port, &path)
	if err != nil {
		return nil, err
	}

	return &Gopherspace{
		Title:  title,
		Domain: domain,
		Port:   port,
		Path:   path,
	}, nil
}

func (d *Database) NeighborGopherspace(id int, isNext bool) (*Gopherspace, error) {
	var (
		title      string
		domain     string
		port       int
		path       string
		columnName string
	)

	if isNext {
		columnName = "webring.previous"
	} else {
		columnName = "webring.next"
	}

	stmtSql := fmt.Sprintf(
		`SELECT title, domain, port, path
		FROM gopherspaces
		INNER JOIN webring on gopherspaces.id = webring.gopherspace_id
		WHERE %s = ?`,
		columnName,
	)
	err := d.db.QueryRow(stmtSql, id).Scan(&title, &domain, &port, &path)
	if err != nil {
		return nil, err
	}

	return &Gopherspace{
		Title:  title,
		Domain: domain,
		Port:   port,
		Path:   path,
	}, nil
}

func (d *Database) NextGopherspace(id int) (*Gopherspace, error) {
	return d.NeighborGopherspace(id, true)
}

func (d *Database) PreviousGopherspace(id int) (*Gopherspace, error) {
	return d.NeighborGopherspace(id, false)
}
