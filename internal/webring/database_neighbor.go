package webring

import "fmt"

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
