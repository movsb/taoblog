package service

/*
func (s *ImplServer) ListTagsWithCount(limit int64, mergeAlias bool) []*models.TagWithCount {
	query, args := sql_helpers.NewSelect().
		From("post_tags", "pt").From("tags", "t").
		Select("t.*,COUNT(pt.id) size").
		Where("pt.tag_id=t.id").
		GroupBy("t.id").
		OrderBy("size DESC").
		Limit(limit).
		SQL()
	rows, err := s.db.Query(query, args...)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	rootTags := make([]*models.TagWithCount, 0)
	aliasTags := make([]*models.TagWithCount, 0)
	rootMap := make(map[int64]*models.TagWithCount, 0)

	for rows.Next() {
		var tag models.TagWithCount
		err = rows.Scan(&tag.ID, &tag.Name, &tag.Alias, &tag.Count)
		if err != nil {
			panic(err)
		}
		if !mergeAlias {
			rootTags = append(rootTags, &tag)
		} else {
			if tag.Alias == 0 {
				rootTags = append(rootTags, &tag)
				rootMap[tag.ID] = &tag
			} else {
				aliasTags = append(aliasTags, &tag)
			}
		}
	}

	if mergeAlias {
		for _, tag := range aliasTags {
			if root, ok := rootMap[tag.Alias]; ok {
				root.Count += tag.Count
			}
		}
	}

	return rootTags
}
*/

func (s *ImplServer) getObjectTagIDs(postID int64, alias bool) (ids []int64) {
	sql := `SELECT tag_id FROM post_tags WHERE post_id=?`
	rows, err := s.db.Query(sql, postID)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		err = rows.Scan(&id)
		if err != nil {
			panic(err)
		}
		ids = append(ids, id)
	}
	if alias {
		ids = s.getAliasTagsAll(ids)
	}

	return
}

func (s *ImplServer) GetObjectTagNames(postID int64) []string {
	query := `select tags.name from post_tags,tags where post_tags.post_id=? and post_tags.tag_id=tags.id`
	args := []interface{}{postID}
	rows, err := s.db.Query(query, args...)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	names := make([]string, 0)
	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			panic(err)
		}
		names = append(names, name)
	}
	return names
}

func (s *ImplServer) getAliasTagsAll(ids []int64) []int64 {
	sids := joinInts(ids, ",")

	sql1 := `SELECT alias FROM tags WHERE id in (?)`
	sql2 := `SELECT id FROM tags WHERE alias in (?)`

	rows, err := s.db.Query(sql1, sids)
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var alias int64
		if err = rows.Scan(&alias); err != nil {
			panic(err)
		}

		if alias > 0 {
			ids = append(ids, alias)
		}
	}

	rows.Close()

	rows, err = s.db.Query(sql2, sids)
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var id int64
		if err = rows.Scan(&id); err != nil {
			panic(err)
		}

		ids = append(ids, id)
	}

	rows.Close()

	return ids
}
