package service

import (
	"fmt"
	"hash/fnv"
	"io"
	"math"
	"net/http"
	"strings"
	"sync"

	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service/modules/avatar"
)

type AvatarCache struct {
	id2email map[int]string
	email2id map[string]int
	lock     sync.RWMutex
}

func NewAvatarCache() *AvatarCache {
	return &AvatarCache{
		id2email: make(map[int]string),
		email2id: make(map[string]int),
	}
}

// 简单的“一致性”哈希生成算法。
// 此“一致性”与分布式中的“一致性”不是同一种事物。
// &math.MaxIn32：不是必须的，只是简单地为了数值更小、不要负数。
func (c *AvatarCache) id(email string) int {
	hash := fnv.New32()
	hash.Write([]byte(email))
	sum := hash.Sum32() & math.MaxInt32
	for {
		e, ok := c.id2email[int(sum)]
		if ok && e != email {
			sum++
			sum &= math.MaxInt32
			continue
		}
		break
	}
	return int(sum)
}

func (c *AvatarCache) ID(email string) int {
	c.lock.Lock()
	defer c.lock.Unlock()

	email = strings.ToLower(email)

	if id, ok := c.email2id[email]; ok {
		return id
	}

	next := c.id(email)

	c.email2id[email] = next
	c.id2email[next] = email

	return next
}

func (c *AvatarCache) Email(id int) string {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.id2email[id]
}

// GetAvatar ...
func (s *Service) GetAvatar(in *protocols.GetAvatarRequest) {
	email := s.avatarCache.Email(in.Ephemeral)
	if email == "" {
		in.SetStatus(http.StatusNotFound)
		return
	}

	p := avatar.Params{
		Headers: make(http.Header),
	}

	if in.IfModifiedSince != "" {
		p.Headers.Add("If-Modified-Since", in.IfModifiedSince)
	}
	if in.IfNoneMatch != "" {
		p.Headers.Add("If-None-Match", in.IfNoneMatch)
	}

	// TODO 并没有限制获取未公开发表文章的评论。
	resp, err := avatar.Get(email, &p)
	if err != nil {
		in.SetStatus(500)
		fmt.Fprint(in.W, err)
		return
	}

	defer resp.Body.Close()

	// 删除可能有隐私的头部字段。
	// TODO：内部缓存，只正向代理 body。
	for _, k := range knownHeaders {
		if v := resp.Header.Get(k); v != "" {
			in.SetHeader(k, v)
		}
	}

	in.SetStatus(resp.StatusCode)

	io.Copy(in.W, resp.Body)
}

var knownHeaders = []string{
	`Content-Length`,
	`Content-Type`,
	`Last-Modified`,
	`Expires`,
	`Cache-Control`,
}
