package models

import (
	"time"

	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/protocols"
	"github.com/xeonx/timeago"
)

// Comment in database.
type Comment struct {
	ID         int64  `json:"id"`
	Parent     int64  `json:"parent"`
	Root       int64  `json:"root"`
	PostID     int64  `json:"post_id"`
	Author     string `json:"author"`
	Email      string `json:"email"`
	URL        string `json:"url"`
	IP         string `json:"ip"`
	Date       int32  `json:"date"`
	SourceType string `json:"source_type"`
	Source     string `json:"source"`
	Content    string `json:"content"`
}

// TableName ...
func (Comment) TableName() string {
	return `comments`
}

// ToProtocols ...
func (c *Comment) ToProtocols(isAdmin func(email string) bool, user *auth.User, geo func(ip string) string, userIP string, avatar func(email string) int) *protocols.Comment {
	comment := protocols.Comment{
		Id:         c.ID,
		Parent:     c.Parent,
		Root:       c.Root,
		PostId:     c.PostID,
		Author:     c.Author,
		Url:        c.URL,
		Date:       c.Date,
		SourceType: c.SourceType,
		Source:     c.Source,
		Content:    c.Content,
		IsAdmin:    isAdmin(c.Email),
		DateFuzzy:  timeago.Chinese.Format(time.Unix(int64(c.Date), 0)),
		Avatar:     int32(avatar(c.Email)),
	}

	if user.IsAdmin() {
		comment.Email = c.Email
		comment.Ip = c.IP
		if geo != nil {
			comment.GeoLocation = geo(c.IP)
		}
	}

	// 管理员、（同 IP 用户 & 5️⃣分钟内） 可编辑。
	// TODO: IP：并不严格判断，比如网吧、办公室可能具有相同 IP。所以限制了时间范围。
	comment.CanEdit = c.SourceType == `markdown` && (user.IsAdmin() || (userIP == c.IP && In5min(c.Date)))

	return &comment
}

func In5min(t int32) bool {
	return time.Since(time.Unix(int64(t), 0)) < time.Minute*5
}

// Comments ...
type Comments []*Comment

// ToProtocols ...
func (cs Comments) ToProtocols(isAdmin func(s string) bool, user *auth.User, geo func(ip string) string, userIP string, avatar func(email string) int) []*protocols.Comment {
	comments := make([]*protocols.Comment, 0, len(cs))
	for _, comment := range cs {
		comments = append(comments, comment.ToProtocols(isAdmin, user, geo, userIP, avatar))
	}
	return comments
}
