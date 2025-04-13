package cache

import (
	"context"
	"encoding/json"
	"hash/fnv"
	"reflect"
	"sync"
	"time"

	"github.com/movsb/taoblog/setup/migration"
	"github.com/movsb/taorm"
)

// Loader 不带 Key 和 Context，因为 GetOrLoad 的时候使用闭包均已知，无需再传过来。
type Loader = func() (any, error)
type Getter = func(key any, ttl time.Duration, out any, loader Loader) error

func DirectLoader(key any, ttl time.Duration, out any, loader Loader) error {
	val, err := loader()
	if err == nil {
		reflect.ValueOf(out).Elem().Set(reflect.ValueOf(val))
	}
	return err
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

type FileCache struct {
	db   *taorm.DB
	lock sync.Mutex
}

func NewFileCache(ctx context.Context, path string) *FileCache {
	db := migration.InitCache(path)
	cc := &FileCache{
		db: taorm.NewDB(db),
	}
	go cc.deleteExpired(ctx)
	return cc
}

func (c *FileCache) deleteExpired(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Minute * 10):
			c.db.Model(_FileCacheItem{}).Where(`expiring_at<?`, time.Now().Unix()).Delete()
		}
	}
}

// key 应该为结构体，并可被 json 化。
func (c *FileCache) GetOrLoad(key any, ttl time.Duration, out any, loader Loader) error {
	if err := c.getFromDB(key, out); err == nil {
		return nil
	} else if !taorm.IsNotFoundError(err) {
		return err
	}

	// TODO 全局锁
	c.lock.Lock()
	defer c.lock.Unlock()

	if err := c.getFromDB(key, out); err == nil {
		return nil
	} else if !taorm.IsNotFoundError(err) {
		return err
	}

	value, err := loader()
	if err != nil {
		return err
	}

	// reflect.ValueOf(out).Elem().Set(reflect.ValueOf(value))

	item := _FileCacheItem{
		Hash:       typeHash(key),
		CreatedAt:  time.Now().Unix(),
		ExpiringAt: time.Now().Add(ttl).Unix(),
		Key:        encode(key),
		Data:       encode(value),
	}
	if err := c.db.Model(&item).Create(); err != nil {
		return err
	}

	return c.getFromDB(key, out)
}

func (c *FileCache) getFromDB(key any, out any) error {
	var item _FileCacheItem
	err := c.db.Where(`hash=? AND key=?`, typeHash(key), encode(key)).Find(&item)
	if err == nil {
		decode(item.Data, out)
		return nil
	}
	return err
}

func (c *FileCache) Delete(key any) {
	c.db.Model(_FileCacheItem{}).Where(`hash=? AND key=?`, typeHash(key), encode(key)).Delete()
}

func encode(k any) []byte {
	j, _ := json.Marshal(k)
	return j
}
func decode(d []byte, out any) {
	json.Unmarshal(d, out)
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
