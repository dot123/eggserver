package contextx

import (
	"context"
	"eggServer/pkg/redisbackend"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type (
	gormDBCtxKey       struct{}
	redisBackendCtxKey struct{}
	userIDCtxKey       struct{}
	roleIDCtxKey       struct{}
	traceIDCtxKey      struct{}
	loggerCtxKey       struct{}
)

func NewUserID(ctx context.Context, userID uint64) context.Context {
	return context.WithValue(ctx, userIDCtxKey{}, userID)
}

func FromUserID(ctx context.Context) uint64 {
	v := ctx.Value(userIDCtxKey{})
	if v != nil {
		if s, ok := v.(uint64); ok {
			return s
		}
	}
	return 0
}

func NewRoleID(ctx context.Context, roleID uint64) context.Context {
	return context.WithValue(ctx, roleIDCtxKey{}, roleID)
}

func FromRoleID(ctx context.Context) uint64 {
	v := ctx.Value(roleIDCtxKey{})
	if v != nil {
		if s, ok := v.(uint64); ok {
			return s
		}
	}
	return 0
}

func NewGormDB(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, gormDBCtxKey{}, db)
}

func FromGormDB(ctx context.Context) *gorm.DB {
	v := ctx.Value(gormDBCtxKey{})
	if v != nil {
		if db, ok := v.(*gorm.DB); ok {
			return db
		}
	}
	return nil
}

func NewRB(ctx context.Context, client *redisbackend.RedisBackend) context.Context {
	return context.WithValue(ctx, redisBackendCtxKey{}, client)
}

func FromRB(ctx context.Context) *redisbackend.RedisBackend {
	v := ctx.Value(redisBackendCtxKey{})
	if v != nil {
		if client, ok := v.(*redisbackend.RedisBackend); ok {
			return client
		}
	}
	return nil
}

func NewTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDCtxKey{}, traceID)
}

func FromTraceID(ctx context.Context) (string, bool) {
	v := ctx.Value(traceIDCtxKey{})
	if v != nil {
		if s, ok := v.(string); ok {
			return s, s != ""
		}
	}
	return "", false
}

func NewLogger(ctx context.Context, l *logrus.Entry) context.Context {
	return context.WithValue(ctx, loggerCtxKey{}, l)
}

func FromLogger(ctx context.Context) *logrus.Entry {
	l := ctx.Value(loggerCtxKey{})
	if l == nil {
		fields := logrus.Fields{}
		fields["source"] = "eggserver"
		return logrus.WithContext(ctx).WithFields(fields)
	}

	return l.(*logrus.Entry)
}
