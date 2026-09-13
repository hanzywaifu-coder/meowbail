package protocol

import (
	"go.mau.fi/whatsmeow/types"
	"strings"
	"sync"
	"time"
)

// AntiSpamGuard mengamankan bot dari flood pesan dan trigger spam loop.
// Menggunakan algoritma Token Bucket per-chat dan per-pengguna.
type AntiSpamGuard struct {
	mu          sync.Mutex
	userBuckets map[string]*userRateInfo
	rateLimit   int           // Max request yang diizinkan dalam rentang interval
	interval    time.Duration // Jendela waktu evaluasi
}
type userRateInfo struct {
	lastSeen   time.Time
	count      int
	isCooldown bool
	cooldownTo time.Time
}

var defaultGuard = NewAntiSpamGuard(8, 5*time.Second)

// NewAntiSpamGuard membuat instans baru proteksi anti-spam
func NewAntiSpamGuard(rateLimit int, interval time.Duration) *AntiSpamGuard {
	guard := &AntiSpamGuard{
		userBuckets: make(map[string]*userRateInfo),
		rateLimit:   rateLimit,
		interval:    interval,
	}
	// Cleaner loop untuk menghapus memori user yang sudah tidak aktif
	go func() {
		ticker := time.NewTicker(2 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			guard.mu.Lock()
			now := time.Now()
			for k, v := range guard.userBuckets {
				if now.Sub(v.lastSeen) > 5*time.Minute {
					delete(guard.userBuckets, k)
				}
			}
			guard.mu.Unlock()
		}
	}()
	return guard
}

// Allow mengecek apakah suatu request diperbolehkan atau harus di-throttle
func (g *AntiSpamGuard) Allow(userKey string) (allowed bool, cooldownRemaining time.Duration) {
	g.mu.Lock()
	defer g.mu.Unlock()
	now := time.Now()
	info, exists := g.userBuckets[userKey]
	if !exists {
		g.userBuckets[userKey] = &userRateInfo{
			lastSeen: now,
			count:    1,
		}
		return true, 0
	}
	// Cek jika sedang dalam masa cooldown
	if info.isCooldown {
		if now.Before(info.cooldownTo) {
			return false, info.cooldownTo.Sub(now)
		}
		// Reset cooldown
		info.isCooldown = false
		info.count = 1
		info.lastSeen = now
		return true, 0
	}
	// Evaluasi jendela waktu
	if now.Sub(info.lastSeen) > g.interval {
		info.count = 1
		info.lastSeen = now
		return true, 0
	}
	info.count++
	info.lastSeen = now
	if info.count > g.rateLimit {
		info.isCooldown = true
		info.cooldownTo = now.Add(10 * time.Second)
		return false, 10 * time.Second
	}
	return true, 0
}

// CheckSpam memeriksa pesan dari client
func (c *Client) CheckSpam(sender types.JID) (bool, time.Duration) {
	key := sender.User
	if key == "" {
		key = sender.String()
	}
	return defaultGuard.Allow(key)
}

// ParseJIDHelper mengonversi input nomor/string menjadi types.JID yang valid
func ParseJIDHelper(target string) (types.JID, error) {
	target = strings.TrimSpace(target)
	target = strings.TrimPrefix(target, "@")
	target = strings.ReplaceAll(target, "+", "")
	target = strings.ReplaceAll(target, "-", "")
	target = strings.ReplaceAll(target, " ", "")
	if strings.Contains(target, "@") {
		return types.ParseJID(target)
	}
	if len(target) >= 18 && strings.HasPrefix(target, "120363") {
		// Asumsi Group JID atau Newsletter JID jika panjang
		return types.NewJID(target, types.GroupServer), nil
	}
	return types.NewJID(target, types.DefaultUserServer), nil
}
