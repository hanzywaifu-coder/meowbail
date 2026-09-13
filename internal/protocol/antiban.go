package protocol

import (
	"math/rand"
	"sync"
	"time"
)

// AntiBanConfig konfigurasi proteksi akun WhatsApp dari deteksi bot otomatis Meta
type AntiBanConfig struct {
	Enabled           bool
	MinTypingDelayMs  int
	MaxTypingDelayMs  int
	RateLimitPerMin   int
	BurstLimit        int
	SimulateCircadian bool
}

// DefaultAntiBanConfig mengembalikan konfigurasi cerdas anti-banned
func DefaultAntiBanConfig() *AntiBanConfig {
	return &AntiBanConfig{
		Enabled:           true,
		MinTypingDelayMs:  800,
		MaxTypingDelayMs:  2500,
		RateLimitPerMin:   30,
		BurstLimit:        5,
		SimulateCircadian: true,
	}
}

// AntiBanEngine mengelola ritme pengiriman manusiawi (human entropy, presence delay, token bucket)
type AntiBanEngine struct {
	cfg     *AntiBanConfig
	mu      sync.Mutex
	history map[string][]time.Time
}

func NewAntiBanEngine(cfg *AntiBanConfig) *AntiBanEngine {
	if cfg == nil {
		cfg = DefaultAntiBanConfig()
	}
	return &AntiBanEngine{
		cfg:     cfg,
		history: make(map[string][]time.Time),
	}
}

// ComputeHumanDelay menghitung variasi jeda acak manusia sebelum mengirim pesan
func (a *AntiBanEngine) ComputeHumanDelay(textLength int) time.Duration {
	if !a.cfg.Enabled {
		return 0
	}
	// 1. Ketikan manusia ~30-50ms per karakter
	charDelay := textLength * (30 + rand.Intn(20))
	if charDelay < a.cfg.MinTypingDelayMs {
		charDelay = a.cfg.MinTypingDelayMs
	}
	if charDelay > a.cfg.MaxTypingDelayMs {
		charDelay = a.cfg.MaxTypingDelayMs
	}
	// 2. Tambah jitter acak (entropy)
	jitter := rand.Intn(350)
	return time.Duration(charDelay+jitter) * time.Millisecond
}

// AllowSend memeriksa apakah target penerima melebihi batas wajar frekuensi (Spam Shield)
func (a *AntiBanEngine) AllowSend(targetJID string) (allowed bool, waitTime time.Duration) {
	if !a.cfg.Enabled {
		return true, 0
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	now := time.Now()
	oneMinAgo := now.Add(-1 * time.Minute)
	times := a.history[targetJID]
	var recent []time.Time
	for _, t := range times {
		if t.After(oneMinAgo) {
			recent = append(recent, t)
		}
	}
	// Bersihkan entri lama jika kosong untuk cegah memory leak
	if len(recent) == 0 && len(times) > 0 {
		delete(a.history, targetJID)
	}
	if len(recent) >= a.cfg.RateLimitPerMin {
		// Tunggu sisa detik menuju reset
		oldest := recent[0]
		remaining := oldest.Add(1 * time.Minute).Sub(now)
		return false, remaining
	}
	recent = append(recent, now)
	a.history[targetJID] = recent
	// Periodik garbage collection history jika map tumbuh terlalu besar (>5000 entri)
	if len(a.history) > 5000 {
		for jid, h := range a.history {
			if len(h) == 0 || h[len(h)-1].Before(oneMinAgo) {
				delete(a.history, jid)
			}
		}
	}
	return true, 0
}
