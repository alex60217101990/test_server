package ticker_service

import (
	"context"
	"log"
	"math/rand"
	"runtime/debug"
	"time"

	"github.com/alex60217101990/test_server/internal/cache"
)

var src rand.Source

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6
	letterIdxMask = 1<<letterIdxBits - 1
	letterIdxMax  = 63 / letterIdxBits
)

func init() {
	src = rand.NewSource(time.Now().UnixNano())
}

type Ticker struct {
	cache  cache.CacheService
	logger *log.Logger
}

func NewTicker(logger *log.Logger) TickerService {
	ticker := &Ticker{
		cache:  &cache.MemoryCache{},
		logger: logger,
	}
	ticker.cache.Connect(logger)
	return ticker
}

func (t *Ticker) Loop(ctx context.Context, secondInterval int) {
	go func() {
		interval := time.Duration(secondInterval) * time.Second
		ticker := time.NewTicker(interval)
		defer func() {
			if r := recover(); r != nil {
				if t.logger != nil {
					t.logger.Printf("package: 'ticker_service', type: 'Ticker', method: 'Loop', fatal: %v, stack: %s", r, string(debug.Stack()))
				} else {
					log.Printf("package: 'ticker_service', type: 'Ticker', method: 'Loop', fatal: %v, stack: %s", r, string(debug.Stack()))
				}
			}
			ticker.Stop()
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case tickTime, ok := <-ticker.C:
				if ok {
					randStr := randStringBytesMaskImprSrc(6)
					if err := t.cache.SetItem(tickTime.Format("2006-01-02_10:04:05"), randStr); err != nil {
						if t.logger != nil {
							t.logger.Printf("package: 'ticker_service', type: 'Ticker', method: 'Loop', error: %v, stack: %s", err, string(debug.Stack()))
						} else {
							log.Printf("package: 'ticker_service', type: 'Ticker', method: 'Loop', error: %v, stack: %s", err, string(debug.Stack()))
						}
					}
				}
			}
		}
	}()
}

func (t *Ticker) GetLatestValues() ([]string, error) {
	return t.cache.GetItems()
}

func randStringBytesMaskImprSrc(n int) string {
	b := make([]byte, n)
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return string(b)
}
