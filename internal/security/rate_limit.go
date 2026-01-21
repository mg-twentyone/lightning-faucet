package security

import (
	"sync"
	"time"
	"faucet/internal/configs"
)

type RateLimiter struct {
	claims map[string]time.Time
	mu     sync.RWMutex
}

func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		claims: make(map[string]time.Time),
	}
	// janitor for resources management
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) IsAllowed(ip string) (bool, time.Duration) {
	rl.mu.RLock()
	lastClaim, exists := rl.claims[ip]
	rl.mu.RUnlock()

	if exists {
		window := time.Duration(configs.RateLimitWindow) * time.Second
		elapsed := time.Since(lastClaim)
		if elapsed < window {
			return false, window - elapsed
		}
	}
	return true, 0
}

func (rl *RateLimiter) RecordClaim(ip string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.claims[ip] = time.Now()
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		for ip, lastClaim := range rl.claims {
			if time.Since(lastClaim).Seconds() > float64(configs.RateLimitWindow) {
				delete(rl.claims, ip)
			}
		}
		rl.mu.Unlock()
	}
}