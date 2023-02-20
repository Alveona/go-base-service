package base_swagger_service

import (
	"context"

	"github.com/sirupsen/logrus"
)

type DebugLogger struct{}

func NewDebugLogger() *DebugLogger {
	return &DebugLogger{}
}

func (dl *DebugLogger) IsEnable(ctx context.Context) bool {
	return true
}

func (dl *DebugLogger) LogRequest(ctx context.Context, request string) {
	logrus.Infof(request)
}

func (dl *DebugLogger) LogResponse(ctx context.Context, response string, kvs ...interface{}) {
	logrus.Infof(response, kvs...)
}
