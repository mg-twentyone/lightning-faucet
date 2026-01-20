package security

import (
	"fmt"
	"sync"
	"testing"
	"time"
	"faucet/app/config"
)

// test rate limiter business logic
func TestRateLimiter(t *testing.T) {
	config.RateLimitWindow = 1 // 1 second 
	rl := NewRateLimiter()
	ip := "192.168.1.1"

	allowed, _ := rl.IsAllowed(ip)
	if !allowed {
		t.Errorf("Expected first claim to be allowed")
	}

	rl.RecordClaim(ip)

	allowed, _ = rl.IsAllowed(ip)
	if allowed {
		t.Errorf("Expected second claim to be blocked")
	}

	// wait for window expiration
	time.Sleep(1100 * time.Millisecond)

	allowed, _ = rl.IsAllowed(ip)
	if !allowed {
		t.Errorf("Expected allowed after timeout")
	}
}

// test rate limit concurrency
func TestRateLimiter_Concurrency(t *testing.T) {
	rl := NewRateLimiter()
	concurrency := 100 
	
	var wg sync.WaitGroup
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func(id int) {
			defer wg.Done()
			ip := fmt.Sprintf("192.168.1.%d", id)
			rl.RecordClaim(ip)
			rl.IsAllowed(ip)
		}(i)
	}

	wg.Wait()
}