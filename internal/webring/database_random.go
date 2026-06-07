package webring

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
