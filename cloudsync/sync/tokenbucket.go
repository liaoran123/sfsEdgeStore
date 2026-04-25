package sync

import (
	"sync"
	"time"
)

// TokenBucket 令牌桶算法
type TokenBucket struct {
	capacity     int           // 桶容量
	rate         int           // 每秒生成令牌数
	tokens       int           // 当前令牌数
	lastRefillTime time.Time    // 上次填充时间
	mutex        sync.Mutex    // 互斥锁
}

// NewTokenBucket 创建令牌桶
func NewTokenBucket(capacity, rate int) *TokenBucket {
	return &TokenBucket{
		capacity:     capacity,
		rate:         rate,
		tokens:       capacity,
		lastRefillTime: time.Now(),
	}
}

// TryTake 尝试获取令牌
func (tb *TokenBucket) TryTake(amount int) bool {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	// 填充令牌
	tb.refill()

	// 检查是否有足够的令牌
	if tb.tokens >= amount {
		tb.tokens -= amount
		return true
	}

	return false
}

// refill 填充令牌
func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefillTime).Seconds()

	// 计算应该生成的令牌数
	newTokens := int(elapsed * float64(tb.rate))
	if newTokens > 0 {
		tb.tokens += newTokens
		// 不能超过桶容量
		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity
		}
		tb.lastRefillTime = now
	}
}