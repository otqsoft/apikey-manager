// Package semaphore 按供应商维度的简单并发信号量（契约：max_concurrency 限流）。
// 每个供应商一个带缓冲 channel；max<=0 视为不限流，直接放行。
package semaphore

import (
	"context"
	"sync"
)

type pool struct {
	mu    sync.Mutex
	chans map[uint]chan struct{}
}

var defaultPool = &pool{chans: make(map[uint]chan struct{})}

// Acquire 获取一个供应商许可；返回释放函数；ctx 取消时返回错误。
func Acquire(ctx context.Context, providerID uint, max int) (release func(), err error) {
	if max <= 0 {
		return func() {}, nil
	}
	defaultPool.mu.Lock()
	ch, ok := defaultPool.chans[providerID]
	if !ok {
		ch = make(chan struct{}, max)
		defaultPool.chans[providerID] = ch
	}
	defaultPool.mu.Unlock()

	select {
	case ch <- struct{}{}:
		return func() { <-ch }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
