package logging

import (
	"context"

	"github.com/sirupsen/logrus"
)

type contextKey string

const (
	loggerContextKey contextKey = "logger.entry"
)

// FromContext 从上下文读取结构化 logger；缺失时返回全局 logger entry。
func FromContext(ctx context.Context) *logrus.Entry {
	if ctx == nil {
		return logrus.NewEntry(logrus.StandardLogger())
	}
	if entry, ok := ctx.Value(loggerContextKey).(*logrus.Entry); ok && entry != nil {
		return entry
	}
	return logrus.NewEntry(logrus.StandardLogger())
}

// WithField 基于上下文 logger 添加单个字段。
func WithField(ctx context.Context, key string, value any) context.Context {
	entry := FromContext(ctx).WithField(key, value)
	return context.WithValue(ctx, loggerContextKey, entry)
}

// WithFields 基于上下文 logger 添加多个字段。
func WithFields(ctx context.Context, fields logrus.Fields) context.Context {
	entry := FromContext(ctx).WithFields(fields)
	return context.WithValue(ctx, loggerContextKey, entry)
}
