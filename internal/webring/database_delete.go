package webring

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
