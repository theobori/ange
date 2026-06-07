package webring

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
