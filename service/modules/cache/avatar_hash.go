package cache

import (
	"encoding/binary"
	"hash/fnv"
	"math"
	"strings"
	"sync"
)

type AvatarHash struct {
	id2email map[uint32]string
	email2id map[string]uint32

	id2user map[uint32]int
	user2id map[int]uint32

	lock sync.RWMutex
}

func NewAvatarHash() *AvatarHash {
	return &AvatarHash{
		id2email: make(map[uint32]string),
		email2id: make(map[string]uint32),
		id2user:  make(map[uint32]int),
		user2id:  make(map[int]uint32),
	}
}

const (
	typeBits = 0b1
	_User    = 0b0
	_Email   = 0b1
	maxSum   = uint32(math.MaxInt32 >> typeBits)
)

// 简单的“一致性”哈希生成算法。
// 此“一致性”与分布式中的“一致性”不是同一种事物。
// 使用开放寻址法（Open Addressing）解决冲突。
func (c *AvatarHash) id(data []byte, exists func(sum uint32) bool) uint32 {
	hash := fnv.New32()
	hash.Write(data)
	sum := hash.Sum32()
	for {
		sum &= maxSum
		if !exists(sum) {
			return sum
		}
		sum++
	}
}

func makeType(out *uint32, ty uint32) {
	*out <<= typeBits
	*out |= ty
}

func (c *AvatarHash) Email(email string) (out uint32) {
	c.lock.Lock()
	defer c.lock.Unlock()

	defer makeType(&out, _Email)

	email = strings.ToLower(email)
	if id, ok := c.email2id[email]; ok {
		return id
	}

	next := c.id(
		[]byte(email),
		func(sum uint32) bool {
			_, ok := c.id2email[sum]
			return ok
		})

	c.email2id[email] = next
	c.id2email[next] = email

	return next
}

func (c *AvatarHash) User(user int) (out uint32) {
	c.lock.Lock()
	defer c.lock.Unlock()

	defer makeType(&out, _User)

	if id, ok := c.user2id[user]; ok {
		return id
	}

	next := c.id(
		binary.LittleEndian.AppendUint32(nil, uint32(user)),
		func(sum uint32) bool {
			_, ok := c.id2user[sum]
			return ok
		},
	)

	c.user2id[user] = next
	c.id2user[next] = user

	return next
}

func (c *AvatarHash) Resolve(id uint32) (user int, email string) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	sum := id >> typeBits

	switch id & typeBits {
	case _User:
		user = c.id2user[sum]
	case _Email:
		email = c.id2email[sum]
	}

	return
}
