package main

import (
	"database/sql"
	"fmt"

	"github.com/movsb/taoblog/modules/taorm"

	"github.com/movsb/taoblog/modules/datetime"
)

// PostForArchiveQuery is an archive query result.
type PostForArchiveQuery struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

type PostForRelated struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Relevance uint   `json:"relevance"`
}

type PostForDate struct {
	Year  int `json:"year"`
	Month int `json:"month"`
	Count int `json:"count"`
}

// PostForManagement for post management.
type PostForManagement struct {
	ID           int64  `json:"id"`
	Date         string `json:"date"`
	Modified     string `json:"modified"`
	Title        string `json:"title"`
	PageView     uint   `json:"page_view"`
	SourceType   string `json:"source_type"`
	CommentCount uint   `json:"comment_count" taorm:"name:comments"`
}

func (p *PostForManagement) ToLocalTime() {
	p.Date = datetime.My2Local(p.Date)
	p.Modified = datetime.My2Local(p.Modified)
}

// PostManager manages posts.
type PostManager struct {
}

// NewPostManager news a post manager.
func NewPostManager() *PostManager {
	return &PostManager{}
}

// Has returns true if post id exists.
func (z *PostManager) Has(tx Querier, id int64) (bool, error) {
	query := `SELECT id FROM posts WHERE id=` + fmt.Sprint(id)
	rows := tx.QueryRow(query)
	pid := 0
	err := rows.Scan(&pid)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return pid > 0, err
}

func (z *PostManager) getRowPosts(tx Querier, query string, args ...interface{}) ([]*PostForArchiveQuery, error) {
	var ps []*PostForArchiveQuery
	if err := taorm.QueryRows(&ps, tx, query, args...); err != nil {
		return nil, err
	}
	return ps, nil
}

// GetPostsByCategory gets category posts.
func (z *PostManager) GetPostsByCategory(tx Querier, catID int64) ([]*PostForArchiveQuery, error) {
	query := `select id,title from posts where taxonomy=? and type='post' order by date desc`
	args := []interface{}{catID}
	return z.getRowPosts(tx, query, args...)
}

/*
// GetPostsByDate gets date posts.
func (z *PostManager) GetPostsByDate(tx Querier, yy, mm int64) ([]*PostForArchiveQuery, error) {
	seldb := sql_helpers.NewSelect().From("posts", "").
		Select("id,title").
		Where("type='post'")
	if yy > 1970 {
		var start, end string
		if 1 <= mm && mm <= 12 {
			start, end = datetime.MonthStartEnd(int(yy), int(mm))
		} else {
			start, end = datetime.YearStartEnd(int(yy))
		}
		seldb.Where("date>=? AND date<=?", start, end)
	}
	seldb.OrderBy("date DESC")
	query, args := seldb.SQL()
	return z.getRowPosts(tx, query, args...)
}
*/

/*
// ListAllPosts lists
func (z *PostManager) ListAllPosts(tx Querier) ([]*PostForArchiveQuery, error) {
	query, args := sql_helpers.NewSelect().From("posts", "").
		Select("id,title").
		Where("type='post'").
		OrderBy("date DESC").
		SQL()
	return z.getRowPosts(tx, query, args...)
}

/*
// GetVars gets custom column values.
func (z *PostManager) GetVars(tx Querier, fields string, wheres string, outs ...interface{}) error {
	query, args := sql_helpers.NewSelect().From("posts", "").
		Select(fields).Where(wheres).Limit(1).SQL()
	row := tx.QueryRow(query, args...)
	return row.Scan(outs...)
}
*/

/*
func (z *PostManager) GetPostsForManagement(tx Querier) ([]*PostForManagement, error) {
	query, args := sql_helpers.NewSelect().From("posts", "").
		Select("id,date,modified,title,page_view,source_type,comments").
		Where("type='post'").OrderBy("id DESC").SQL()
	var posts []*PostForManagement
	if err := taorm.QueryRows(&posts, tx, query, args...); err != nil {
		return nil, err
	}
	for _, p := range posts {
		p.ToLocalTime()
	}
	return posts, nil
}

// GetCountOfType gets
func (z *PostManager) GetCountOfType(tx Querier, typ string) (int64, error) {
	query, args := sql_helpers.NewSelect().From("posts", "").
		Select("count(*) size").
		Where("type=?", typ).
		SQL()
	row := tx.QueryRow(query, args...)
	var count int64
	return count, row.Scan(&count)
}

// CreatePost creates a new post into database.
func (z *PostManager) CreatePost(tx Querier, post *Post) error {
	var err error
	if err = post.Create(tx); err != nil {
		return err
	}
	if err = tagmgr.UpdateObjectTags(tx, post.ID, post.Tags); err != nil {
		return err
	}

	lastTime := optmgr.GetDef(tx, "last_post_time", "")
	if lastTime == "" || lastTime < post.Date {
		optmgr.Set(tx, "last_post_time", post.Date)
	}

	if post.Type == "post" {
		count, err := z.GetCountOfType(tx, "post")
		if err != nil {
			return err
		}
		optmgr.Set(tx, "post_count", count)
	} else if post.Type == "page" {
		count, err := z.GetCountOfType(tx, "page")
		if err != nil {
			return err
		}
		optmgr.Set(tx, "page_count", count)
	}

	return nil
}

// UpdatePost updates a post.
func (z *PostManager) UpdatePost(tx Querier, post *Post) error {
	var err error

	if has, err := z.Has(tx, post.ID); true {
		if err != nil {
			fmt.Println(err)
			return err
		}
		if !has {
			return fmt.Errorf("没有这篇文章")
		}
	}

	if err = post.Update(tx); err != nil {
		return err
	}
	if err = tagmgr.UpdateObjectTags(tx, post.ID, post.Tags); err != nil {
		return err
	}

	return nil
}

func (z *PostManager) GetDateArchives(tx Querier) ([]*PostForDate, error) {
	query := "SELECT year,month,count(id) count FROM (SELECT id,date,year(date) year,month(date) month FROM(SELECT id,DATE_ADD(date,INTERVAL 8 HOUR) date FROM posts WHERE type='post') x) x GROUP BY year,month ORDER BY year DESC, month DESC"
	var ps []*PostForDate
	if err := taorm.QueryRows(&ps, tx, query); err != nil {
		return nil, err
	}
	return ps, nil
}

*/
