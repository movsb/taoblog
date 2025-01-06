package logs

import (
	"context"

	"github.com/movsb/taoblog/service/models"
)

type Logger interface {
	CreateLog(ctx context.Context, ty, subType string, version int, data any)
	FindLog(ctx context.Context, ty, subType string, data any) *models.Log
	DeleteLog(ctx context.Context, id int64)
}
