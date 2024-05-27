package twitter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"log"
	"net/url"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"
)

// 代表一条从 data/tweets.js 里面提取的推文数据。
// TODO 不是很清楚两个 entities 之间的关系，有时候没有后者，有时候有。
// TODO 可能需要合并？目前部分数据只使用 expanded。
type Tweet struct {
	// 不是很清楚这个字段。
	// 看起来都是 false。
	Retweeted bool `json:"retweeted"`

	// 推文的 ID
	ID string `json:"id"`

	Truncated bool `json:"truncated"`

	// 是否是回复。
	InReplyToStatusID string `json:"in_reply_to_status_id"`

	// 正文内容。
	// 使用  Markdown 代替。
	FullText FullText `json:"full_text"`

	// 创建时间。
	CreatedAt CreatedAt `json:"created_at"`

	ExtendedEntities struct {
		Media []struct {
			HTTP      string `json:"media_url"`
			HTTPS     string `json:"media_url_https"`
			VideoInfo struct {
				Variants []struct {
					Bitrate     int64  `json:"bitrate,string"`
					ContentType string `json:"content_type"`
					URL         string `json:"url"`
				} `json:"variants"`
			} `json:"video_info"`
		} `json:"media"`
		HashTags []struct {
			Text string `json:"text"`
		} `json:"hashtags"`
		URLs         []EntityURL     `json:"urls"`
		UserMentions []MentionedUser `json:"user_mentions"`
	} `json:"extended_entities"`

	// 有  expanded 就不用
	Entities struct {
		Media []struct {
			HTTP      string `json:"media_url"`
			HTTPS     string `json:"media_url_https"`
			VideoInfo struct {
				Variants []struct {
					Bitrate     int64  `json:"bitrate,string"`
					ContentType string `json:"content_type"`
					URL         string `json:"url"`
				} `json:"variants"`
			} `json:"video_info"`
		} `json:"media"`
		HashTags []struct {
			Text string `json:"text"`
		} `json:"hashtags"`
		URLs         []EntityURL     `json:"urls"`
		UserMentions []MentionedUser `json:"user_mentions"`
	} `json:"entities"`

	children []*Tweet
}

type EntityURL struct {
	URL      string `json:"url"`
	Expanded string `json:"expanded_url"`
}

type MentionedUser struct {
	Name       string `json:"name"`        // 昵称
	ScreenName string `json:"screen_name"` // 帐号
	ID         string `json:"id"`
}

type CreatedAt int32

func (t *CreatedAt) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	const layout = `Mon Jan 02 15:04:05 -0700 2006`
	t2, err := time.Parse(layout, s)
	if err != nil {
		return fmt.Errorf(`error parsing: %s, %w`, s, err)
	}
	*t = CreatedAt(t2.Unix())
	return nil
}

type FullText string

var reRemoveUrl = regexp.MustCompile(`https://t\.co/[^\s]+`)
var replaceUserMention = regexp.MustCompile(`@\w+\b`)

func (t *Tweet) Markdown() string {
	s := reRemoveUrl.ReplaceAllStringFunc(string(t.FullText), func(s string) string {
		expanded := ""
		if p := slices.IndexFunc(t.ExtendedEntities.URLs, func(u EntityURL) bool {
			return u.URL == s
		}); p != -1 {
			expanded = t.ExtendedEntities.URLs[p].Expanded
		}
		if p := slices.IndexFunc(t.Entities.URLs, func(u EntityURL) bool {
			return u.URL == s
		}); p != -1 {
			expanded = t.Entities.URLs[p].Expanded
		}
		if expanded != "" {
			// 粗暴判断不是站内链接。
			if !strings.Contains(expanded, `/status/`) {
				return fmt.Sprintf(`<%s>`, expanded)
			}
		}
		return "" // 原本推文自身/转推/站内。
	})
	s = strings.ReplaceAll(s, "\n", "\n\n")
	s = replaceUserMention.ReplaceAllStringFunc(s, func(s string) string {
		if p := slices.IndexFunc(t.Entities.UserMentions, func(u MentionedUser) bool {
			return u.ScreenName == s[1:]
		}); p != -1 {
			u := t.Entities.UserMentions[p]
			// ID转用户名。
			// https://stackoverflow.com/a/56924385/3628322
			// https://twitter.com/i/user/50988711
			buf := bytes.NewBuffer(nil)
			if err := userMentionTmpl.Execute(buf, u); err != nil {
				panic(err)
			}
			return buf.String()
		}
		return s
	})
	return s
}

var userMentionTmpl = template.Must(template.New(`mention`).Parse(`<a class="tweet-user-mention" href="https://twitter.com/i/user/{{.ID}}" title="{{.Name}}" target="_blank">@{{.ScreenName}}</a>`))

func (s *FullText) UnmarshalJSON(data []byte) error {
	var x string
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}

	// 没找到文档具体转了些啥，看样子转。
	x = html.UnescapeString(x)

	// x = reRemoveUrl.ReplaceAllStringFunc(x, "")

	*s = FullText(x)

	return nil
}

// 是否是独立发推（而不是回复之类）。
func (t *Tweet) IsSelfTweet() bool {
	return t.InReplyToStatusID == ""
}

type VideoAsset struct {
	FileName      string
	PosterURL     string
	Poster        string
	ContentType   string
	Width, Height int
}

func (t *Tweet) Assets(all bool) ([]string, []VideoAsset) {
	images, videos := assets(t)

	// 为简单起见，直接嵌入原推。
	var recurse func(t *Tweet)
	recurse = func(t *Tweet) {
		i, v := assets(t)
		images = append(images, i...)
		videos = append(videos, v...)
		for _, c := range t.children {
			recurse(c)
		}
	}

	if all {
		for _, c := range t.children {
			recurse(c)
		}
	}

	return images, videos
}

func assets(t *Tweet) ([]string, []VideoAsset) {
	var images []string
	var videos []VideoAsset

	for _, medium := range t.ExtendedEntities.Media {
		if len(medium.VideoInfo.Variants) == 0 {
			u, err := url.Parse(medium.HTTPS)
			if err != nil {
				panic(err)
			}
			image := fmt.Sprintf(`%s-%s`, t.ID, filepath.Base(u.Path))
			images = append(images, image)
		} else {
			var max int64
			var index int
			for i, v := range medium.VideoInfo.Variants {
				if v.Bitrate > max {
					max = v.Bitrate
					index = i
				}
			}
			v := medium.VideoInfo.Variants[index]
			u, err := url.Parse(v.URL)
			if err != nil {
				log.Fatalln(err)
			}
			video := fmt.Sprintf(`%s-%s`, t.ID, filepath.Base(u.Path))

			var width, height int
			if match := reDimension.FindString(v.URL); match != "" {
				fmt.Sscanf(match, `/%dx%d/`, &width, &height)
			}

			var poster string
			if u, err := url.Parse(medium.HTTPS); err == nil {
				poster = filepath.Base(u.Path)
			}

			videos = append(videos, VideoAsset{
				FileName:    video,
				PosterURL:   medium.HTTPS,
				Poster:      poster,
				ContentType: v.ContentType,
				Width:       width,
				Height:      height,
			})
		}
	}

	return images, videos
}

var reDimension = regexp.MustCompile(`/\d+x\d+/`)

func (t *Tweet) TagNames() []string {
	names := map[string]struct{}{}
	for _, t := range t.ExtendedEntities.HashTags {
		names[t.Text] = struct{}{}
	}
	for _, t := range t.Entities.HashTags {
		names[t.Text] = struct{}{}
	}
	list := make([]string, 0, len(names))
	for n := range names {
		list = append(list, n)
	}
	return list
}
