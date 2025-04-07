package db

import (
	"context"

	"github.com/movsb/taorm"
)

type dbKey struct{}

func FromContext(ctx context.Context) *taorm.DB {
	val, _ := ctx.Value(dbKey{}).(*taorm.DB)
	return val
}

func FromContextDefault(ctx context.Context, db *taorm.DB) *taorm.DB {
	if val := FromContext(ctx); val != nil {
		return val
	}
	return db
}

func WithContext(ctx context.Context, db *taorm.DB) context.Context {
	return context.WithValue(ctx, dbKey{}, db)
}
