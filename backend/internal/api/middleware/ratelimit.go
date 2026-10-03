package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type ipBucket struct {
	tokens     int
	lastRefill time.Time
}

// RateLimiter implements a thread-safe token bucket rate limiter by client IP.
type RateLimiter struct {
	mu          sync.Mutex
	rate        int           // tokens added per refill window
	burst       int           // max capacity
	window      time.Duration // refill window duration
	buckets     map[string]*ipBucket
	stopJanitor chan struct{}
}

// NewRateLimiter creates a new RateLimiter instance.
// rate: number of requests allowed per window.
// burst: maximum burst capacity.
// window: time duration for token replenishment.
func NewRateLimiter(rate, burst int, window time.Duration) *RateLimiter {
	if rate <= 0 {
		rate = 120
	}
	if burst <= 0 {
		burst = rate
	}
	if window <= 0 {
		window = time.Minute
	}

	rl := &RateLimiter{
		rate:        rate,
		burst:       burst,
		window:      window,
		buckets:     make(map[string]*ipBucket),
		stopJanitor: make(chan struct{}),
	}

	// Janitor to prune stale buckets every 5 minutes
	go rl.janitor(5 * time.Minute)

	return rl
}

// Close stops background janitor goroutine.
func (rl *RateLimiter) Close() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	if rl.stopJanitor != nil {
		close(rl.stopJanitor)
		rl.stopJanitor = nil
	}
}

func (rl *RateLimiter) janitor(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			now := time.Now()
			for ip, b := range rl.buckets {
				if now.Sub(b.lastRefill) > 2*rl.window {
					delete(rl.buckets, ip)
				}
			}
			rl.mu.Unlock()
		case <-rl.stopJanitor:
			return
		}
	}
}

// extractIP determines the real client IP from headers or RemoteAddr.
func extractIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		parts := strings.Split(xff, ",")
		clientIP := strings.TrimSpace(parts[0])
		if clientIP != "" {
			return clientIP
		}
	}

	xRealIP := r.Header.Get("X-Real-IP")
	if xRealIP != "" {
		return strings.TrimSpace(xRealIP)
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// Handler returns Chi-compatible middleware enforcing rate limits.
func (rl *RateLimiter) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Bypass rate limiting for healthcheck
		if r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}

		ip := extractIP(r)
		now := time.Now()

		rl.mu.Lock()
		b, exists := rl.buckets[ip]
		if !exists {
			b = &ipBucket{
				tokens:     rl.burst,
				lastRefill: now,
			}
			rl.buckets[ip] = b
		}

		// Refill tokens based on elapsed windows
		elapsed := now.Sub(b.lastRefill)
		if elapsed >= rl.window {
			windowsPassed := int(elapsed / rl.window)
			b.tokens = b.tokens + (windowsPassed * rl.rate)
			if b.tokens > rl.burst {
				b.tokens = rl.burst
			}
			b.lastRefill = now
		}

		allowed := false
		remaining := b.tokens
		resetSeconds := int((rl.window - elapsed%rl.window).Seconds())
		if resetSeconds <= 0 {
			resetSeconds = int(rl.window.Seconds())
		}

		if b.tokens > 0 {
			b.tokens--
			remaining = b.tokens
			allowed = true
		}
		rl.mu.Unlock()

		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.burst))
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", now.Add(time.Duration(resetSeconds)*time.Second).Unix()))

		if !allowed {
			w.Header().Set("Retry-After", fmt.Sprintf("%d", resetSeconds))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(fmt.Sprintf(`{"error":"rate_limit_exceeded","message":"Too many requests. Please retry after %d seconds"}`, resetSeconds)))
			return
		}

		next.ServeHTTP(w, r)
	})
}
