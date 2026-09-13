package protocol

import (
	"go.mau.fi/whatsmeow/types"
	"strings"
	"sync"
	"time"
)

// LIDResolver memelihara mapping dua arah antara JID Nomor Telepon (PN / s.whatsapp.net)
// dan JID Identitas Terhubung (LID / lid) untuk mencegah error dekripsi ("Bad MAC" / Session Mismatch).
type LIDResolver struct {
	mu      sync.RWMutex
	lidToPN map[string]lidPNEntry // "123456789@lid" -> {pn: "6283143961588@s.whatsapp.net", expires: time.Time}
	pnToLID map[string]lidPNEntry // "6283143961588@s.whatsapp.net" -> {pn: "123456789@lid", expires: time.Time}
	ttl     time.Duration
	stopCh  chan struct{}
}
type lidPNEntry struct {
	jid     string
	expires time.Time
}

func NewLIDResolver() *LIDResolver {
	lr := &LIDResolver{
		lidToPN: make(map[string]lidPNEntry),
		pnToLID: make(map[string]lidPNEntry),
		ttl:     24 * time.Hour, // 24 jam TTL
		stopCh:  make(chan struct{}),
	}
	// Jalankan periodic cleanup tiap 1 jam
	go lr.cleanupLoop()
	return lr
}

// RegisterMapping mendaftarkan pasangan JID LID dan Phone Number
func (lr *LIDResolver) RegisterMapping(lid, pn string) {
	if lid == "" || pn == "" {
		return
	}
	lidParsed, errL := types.ParseJID(strings.TrimSpace(lid))
	pnParsed, errP := types.ParseJID(strings.TrimSpace(pn))
	var lidNorm, pnNorm string
	if errL == nil {
		lidNorm = strings.ToLower(lidParsed.ToNonAD().String())
	} else {
		lidNorm = strings.ToLower(strings.TrimSpace(lid))
	}
	if errP == nil {
		pnNorm = strings.ToLower(pnParsed.ToNonAD().String())
	} else {
		pnNorm = strings.ToLower(strings.TrimSpace(pn))
	}
	now := time.Now()
	expires := now.Add(lr.ttl)
	lr.mu.Lock()
	defer lr.mu.Unlock()
	lr.lidToPN[lidNorm] = lidPNEntry{jid: pnNorm, expires: expires}
	lr.pnToLID[pnNorm] = lidPNEntry{jid: lidNorm, expires: expires}
}

// ResolveToPN mengembalikan Phone Number JID jika input berupa LID JID
func (lr *LIDResolver) ResolveToPN(jid types.JID) types.JID {
	if jid.Server != types.HiddenUserServer && jid.Server != types.HostedLIDServer && jid.Server != "lid" {
		return jid
	}
	normKey := strings.ToLower(jid.ToNonAD().String())
	lr.mu.RLock()
	entry, exists := lr.lidToPN[normKey]
	lr.mu.RUnlock()
	if exists && time.Now().Before(entry.expires) {
		parsed, err := types.ParseJID(entry.jid)
		if err == nil {
			return parsed
		}
	}
	return jid
}

// ResolveToLID mengembalikan LID JID jika input berupa Phone Number JID
func (lr *LIDResolver) ResolveToLID(jid types.JID) types.JID {
	if jid.Server == types.HiddenUserServer || jid.Server == types.HostedLIDServer || jid.Server == "lid" {
		return jid
	}
	normKey := strings.ToLower(jid.ToNonAD().String())
	lr.mu.RLock()
	entry, exists := lr.pnToLID[normKey]
	lr.mu.RUnlock()
	if exists && time.Now().Before(entry.expires) {
		parsed, err := types.ParseJID(entry.jid)
		if err == nil {
			return parsed
		}
	}
	return jid
}

// Stats mengembalikan statistik cache LID
func (lr *LIDResolver) Stats() (lidCount, pnCount int, expired int) {
	now := time.Now()
	lr.mu.RLock()
	defer lr.mu.RUnlock()
	lidCount = len(lr.lidToPN)
	pnCount = len(lr.pnToLID)
	for _, e := range lr.lidToPN {
		if now.After(e.expires) {
			expired++
		}
	}
	return
}

// Stop menghentikan cleanup loop
func (lr *LIDResolver) Stop() {
	close(lr.stopCh)
}

// cleanupLoop membersihkan entry yang expired secara berkala
func (lr *LIDResolver) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			lr.cleanup()
		case <-lr.stopCh:
			return
		}
	}
}
func (lr *LIDResolver) cleanup() {
	now := time.Now()
	lr.mu.Lock()
	defer lr.mu.Unlock()
	for k, e := range lr.lidToPN {
		if now.After(e.expires) {
			delete(lr.lidToPN, k)
		}
	}
	for k, e := range lr.pnToLID {
		if now.After(e.expires) {
			delete(lr.pnToLID, k)
		}
	}
}
