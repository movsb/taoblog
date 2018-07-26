package main

type xAllPostResult struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

type xAllPostResults []xAllPostResult

func getAllPosts(tx Querier) (xAllPostResults, error) {
	var err error
	query := `SELECT id,title FROM posts WHERE type='post' AND status='public' ORDER BY date DESC`
	rows, err := tx.Query(query)
	if err != nil {
		return nil, err
	}

	//defer rows.Close()

	rets := make(xAllPostResults, 0)

	for rows.Next() {
		var ret xAllPostResult
		err = rows.Scan(&ret.ID, &ret.Title)
		if err != nil {
			return nil, err
		}
		rets = append(rets, ret)
	}

	return rets, nil
}
