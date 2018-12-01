package main

import (
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/movsb/taoblog/server/modules/post_translators"
	"github.com/movsb/taoblog/server/modules/sql_helpers"
	"github.com/movsb/taoblog/server/modules/utils/datetime"
)

// PostForArchiveQuery is an archive query result.
type PostForArchiveQuery struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

type PostForLatest struct {
	ID    int64         `json:"id"`
	Title template.HTML `json:"title"`
	Type  string        `json:"type"`
}

func (p *PostForLatest) Link() string {
	return fmt.Sprintf("/%d/", p.ID)
}

type PostForRelated struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
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
	CommentCount uint   `json:"comment_count"`
}

func (p *PostForManagement) Fields() string {
	cols := "id,date,modified,title,page_view,source_type,comments"
	return cols
}

func (p *PostForManagement) Pointers() []interface{} {
	ptrs := []interface{}{
		&p.ID,
		&p.Date, &p.Modified,
		&p.Title, &p.PageView,
		&p.SourceType, &p.CommentCount,
	}
	return ptrs
}

func (p *PostForManagement) ToLocalTime() {
	p.Date = datetime.My2Local(p.Date)
	p.Modified = datetime.My2Local(p.Modified)
}

// PostForRss for Rss Post.
type PostForRss struct {
	ID      int64
	Link    string
	Title   string
	Content template.HTML
	Date    string
}

// RssData for Rss template.
type RssData struct {
	BlogName    string
	Home        string
	Description string
	Posts       []*PostForRss
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

// internal use
func (z *PostManager) update(tx Querier, id int64, typ string, source string) error {
	var tr post_translators.PostTranslator
	var content string
	var err error

	switch typ {
	case "html":
		tr = &post_translators.HTMLTranslator{}
	case "markdown":
		tr = &post_translators.MarkdownTranslator{}
	}

	if tr == nil {
		return errors.New("no translator found for " + typ)
	}

	content, err = tr.Translate(source)
	if err != nil {
		return err
	}

	modTime := time.Now().UTC().Format("2006:01:02 15:04:05")

	ret, err := tx.Exec(
		"UPDATE posts SET content=?,source=?,source_type=?,modified=? WHERE id=? LIMIT 1",
		content,
		source,
		typ,
		modTime,
		id,
	)

	if err != nil {
		return err
	}

	_ = ret

	return nil
}

// GetCommentCount gets the comment count of a post.
func (z *PostManager) GetCommentCount(tx Querier, pid int64) (count int) {
	query, args := sql_helpers.NewSelect().From("posts", "").
		Select("comments").Where("id=?", pid).Limit(1).SQL()
	row := tx.QueryRow(query, args...)
	switch row.Scan(&count) {
	case sql.ErrNoRows:
		count = -1
	case nil:
		break
	default:
		count = -1
	}
	return
}

func (z *PostManager) getRowPosts(tx Querier, query string, args ...interface{}) ([]*PostForArchiveQuery, error) {
	rows, err := tx.Query(query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	ps := make([]*PostForArchiveQuery, 0)

	for rows.Next() {
		p := &PostForArchiveQuery{}
		if err = rows.Scan(&p.ID, &p.Title); err != nil {
			return nil, err
		}
		ps = append(ps, p)
	}

	return ps, rows.Err()
}

// GetPostByID gets
func (z *PostManager) GetPostByID(tx Querier, id int64, modified string) (*Post, error) {
	seldb := sql_helpers.NewSelect().From("posts", "").Select("*").Where("id=?", id)
	if datetime.IsValidMy(modified) {
		seldb.Where("modified>?", modified)
	}
	seldb.OrderBy("date DESC")
	query, args := seldb.SQL()
	row := tx.QueryRow(query, args...)
	p := Post{}
	if err := row.Scan(&p.ID, &p.Date, &p.Modified, &p.Title, &p.Content, &p.Slug, &p.Type, &p.Category, &p.Status, &p.PageView, &p.CommentStatus, &p.Comments, &p.Metas, &p.Source, &p.SourceType); err != nil {
		return nil, err
	}
	p.Date = datetime.My2Local(p.Date)
	p.Modified = datetime.My2Local(p.Modified)
	p.Tags, _ = tagmgr.GetObjectTagNames(gdb, p.ID)
	return &p, nil
}

// GetPostBySlug gets
func (z *PostManager) GetPostBySlug(tx Querier, taxTreeOrParents string, slug string, modified string, isPage bool) (*Post, error) {
	var catID int64
	var err error
	if !isPage {
		catID, err = catmgr.ParseTree(tx, taxTreeOrParents)
	} else {
		catID, err = z.GetPageParentID(tx, taxTreeOrParents)
		// this is a hack
		if err != nil {
			err = sql.ErrNoRows
		}
	}
	if err != nil {
		return nil, err
	}
	if modified != "" && !datetime.IsValidMy(modified) {
		return nil, fmt.Errorf("invalid modified")
	}
	query, args := sql_helpers.NewSelect().From("posts", "").
		Select("*").
		Where("slug=? AND taxonomy=?", slug, catID).
		WhereIf(datetime.IsValidMy(modified), "modified>?", modified).
		WhereIf(isPage, "type = 'page'").
		OrderBy("date DESC").
		SQL()
	row := tx.QueryRow(query, args...)
	p := Post{}
	if err := row.Scan(&p.ID, &p.Date, &p.Modified, &p.Title, &p.Content, &p.Slug, &p.Type, &p.Category, &p.Status, &p.PageView, &p.CommentStatus, &p.Comments, &p.Metas, &p.Source, &p.SourceType); err != nil {
		return nil, err
	}
	p.Date = datetime.My2Local(p.Date)
	p.Modified = datetime.My2Local(p.Modified)
	p.Tags, _ = tagmgr.GetObjectTagNames(gdb, p.ID)
	return &p, nil
}

// GetPostsByCategory gets category posts.
func (z *PostManager) GetPostsByCategory(tx Querier, catID int64) ([]*PostForArchiveQuery, error) {
	query, args := sql_helpers.NewSelect().From("posts", "").
		Select("id,title").
		Where("taxonomy=?", catID).
		Where("type='post'").
		OrderBy("date DESC").
		SQL()
	return z.getRowPosts(tx, query, args...)
}

// GetPostsByTags gets tag posts.
func (z *PostManager) GetPostsByTags(tx Querier, tag string) ([]*PostForArchiveQuery, error) {
	tagObj, err := tagmgr.GetTagByName(tx, tag)
	if err != nil {
		return nil, err
	}

	ids := tagmgr.getAliasTagsAll(tx, []int64{tagObj.ID})

	query, args := sql_helpers.NewSelect().From("posts,post_tags", "").
		Select("posts.id,posts.title").
		Where("posts.id=post_tags.post_id").
		Where("post_tags.tag_id in (?)", ids).
		SQL()
	return z.getRowPosts(tx, query, args...)
}

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

// ListAllPosts lists
func (z *PostManager) ListAllPosts(tx Querier) ([]*PostForArchiveQuery, error) {
	query, args := sql_helpers.NewSelect().From("posts", "").
		Select("id,title").
		Where("type='post'").
		OrderBy("date DESC").
		SQL()
	return z.getRowPosts(tx, query, args...)
}

// GetLatest gets
func (z *PostManager) GetLatest(tx Querier, limit int64) ([]*PostForLatest, error) {
	query, args := sql_helpers.NewSelect().From("posts", "").
		Select("id,title,type").
		Where("type='post'").
		OrderBy("date DESC").
		Limit(limit).
		SQL()
	rows, err := tx.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	posts := make([]*PostForLatest, 0)
	for rows.Next() {
		var p PostForLatest
		if err := rows.Scan(&p.ID, &p.Title, &p.Type); err != nil {
			return nil, err
		}
		posts = append(posts, &p)
	}
	return posts, nil
}

// GetPostsForRss gets
func (z *PostManager) GetPostsForRss(tx Querier) ([]*PostForRss, error) {
	query, args := sql_helpers.NewSelect().From("posts", "").
		Select("id,date,title,content").OrderBy("date DESC").Limit(10).SQL()
	rows, err := tx.Query(query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	posts := make([]*PostForRss, 0)

	for rows.Next() {
		var p PostForRss
		if err := rows.Scan(&p.ID, &p.Date, &p.Title, &p.Content); err != nil {
			return nil, err
		}
		posts = append(posts, &p)
	}

	home, _ := optmgr.Get(tx, "home")

	for _, p := range posts {
		p.Date = datetime.My2Feed(p.Date)
		p.Link = fmt.Sprintf("https://%s/%d/", home, p.ID)
		p.Content = template.HTML("<![CDATA[" + strings.Replace(string(p.Content), "]]>", "]]]]><!CDATA[>", -1) + "]]>")
	}

	return posts, nil
}

// GetVars gets custom column values.
func (z *PostManager) GetVars(tx Querier, fields string, wheres string, outs ...interface{}) error {
	query, args := sql_helpers.NewSelect().From("posts", "").
		Select(fields).Where(wheres).Limit(1).SQL()
	row := tx.QueryRow(query, args...)
	return row.Scan(outs...)
}

func (z *PostManager) GetPostsForManagement(tx Querier) ([]*PostForManagement, error) {
	var dummy PostForManagement
	query, args := sql_helpers.NewSelect().From("posts", "").
		Select(dummy.Fields()).Where("type='post'").OrderBy("id DESC").SQL()
	rows, err := tx.Query(query, args...)
	if err != nil {
		return nil, err
	}

	posts := make([]*PostForManagement, 0)

	for rows.Next() {
		var post PostForManagement
		if err := rows.Scan(post.Pointers()...); err != nil {
			return nil, err
		}
		post.ToLocalTime()
		posts = append(posts, &post)
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

// IncrementPageView increases page view by one.
func (z *PostManager) IncrementPageView(tx Querier, id int64) {
	query := "UPDATE posts SET page_view=page_view+1 WHERE id=? LIMIT 1"
	tx.Exec(query, id)
}

func (z *PostManager) GetDateArchives(tx Querier) ([]*PostForDate, error) {
	query := "SELECT year,month,count(id) count FROM (SELECT id,date,year(date) year,month(date) month FROM(SELECT id,DATE_ADD(date,INTERVAL 8 HOUR) date FROM posts WHERE type='post') x) x GROUP BY year,month"
	rows, err := tx.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	ps := make([]*PostForDate, 0)

	for rows.Next() {
		p := &PostForDate{}
		if err = rows.Scan(&p.Year, &p.Month, &p.Count); err != nil {
			return nil, err
		}
		ps = append(ps, p)
	}

	return ps, rows.Err()
}

func (z *PostManager) GetPageParentID(tx Querier, parents string) (int64, error) {
	if len(parents) == 0 {
		return 0, nil
	}
	parents = parents[1:]
	slugs := strings.Split(parents, "/")
	query, args := sql_helpers.NewSelect().From("posts", "").
		Select("id,slug,taxonomy").
		Where("slug IN (?)", slugs).
		Where("type = 'page'").
		SQL()
	rows, err := tx.Query(query, args...)
	if err != nil {
		return 0, err
	}

	type Result struct {
		ID     int64
		Slug   string
		Parent int64
	}

	var results []*Result
	for rows.Next() {
		var r Result
		if err := rows.Scan(&r.ID, &r.Slug, &r.Parent); err != nil {
			return 0, err
		}
		results = append(results, &r)
	}

	var parent int64
	for i := 0; i < len(slugs); i++ {
		found := false
		for _, r := range results {
			if r.Parent == parent && r.Slug == slugs[i] {
				parent = r.ID
				found = true
				break
			}
			if !found {
				return 0, fmt.Errorf("找不到父页面：%s", slugs[i])
			}
		}
	}

	return parent, nil
}

func (z *PostManager) GetRelatedPosts(tx Querier, id int64) ([]*PostForRelated, error) {
	tagIDs := tagmgr.getTagIDs(tx, id, true)
	if len(tagIDs) == 0 {
		return []*PostForRelated{}, nil
	}
	query, args := sql_helpers.NewSelect().
		From("posts", "p").
		From("post_tags", "pt").
		Select("p.id,p.title,COUNT(p.id) relevance").
		Where("pt.post_id != ?", id).
		Where("p.id = pt.post_id").
		Where("pt.tag_id IN (?)", tagIDs).
		GroupBy("p.id").
		OrderBy("relevance DESC").
		Limit(9).
		SQL()

	rows, err := tx.Query(query, args...)
	if err != nil {
		return nil, err
	}

	relates := make([]*PostForRelated, 0)
	for rows.Next() {
		var rel PostForRelated
		var dummy int64
		if err := rows.Scan(&rel.ID, &rel.Title, &dummy); err != nil {
			return nil, err
		}
		relates = append(relates, &rel)
	}

	return relates, nil
}
