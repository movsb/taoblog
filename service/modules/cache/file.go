package cache

import (
	"context"
	"hash/fnv"
	"reflect"
	"sync"
	"time"

	"github.com/movsb/taoblog/setup/migration"
	"github.com/movsb/taorm"
)

const (
	OneDay   = time.Hour * 24
	OneWeek  = OneDay * 7
	OneMonth = OneDay * 30
)

// Loader 不带 Key 和 Context，因为 GetOrLoad 的时候使用闭包均已知，无需再传过来。
type Loader = func() (value []byte, ttl time.Duration, err error)

type FileCache interface {
	GetOrLoad(key FileCacheKey, loader Loader) ([]byte, error)
	Delete(key FileCacheKey)
}

type _FileCacheItem struct {
	ID         int
	Hash       int64
	CreatedAt  int64
	ExpiringAt int64
	Key        []byte
	Data       []byte
}

func (_FileCacheItem) TableName() string {
	return `cache`
}

type _FileCache struct {
	db   *taorm.DB
	lock sync.Mutex
}

type FileCacheKey interface {
	CacheKey() []byte
}

func NewFileCache(ctx context.Context, path string) FileCache {
	db := migration.InitCache(path)
	cc := &_FileCache{
		db: taorm.NewDB(db),
	}
	go cc.deleteExpired(ctx)
	return cc
}

func (c *_FileCache) deleteExpired(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Minute * 10):
			c.db.Model(_FileCacheItem{}).Where(`expiring_at<?`, time.Now().Unix()).Delete()
		}
	}
}

func (c *_FileCache) GetOrLoad(key FileCacheKey, loader Loader) ([]byte, error) {
	if val, err := c.getFromDB(key); err == nil {
		return val, nil
	} else if !taorm.IsNotFoundError(err) {
		return nil, err
	}

	// TODO 全局锁
	c.lock.Lock()
	defer c.lock.Unlock()

	if val, err := c.getFromDB(key); err == nil {
		return val, nil
	} else if !taorm.IsNotFoundError(err) {
		return nil, err
	}

	value, ttl, err := loader()
	if err != nil {
		return nil, err
	}
	item := _FileCacheItem{
		Hash:       typeHash(key),
		CreatedAt:  time.Now().Unix(),
		ExpiringAt: time.Now().Add(ttl).Unix(),
		Key:        key.CacheKey(),
		Data:       value,
	}
	if err := c.db.Model(&item).Create(); err != nil {
		return nil, err
	}

	return value, nil
}

func (c *_FileCache) getFromDB(key FileCacheKey) ([]byte, error) {
	var item _FileCacheItem
	err := c.db.Where(`hash=? AND key=?`, typeHash(key), key.CacheKey()).Find(&item)
	if err == nil {
		return item.Data, nil
	}
	return nil, err
}

func (c *_FileCache) Delete(key FileCacheKey) {
	c.db.Model(_FileCacheItem{}).Where(`hash=? AND key=?`, typeHash(key), key.CacheKey()).Delete()
}

var (
	typesRegistry = make(map[int64]string)
	typesLock     sync.Mutex
)

// 计算类型hash。
// 如果不同类型名得到相同hash，会panic
func typeHash(key any) int64 {
	ti := reflect.TypeOf(key)
	pp, name := ti.PkgPath(), ti.Name()
	if pp == `` || name == `` {
		panic(`无效的缓存键类型`)
	}
	tt := pp + name // 没写“，”，问题不大。

	hh := fnv.New64a()
	hh.Write([]byte(tt))
	sum := int64(hh.Sum64() >> 1)

	typesLock.Lock()
	defer typesLock.Unlock()
	if old, ok := typesRegistry[sum]; ok {
		if old != tt {
			panic(`类型名冲突`)
		}
	} else {
		typesRegistry[sum] = tt
	}

	return int64(sum)
}
