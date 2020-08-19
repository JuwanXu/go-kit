package lb

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-kit/kit/endpoint"
)

// 重拾的异常
type RetryError struct {
	RawErrors []error // all errors encountered from endpoints directly
	Final     error   // the final, terminating error
}

func (e RetryError) Error() string {
	var suffix string
	if len(e.RawErrors) > 1 {
		a := make([]string, len(e.RawErrors)-1)
		for i := 0; i < len(e.RawErrors)-1; i++ { // last one is Final
			a[i] = e.RawErrors[i].Error()
		}
		suffix = fmt.Sprintf(" (previously: %s)", strings.Join(a, "; "))
	}
	return fmt.Sprintf("%v%s", e.Final, suffix)
}

// 重试函数
type Callback func(n int, received error) (keepTrying bool, replacement error)

// 返回一个Endpoint
// 用户使用需要先将Endpointer封装为Balancer
// 再将Balancer用Retry封装一层
// 用户再定义结构体封装Retry封装过的Endpoint
func Retry(max int, timeout time.Duration, b Balancer) endpoint.Endpoint {
	return RetryWithCallback(timeout, b, maxRetries(max))
}

func maxRetries(max int) Callback {
	return func(n int, err error) (keepTrying bool, replacement error) {
		return n < max, nil
	}
}

func alwaysRetry(int, error) (keepTrying bool, replacement error) {
	return true, nil
}

func RetryWithCallback(timeout time.Duration, b Balancer, cb Callback) endpoint.Endpoint {
	if cb == nil {
		cb = alwaysRetry
	}
	if b == nil {
		panic("nil Balancer")
	}
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		var (
			newctx, cancel = context.WithTimeout(ctx, timeout)
			responses      = make(chan interface{}, 1)
			errs           = make(chan error, 1)
			final          RetryError
		)
		defer cancel()
		for i := 1; ; i++ {
			go func() {
				e, err := b.Endpoint()
				if err != nil {
					errs <- err
					return
				}
				response, err := e(newctx, request)
				if err != nil {
					errs <- err
					return
				}
				responses <- response
			}()

			select {
			case <-newctx.Done():
				return nil, newctx.Err()
			case response := <-responses:
				return response, nil
			case err := <-errs:
				final.RawErrors = append(final.RawErrors, err)
				keepTrying, replacement := cb(i, err)
				if replacement != nil {
					err = replacement
				}
				if !keepTrying {
					final.Final = err
					return nil, final
				}
				continue
			}
		}
	}
}
