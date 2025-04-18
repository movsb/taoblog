package logs

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taorm"
)

type Logger interface {
	CreateLog(ctx context.Context, ty, subType string, version int, data any)
	FindLog(ctx context.Context, ty, subType string, data any) *models.Log
	DeleteLog(ctx context.Context, id int64)
}

var _ Logger = (*LogStore)(nil)

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

type LogStore struct {
	tdb *taorm.DB
}

func NewLogStore(db *sql.DB) *LogStore {
	return &LogStore{
		tdb: taorm.NewDB(db),
	}
}

func (s *LogStore) DeleteAllLogs(ctx context.Context) {
	s.tdb.From(models.Log{}).MustDeleteAnyway()
}

func (s *LogStore) CountStaleLogs(ago time.Duration) int {
	minutesAgo := time.Now().Add(-ago)
	var count int
	s.tdb.From(models.Log{}).Where(`time < ?`, minutesAgo.Unix()).Count(&count)
	return count
}

func (s *LogStore) CreateLog(ctx context.Context, ty, subType string, version int, data any) {
	d := utils.Must1(toJSON(data))
	l := models.Log{
		Time:    time.Now().Unix(),
		Type:    ty,
		SubType: subType,
		Version: version,
		Data:    d,
	}
	s.tdb.Model(&l).MustCreate()
	log.Println(`创建日志：`, l.ID, l.Type, l.SubType, l.Data)
}

// 找不到时返回 nil
func (s *LogStore) FindLog(ctx context.Context, ty, subType string, data any) *models.Log {
	var log models.Log
	err := s.tdb.Where(`type=? AND sub_type=?`, ty, subType).OrderBy(`time asc`).Find(&log)
	if err == nil {
		utils.Must(fromJSON(log.Data, data))
		return &log
	}
	if !taorm.IsNotFoundError(err) {
		// TODO 正确处理关闭数据库的情况
		if !strings.Contains(err.Error(), `database is closed`) {
			panic(err)
		}
	}
	return nil
}

func (s *LogStore) DeleteLog(ctx context.Context, id int64) {
	var l models.Log
	s.tdb.Select(`id`).Where(`id=?`, id).MustFind(&l)
	s.tdb.Model(&l).MustDelete()
	log.Println(`删除日志：`, l.ID)
}
