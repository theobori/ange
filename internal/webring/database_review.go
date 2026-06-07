package webring

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
