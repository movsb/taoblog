package main

import (
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"strings"
	"time"

	"./internal/post_translators"
	"./internal/utils/datetime"
)

// PostForArchiveQuery is an archive query result.
type PostForArchiveQuery struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
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
	query := `SELECT comments FROM posts WHERE id=` + fmt.Sprint(pid) + ` LIMIT 1`
	row := tx.QueryRow(query)
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

func (z *PostManager) beforeQuery(q map[string]interface{}) map[string]interface{} {
	if _, ok := q["where"]; !ok {
		q["where"] = []string{}
	}
	ws := q["where"].([]string)
	ws = append(ws, "status='public'")
	q["where"] = ws
	return q
}

func (z *PostManager) getRowPosts(tx Querier, q map[string]interface{}) ([]*PostForArchiveQuery, error) {
	q = z.beforeQuery(q)
	s := BuildQueryString(q)

	rows, err := tx.Query(s)
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

// GetPostsByCategory gets category posts.
func (z *PostManager) GetPostsByCategory(tx Querier, catID int64) ([]*PostForArchiveQuery, error) {
	q := make(map[string]interface{})

	q["select"] = "id,title"
	q["from"] = "posts"
	q["where"] = []string{
		fmt.Sprintf("taxonomy=%d", catID),
		"type='post'",
	}
	q["orderby"] = "date DESC"

	return z.getRowPosts(tx, q)
}

// GetPostsByTags gets tag posts.
func (z *PostManager) GetPostsByTags(tx Querier, tag string) ([]*PostForArchiveQuery, error) {
	tagObj, err := tagmgr.GetTagByName(tx, tag)
	if err != nil {
		return nil, err
	}

	ids := tagmgr.getAliasTagsAll(tx, []int64{tagObj.ID})

	q := make(map[string]interface{})

	q["select"] = "posts.id,posts.title"
	q["from"] = "posts,post_tags"
	q["where"] = []string{
		"posts.id=post_tags.post_id",
		fmt.Sprintf("post_tags.tag_id in (%s)", joinInts(ids, ",")),
	}

	return z.getRowPosts(tx, q)
}

// GetPostsByDate gets date posts.
func (z *PostManager) GetPostsByDate(tx Querier, yy, mm int64) ([]*PostForArchiveQuery, error) {
	q := make(map[string]interface{})
	q["select"] = "id,title"
	q["from"] = "posts"
	q["where"] = []string{
		"type='post'",
	}
	if yy > 1970 {
		var start, end string
		if 1 <= mm && mm <= 12 {
			start, end = datetime.MonthStartEnd(int(yy), int(mm))
		} else {
			start, end = datetime.YearStartEnd(int(yy))
		}
		q["where"] = append(q["where"].([]string), fmt.Sprintf("date>='%s' AND date<='%s'", start, end))
	}

	q["orderby"] = "date DESC"

	return z.getRowPosts(tx, q)
}

// ListAllPosts lists
func (z *PostManager) ListAllPosts(tx Querier) ([]*PostForArchiveQuery, error) {
	q := make(map[string]interface{})
	q["select"] = "id,title"
	q["from"] = "posts"
	q["where"] = []string{
		"type='post'",
	}
	q["orderby"] = "date DESC"

	q = z.beforeQuery(q)

	return z.getRowPosts(tx, q)
}

// GetPostsForRss gets
func (z *PostManager) GetPostsForRss(tx Querier) ([]*PostForRss, error) {
	q := make(map[string]interface{})
	q["select"] = "id,date,title,content"
	q["from"] = "posts"
	q["orderby"] = "date DESC"
	q["limit"] = 10

	query := BuildQueryString(q)

	rows, err := tx.Query(query)
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
	q := make(map[string]interface{})
	q["select"] = fields
	q["from"] = "posts"
	q["where"] = []string{
		wheres,
	}
	q["limit"] = 1

	query := BuildQueryString(q)

	row := tx.QueryRow(query)

	return row.Scan(outs...)
}

func (z *PostManager) GetPostsForManagement(tx Querier) ([]*PostForManagement, error) {
	var dummy PostForManagement
	cols := dummy.Fields()
	q := make(map[string]interface{})
	q["select"] = cols
	q["from"] = "posts"
	q["where"] = []string{"type='post'"}
	q["orderby"] = "id DESC"

	query := BuildQueryString(q)
	rows, err := tx.Query(query)
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
	q := make(map[string]interface{})
	q["select"] = "count(*) as size"
	q["from"] = "posts"
	q["where"] = []string{
		"type=?",
	}
	query := BuildQueryString(q)
	row := tx.QueryRow(query, typ)
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
