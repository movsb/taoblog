package service

import (
	"github.com/movsb/taoblog/modules/sql_helpers"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service/models"
)

func (s *ImplServer) ListTagsWithCount(in *protocols.ListTagsWithCountRequest) *protocols.ListTagsWithCountResponse {
	query, args := sql_helpers.NewSelect().
		From("post_tags", "pt").From("tags", "t").
		Select("t.*,COUNT(pt.id) size").
		Where("pt.tag_id=t.id").
		GroupBy("t.id").
		OrderBy("size DESC").
		Limit(in.Limit).
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
		if !in.MergeAlias {
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

	if in.MergeAlias {
		for _, tag := range aliasTags {
			if root, ok := rootMap[tag.Alias]; ok {
				root.Count += tag.Count
			}
		}
	}

	return &protocols.ListTagsWithCountResponse{
		Tags: models.TagWithCounts(rootTags).Serialize(),
	}
}
