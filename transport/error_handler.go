package transport

import (
	"context"

	"github.com/go-kit/kit/log"
)

// 异常处理器
type ErrorHandler interface {
	Handle(ctx context.Context, err error)
}

// 日志异常处理器
type LogErrorHandler struct {
	logger log.Logger
}

func NewLogErrorHandler(logger log.Logger) *LogErrorHandler {
	return &LogErrorHandler{
		logger: logger,
	}
}

func (h *LogErrorHandler) Handle(ctx context.Context, err error) {
	h.logger.Log("err", err)
}

// 适配器。将func(ctx context.Context, err error)强转为ErrorHandlerFunc
// 从而实现ErrorHandler接口
type ErrorHandlerFunc func(ctx context.Context, err error)

func (f ErrorHandlerFunc) Handle(ctx context.Context, err error) {
	f(ctx, err)
}
