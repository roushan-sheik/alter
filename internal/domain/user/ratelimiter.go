package user

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// rateLimitEntry wraps a rate.Limiter with a last-seen timestamp for eviction.
type rateLimitEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter holds per-email token-bucket limiters for auth endpoints.
// It is safe for concurrent use and evicts stale entries every 5 minutes.
type RateLimiter struct {
	mu           sync.Mutex
	loginMap     map[string]*rateLimitEntry
	forgotMap    map[string]*rateLimitEntry
	loginRate    rate.Limit
	loginBurst   int
	forgotRate   rate.Limit
	forgotBurst  int
	resendMap    map[string]*rateLimitEntry
	resendRate   rate.Limit
	resendBurst  int
}

// newRateLimiter creates a RateLimiter and starts the background eviction goroutine.
//
//   - login:         5 requests per minute, burst 5
//   - forgotPassword: 3 requests per 10 minutes, burst 3
func newRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		loginMap:    make(map[string]*rateLimitEntry),
		forgotMap:   make(map[string]*rateLimitEntry),
		loginRate:   rate.Every(time.Minute / 5),   // 5 req/min
		loginBurst:  5,
		forgotRate:  rate.Every(10 * time.Minute / 3), // 3 req/10min
		forgotBurst: 3,
		resendMap:   make(map[string]*rateLimitEntry),
		resendRate:  rate.Every(1 * time.Minute),      // 1 req/min
		resendBurst: 1,
	}
	go rl.evictLoop()
	return rl
}

// AllowLogin returns true when the given email is within the login rate limit.
func (rl *RateLimiter) AllowLogin(email string) bool {
	return rl.allow(rl.loginMap, email, rl.loginRate, rl.loginBurst)
}

// AllowForgotPassword returns true when the given email is within the forgot-password rate limit.
func (rl *RateLimiter) AllowForgotPassword(email string) bool {
	return rl.allow(rl.forgotMap, email, rl.forgotRate, rl.forgotBurst)
}

// AllowResendOTP returns true when the given email is within the resend rate limit (1 per min).
func (rl *RateLimiter) AllowResendOTP(email string) bool {
	return rl.allow(rl.resendMap, email, rl.resendRate, rl.resendBurst)
}

func (rl *RateLimiter) allow(m map[string]*rateLimitEntry, key string, r rate.Limit, burst int) bool {
	rl.mu.Lock()
	entry, ok := m[key]
	if !ok {
		entry = &rateLimitEntry{limiter: rate.NewLimiter(r, burst)}
		m[key] = entry
	}
	entry.lastSeen = time.Now()
	allow := entry.limiter.Allow()
	rl.mu.Unlock()
	return allow
}

// evictLoop removes entries that have not been seen for more than 10 minutes.
// Runs every 5 minutes in the background.
func (rl *RateLimiter) evictLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.evict()
	}
}

func (rl *RateLimiter) evict() {
	cutoff := time.Now().Add(-10 * time.Minute)
	rl.mu.Lock()
	defer rl.mu.Unlock()
	for k, v := range rl.loginMap {
		if v.lastSeen.Before(cutoff) {
			delete(rl.loginMap, k)
		}
	}
	for k, v := range rl.forgotMap {
		if v.lastSeen.Before(cutoff) {
			delete(rl.forgotMap, k)
		}
	}
	for k, v := range rl.resendMap {
		if v.lastSeen.Before(cutoff) {
			delete(rl.resendMap, k)
		}
	}
}
