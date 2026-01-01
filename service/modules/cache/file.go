package cache

import (
	"context"
	"encoding/json"
	"hash/fnv"
	"log"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/movsb/taoblog/modules/version"
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
	Type       int64
	CreatedAt  int64
	ExpiringAt int64
	Key        []byte
	Data       []byte
}

func (_FileCacheItem) TableName() string {
	return `cache`
}

type FileCache struct {
	db      *taorm.DB
	lock    sync.Mutex
	touched map[int]int64
}

func NewFileCache(ctx context.Context, db *taorm.DB) *FileCache {
	cc := &FileCache{
		db:      db,
		touched: make(map[int]int64),
	}
	go cc.run(ctx)
	// 可能是刚从备份恢复的数据库，立即清理一下。
	go time.AfterFunc(time.Second*10, cc.sync)
	return cc
}

func (c *FileCache) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Hour):
			c.sync()
		}
	}
}

// key 应该为结构体，并可被 json 化。
func (c *FileCache) GetOrLoad(key any, ttl time.Duration, out any, loader Loader) error {
	if err := c.getFromDB(key, ttl, out, false); err == nil {
		return nil
	} else if !taorm.IsNotFoundError(err) {
		return err
	}

	// TODO 全局锁
	c.lock.Lock()
	defer c.lock.Unlock()

	if err := c.getFromDB(key, ttl, out, true); err == nil {
		return nil
	} else if !taorm.IsNotFoundError(err) {
		return err
	}

	value, err := loader()
	if err != nil {
		return err
	}

	if err := c.create(key, value, ttl); err != nil {
		return err
	}

	// 确保数据是正常获取的。
	// reflect.ValueOf(out).Elem().Set(reflect.ValueOf(value))
	return c.getFromDB(key, ttl, out, true)
}

func (c *FileCache) getFromDB(key any, ttl time.Duration, out any, locked bool) error {
	var item _FileCacheItem
	err := c.db.Where(`type=? AND key=?`, typeHash(key), encode(key)).Find(&item)
	if err == nil {
		decode(item.Data, out)
		c.touch(item.ID, ttl, locked)
		return nil
	}
	return err
}

func (c *FileCache) touch(id int, ttl time.Duration, locked bool) {
	if !locked {
		c.lock.Lock()
		defer c.lock.Unlock()
	}

	c.touched[id] = time.Now().Add(ttl).Unix()
}

// TODO 放在事务中。
func (c *FileCache) sync() {
	c.lock.Lock()
	defer c.lock.Unlock()

	for id, t := range c.touched {
		c.db.From(_FileCacheItem{}).Where(`id=?`, id).UpdateMap(taorm.M{
			`expiring_at`: t,
		})
	}

	if len(c.touched) > 0 {
		log.Println(`更新文件缓存并清空已过期缓存：`, len(c.touched))
	}

	clear(c.touched)

	c.db.Model(_FileCacheItem{}).Where(`expiring_at<?`, time.Now().Unix()).Delete()
}

func (c *FileCache) Delete(key any) {
	c.db.Model(_FileCacheItem{}).Where(`type=? AND key=?`, typeHash(key), encode(key)).Delete()
}

func (c *FileCache) DeleteForType(keyZero any) {
	c.db.Model(_FileCacheItem{}).Where(`type=?`, typeHash(keyZero)).Delete()
}

func (c *FileCache) create(key any, value any, ttl time.Duration) error {
	item := _FileCacheItem{
		Type:       typeHash(key),
		CreatedAt:  time.Now().Unix(),
		ExpiringAt: time.Now().Add(ttl).Unix(),
		Key:        encode(key),
		Data:       encode(value),
	}
	return c.db.Model(&item).Create()
}

func (c *FileCache) Set(key, value any, ttl time.Duration) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	typ := typeHash(key)
	enc := encode(key)

	var item _FileCacheItem
	if err := c.db.Select(`id`).Where(`type=? AND key=?`, typ, enc).Find(&item); err != nil {
		if !taorm.IsNotFoundError(err) {
			return err
		}
		return c.create(key, value, ttl)
	} else {
		_, err := c.db.Model(_FileCacheItem{}).Where(`id=?`, item.ID).UpdateMap(taorm.M{
			`expiring_at`: time.Now().Add(ttl).Unix(),
			`data`:        encode(value),
		})
		return err
	}
}

func (c *FileCache) Peek(key any, out any, expiringAt *time.Time) error {
	var item _FileCacheItem
	err := c.db.Where(`type=? AND key=?`, typeHash(key), encode(key)).Find(&item)
	if err != nil {
		return err
	}
	decode(item.Data, out)
	if expiringAt != nil {
		*expiringAt = time.Unix(item.ExpiringAt, 0)
	}
	return nil
}

// 获取某种特定类型的 key 的所有 keys。
// 并不是全部数据库的所有 key。
//
// NOTE: 不会刷新缓存的 TTL 时间。
func (c *FileCache) GetAllKeysFor(out any) {
	t := reflect.TypeOf(out)
	if t.Kind() != reflect.Pointer {
		panic(`not pointer`)
	}
	t = t.Elem()
	if t.Kind() != reflect.Slice {
		panic(`not slice`)
	}
	sliceType := t
	keyType := sliceType.Elem()
	if keyType.Kind() == reflect.Pointer {
		keyType = keyType.Elem()
	}
	keyVal := reflect.New(keyType).Elem().Interface()

	var items []*_FileCacheItem
	c.db.Where(`type=?`, typeHash(keyVal)).MustFind(&items)

	ss := reflect.MakeSlice(sliceType, len(items), len(items))
	for i, item := range items {
		k := reflect.New(keyType)
		decode(item.Key, k.Interface())

		if sliceType.Elem().Kind() == reflect.Pointer {
			ss.Index(i).Set(k)
		} else {
			ss.Index(i).Set(k.Elem())
		}
	}
	reflect.ValueOf(out).Elem().Set(ss)
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
	if ti.Kind() == reflect.Pointer {
		ti = ti.Elem()
	}
	pp, name := ti.PkgPath(), ti.Name()
	if pp == `` || name == `` {
		panic(`无效的缓存键类型`)
	}
	tt := pp + "." + name

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
		_, after, found := strings.Cut(tt, version.NameLowercase)
		if found {
			tt = after[1:]
		}
		log.Println(`注册类型：`, tt, sum)
	}

	return int64(sum)
}
