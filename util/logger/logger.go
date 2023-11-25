package logger

import (
	"context"
	"github.com/hashicorp/go-uuid"
	"github.com/sirupsen/logrus"
)

const ContextKeyRequestId string = "requestId"
const ContextKeyTaskId string = "taskId"

func ContextWithRequestId(ctx context.Context, reqId string) context.Context {
	if reqId == "" {
		reqId, _ = uuid.GenerateUUID()
	}
	return context.WithValue(ctx, ContextKeyRequestId, reqId)
}

func ContextWithTaskId(ctx context.Context, taskId string) context.Context {
	if taskId == "" {
		taskId, _ = uuid.GenerateUUID()
	}
	return context.WithValue(ctx, ContextKeyTaskId, taskId)
}

const LogContextKeyRequestId string = "requestId"
const LogContextKeyTaskId string = "taskId"

func Context(ctx context.Context) *logrus.Entry {
	if ctx == nil {
		return logrus.NewEntry(logrus.StandardLogger())
	}
	reqIdRaw := ctx.Value(ContextKeyRequestId)
	if reqId, ok := reqIdRaw.(string); ok {
		return logrus.WithField(LogContextKeyRequestId, reqId)
	}
	taskIdRaw := ctx.Value(ContextKeyTaskId)
	if taskId, ok := taskIdRaw.(string); ok {
		return logrus.WithField(LogContextKeyTaskId, taskId)
	}
	return logrus.NewEntry(logrus.StandardLogger())
}

func Entry(ctx context.Context) *logrus.Entry {
	return Context(ctx)
}

func Debugf(ctx context.Context, format string, args ...interface{}) {
	Context(ctx).Debugf(format, args)
}

func Infof(ctx context.Context, format string, args ...interface{}) {
	Context(ctx).Infof(format, args)
}

func Warnf(ctx context.Context, format string, args ...interface{}) {
	Context(ctx).Warnf(format, args)
}

func Errorf(ctx context.Context, format string, args ...interface{}) {
	Context(ctx).Errorf(format, args)
}

func Debug(ctx context.Context, format string, fields map[string]interface{}) {
	Context(ctx).WithFields(fields).Debug(format)
}

func Info(ctx context.Context, format string, fields map[string]interface{}) {
	Context(ctx).WithFields(fields).Info(format)
}

func Warn(ctx context.Context, format string, fields map[string]interface{}) {
	Context(ctx).WithFields(fields).Warn(format)
}

func Error(ctx context.Context, format string, fields map[string]interface{}) {
	Context(ctx).WithFields(fields).Error(format)
}
