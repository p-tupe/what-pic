package main

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	rateWindow = time.Hour
	rateLimit  = 5 // images per window
)

type ipEntry struct {
	count       int
	windowStart time.Time
}

type rateLimiter struct {
	mu         sync.Mutex
	entries    map[string]*ipEntry
	trustProxy bool // honour X-Forwarded-For only when behind a known proxy
}

func newRateLimiter(trustProxy bool) *rateLimiter {
	rl := &rateLimiter{
		entries:    make(map[string]*ipEntry),
		trustProxy: trustProxy,
	}
	go rl.cleanup()
	return rl
}

// allow checks whether ip can consume n images. Returns (allowed, remaining).
func (rl *rateLimiter) allow(ip string, n int) (bool, int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	e, ok := rl.entries[ip]
	if !ok || now.Sub(e.windowStart) >= rateWindow {
		rl.entries[ip] = &ipEntry{count: n, windowStart: now}
		return true, rateLimit - n
	}
	if e.count+n > rateLimit {
		return false, rateLimit - e.count
	}
	e.count += n
	return true, rateLimit - e.count
}

func (rl *rateLimiter) cleanup() {
	t := time.NewTicker(10 * time.Minute)
	defer t.Stop()
	for range t.C {
		rl.mu.Lock()
		cutoff := time.Now().Add(-rateWindow)
		for ip, e := range rl.entries {
			if e.windowStart.Before(cutoff) {
				delete(rl.entries, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *rateLimiter) realIP(r *http.Request) string {
	if rl.trustProxy {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			return strings.TrimSpace(strings.SplitN(xff, ",", 2)[0])
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
