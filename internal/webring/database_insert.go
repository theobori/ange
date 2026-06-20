package webring

import (
	"database/sql"

	"github.com/theobori/ange/internal/hash"
)

func (d *Database) AddGopherspace(gopherspace *Gopherspace) (int, error) {
	tokenHash := hash.Generate(gopherspace.Token)

	result, err := d.db.Exec(
		`INSERT INTO gopherspaces(title, domain, port, path, token) VALUES(?, ?, ?, ?, ?)`,
		gopherspace.Title,
		gopherspace.Domain,
		gopherspace.Port,
		gopherspace.Path,
		tokenHash,
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
