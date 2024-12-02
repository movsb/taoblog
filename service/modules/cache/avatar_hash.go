package cache

import (
	"hash/fnv"
	"math"
	"strings"
	"sync"
)

type AvatarHasher interface {
	ID(email string) int
	Email(id int) string
}

type AvatarHash struct {
	id2email map[int]string
	email2id map[string]int
	lock     sync.RWMutex
}

func NewAvatarHash() *AvatarHash {
	return &AvatarHash{
		id2email: make(map[int]string),
		email2id: make(map[string]int),
	}
}

// 简单的“一致性”哈希生成算法。
// 此“一致性”与分布式中的“一致性”不是同一种事物。
// &math.MaxIn32：不是必须的，只是简单地为了数值更小、不要负数。
func (c *AvatarHash) id(email string) int {
	hash := fnv.New32()
	hash.Write([]byte(email))
	sum := hash.Sum32() & math.MaxInt32
	for {
		e, ok := c.id2email[int(sum)]
		if ok && e != email {
			// 开放寻址法（Open Addressing）
			sum++
			sum &= math.MaxInt32
			continue
		}
		break
	}
	return int(sum)
}

func (c *AvatarHash) ID(email string) int {
	c.lock.Lock()
	defer c.lock.Unlock()

	if email == "" {
		panic("错误的邮箱。")
	}

	email = strings.ToLower(email)

	if id, ok := c.email2id[email]; ok {
		return id
	}

	next := c.id(email)

	c.email2id[email] = next
	c.id2email[next] = email

	return next
}

// 根据 ID 获取 Email。
// 如果不存在，返回空。
func (c *AvatarHash) Email(id int) string {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.id2email[id]
}
