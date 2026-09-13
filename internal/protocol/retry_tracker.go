package protocol

import (
	"sync"
	"time"
)

// RetrySpiralingTracker mendeteksi spiral pengulangan pesan (retry loop) yang sering memicu pemblokiran akun
type RetrySpiralingTracker struct {
	mu           sync.Mutex
	maxRetries   int
	retryCounts  map[string]int
	lastAttempts map[string]time.Time
}

func NewRetrySpiralingTracker(maxRetries int) *RetrySpiralingTracker {
	if maxRetries <= 0 {
		maxRetries = 5
	}
	return &RetrySpiralingTracker{
		maxRetries:   maxRetries,
		retryCounts:  make(map[string]int),
		lastAttempts: make(map[string]time.Time),
	}
}

// TrackRetry mencatat percobaan pengulangan pesan dan mengembalikan true jika terdeteksi spiral
func (rt *RetrySpiralingTracker) TrackRetry(messageID string) (isSpiraling bool) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	rt.retryCounts[messageID]++
	rt.lastAttempts[messageID] = time.Now()
	return rt.retryCounts[messageID] >= rt.maxRetries
}

// MarkSuccess mereset counter saat pesan berhasil terkirim atau diakui server
func (rt *RetrySpiralingTracker) MarkSuccess(messageID string) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	delete(rt.retryCounts, messageID)
	delete(rt.lastAttempts, messageID)
}
