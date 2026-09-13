package protocol

import (
	"go.mau.fi/whatsmeow/types"
	"sync"
	"time"
)

// MessageDeduplicator mencegah eksekusi duplikat dari pesan WhatsApp yang sama
// Terutama saat reconnect atau push notification ganda dari server WhatsApp
type MessageDeduplicator struct {
	mu      sync.RWMutex
	seen    map[string]time.Time
	order   []string
	maxSize int
	ttl     time.Duration
}

var (
	GlobalDeduplicator     *MessageDeduplicator
	globalDeduplicatorOnce sync.Once
)

// GetGlobalDeduplicator mengembalikan instance deduplicator global
func GetGlobalDeduplicator() *MessageDeduplicator {
	globalDeduplicatorOnce.Do(func() {
		// Default: kapasitas 5000 ID pesan terakhir, TTL 10 menit
		GlobalDeduplicator = NewMessageDeduplicator(5000, 10*time.Minute)
	})
	return GlobalDeduplicator
}

// NewMessageDeduplicator membuat instance deduplicator baru
func NewMessageDeduplicator(maxSize int, ttl time.Duration) *MessageDeduplicator {
	if maxSize <= 0 {
		maxSize = 2500
	}
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return &MessageDeduplicator{
		seen:    make(map[string]time.Time),
		order:   make([]string, 0, maxSize),
		maxSize: maxSize,
		ttl:     ttl,
	}
}

// CheckAndMark mengembalikan true jika pesan SUDAH PERNAH diproses (duplikat)
// Jika belum pernah, otomatis mencatat ID tersebut dan mengembalikan false
func (d *MessageDeduplicator) CheckAndMark(msgID types.MessageID) bool {
	id := string(msgID)
	if id == "" {
		return false
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	now := time.Now()
	// Jika sudah ada dan belum expired -> duplikat
	if seenAt, exists := d.seen[id]; exists {
		if now.Sub(seenAt) < d.ttl {
			return true
		}
		// Sudah expired, perbarui timestamp
		d.seen[id] = now
		return false
	}
	// Jika antrean penuh, lakukan ring-buffer eviction yang hemat memori
	if len(d.order) >= d.maxSize {
		oldest := d.order[0]
		copy(d.order, d.order[1:])
		d.order = d.order[:len(d.order)-1]
		delete(d.seen, oldest)
	}
	d.order = append(d.order, id)
	d.seen[id] = now
	return false
}

// Stats mengembalikan jumlah ID pesan yang sedang di-cache
func (d *MessageDeduplicator) Stats() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.seen)
}
