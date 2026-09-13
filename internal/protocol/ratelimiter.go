package protocol

import (
	"go.mau.fi/whatsmeow/types"
	"sync"
	"time"
)

// RateLimiter mengontrol frekuensi pengiriman pesan per chat untuk mencegah ban akun WA
type RateLimiter struct {
	mu           sync.Mutex
	lastSent     map[string]time.Time
	minInterval  time.Duration
	burstLimit   int
	tokenBuckets map[string]int
}

// NewRateLimiter membuat instance baru RateLimiter berbasis Token Bucket per target JID
func NewRateLimiter(minInterval time.Duration, burstLimit int) *RateLimiter {
	return &RateLimiter{
		lastSent:     make(map[string]time.Time),
		minInterval:  minInterval,
		burstLimit:   burstLimit,
		tokenBuckets: make(map[string]int),
	}
}

// GlobalRateLimiter default limiter aman untuk bot WhatsApp
var GlobalRateLimiter = NewRateLimiter(500*time.Millisecond, 5)

// WaitOrThrottle memeriksa dan menahan sejenak jika frekuensi pesan melampaui batas aman
func (r *RateLimiter) WaitOrThrottle(chat types.JID) time.Duration {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := chat.String()
	last, exists := r.lastSent[key]
	now := time.Now()
	if !exists {
		r.lastSent[key] = now
		r.tokenBuckets[key] = r.burstLimit - 1
		return 0
	}
	elapsed := now.Sub(last)
	if elapsed < r.minInterval {
		tokens := r.tokenBuckets[key]
		if tokens > 0 {
			r.tokenBuckets[key] = tokens - 1
			r.lastSent[key] = now
			return 0
		}
		waitTime := r.minInterval - elapsed
		r.lastSent[key] = now.Add(waitTime)
		return waitTime
	}
	r.tokenBuckets[key] = r.burstLimit - 1
	r.lastSent[key] = now
	// Periodik garbage collection entri jika map membesar (>3000 target chat)
	if len(r.lastSent) > 3000 {
		expireThreshold := now.Add(-10 * time.Minute)
		for k, t := range r.lastSent {
			if t.Before(expireThreshold) {
				delete(r.lastSent, k)
				delete(r.tokenBuckets, k)
			}
		}
	}
	return 0
}

// Throttle menerapkan proteksi jeda aman sebelum mengirim pesan ke chat
func (c *Client) Throttle(chat types.JID) {
	if GlobalRateLimiter != nil {
		waitTime := GlobalRateLimiter.WaitOrThrottle(chat)
		if waitTime > 0 {
			time.Sleep(waitTime)
		}
	}
}
