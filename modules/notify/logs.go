package notify

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/movsb/taoblog/modules/logs"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taorm"
)

var _ logs.Logger = (*_LogStore)(nil)

func toJSON(data any) (string, error) {
	d, err := json.Marshal(data)
	if err != nil {
		return ``, err
	}
	return string(d), nil
}

func fromJSON(j string, data any) error {
	return json.Unmarshal([]byte(j), data)
}

type _LogStore struct {
	tdb *taorm.DB
}

func NewLogStore(db *sql.DB) *_LogStore {
	return &_LogStore{
		tdb: taorm.NewDB(db),
	}
}

func (s *_LogStore) CreateLog(ctx context.Context, ty, subType string, version int, data any) {
	l := models.Log{
		Time:    time.Now().Unix(),
		Type:    ty,
		SubType: subType,
		Version: version,
		Data:    utils.Must1(toJSON(data)),
	}
	s.tdb.Model(&l).MustCreate()
	log.Println(`创建日志：`, l.ID, l.Type, l.SubType, l.Data)
}

// 找不到时返回 nil
func (s *_LogStore) FindLog(ctx context.Context, ty, subType string, data any) *models.Log {
	var log models.Log
	err := s.tdb.Where(`type=? AND sub_type=?`, ty, subType).OrderBy(`time asc`).Find(&log)
	if err == nil {
		utils.Must(fromJSON(log.Data, data))
		return &log
	}
	if !taorm.IsNotFoundError(err) {
		panic(err)
	}
	return nil
}

func (s *_LogStore) DeleteLog(ctx context.Context, id int64) {
	var l models.Log
	s.tdb.Select(`id`).Where(`id=?`, id).MustFind(&l)
	s.tdb.Model(&l).MustDelete()
	log.Println(`删除日志：`, l.ID)
}
