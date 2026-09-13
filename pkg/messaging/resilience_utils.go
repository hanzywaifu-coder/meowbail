package messaging

import (
	"context"
	"fmt"
	protocol "github.com/hanzywaifu-coder/dongtube-meowbail/internal/protocol"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
	"strings"
	"sync"
	"time"
)

// FloodRateLimiter pencegah spam pesan keluar ke WhatsApp server
type FloodRateLimiter struct {
	chatTimes map[string][]time.Time
	maxBurst  int
	window    time.Duration
	mu        sync.Mutex
}

// NewFloodRateLimiter inisialisasi rate limiter anti-spam
func NewFloodRateLimiter(maxBurst int, window time.Duration) *FloodRateLimiter {
	return &FloodRateLimiter{
		chatTimes: make(map[string][]time.Time),
		maxBurst:  maxBurst,
		window:    window,
	}
}

// Allow mengecek apakah pesan diizinkan dikirim ke target obrolan
func (l *FloodRateLimiter) Allow(chat string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	times, exists := l.chatTimes[chat]
	if !exists {
		l.chatTimes[chat] = []time.Time{now}
		return true
	}
	var valid []time.Time
	for _, t := range times {
		if now.Sub(t) < l.window {
			valid = append(valid, t)
		}
	}
	if len(valid) >= l.maxBurst {
		l.chatTimes[chat] = valid
		return false
	}
	valid = append(valid, now)
	l.chatTimes[chat] = valid
	return true
}

// AutoTypingContext mengirim status typing secara natural sebelum pesan terkirim
func AutoTypingContext(c *protocol.Client, ctx context.Context, chat protocol.JID, duration time.Duration) {
	_ = c.Client.SendChatPresence(ctx, chat, protocol.ChatPresenceComposing, protocol.ChatPresenceMediaText)
	time.Sleep(duration)
	_ = c.Client.SendChatPresence(ctx, chat, protocol.ChatPresencePaused, protocol.ChatPresenceMediaText)
}

// SendTextClean mengirim teks sederhana dengan konteks keamanan default
func SendTextClean(c *protocol.Client, ctx context.Context, chat protocol.JID, text string) error {
	msg := &waE2E.Message{
		Conversation: proto.String(text),
	}
	_, err := c.Client.SendMessage(ctx, chat, msg)
	return err
}

// ParseCommandArgs membedah baris teks menjadi nama command dan argumen
func ParseCommandArgs(rawText, prefix string) (command string, args []string, isCmd bool) {
	clean := strings.TrimSpace(rawText)
	if prefix != "" {
		if !strings.HasPrefix(clean, prefix) {
			return "", nil, false
		}
		clean = strings.TrimPrefix(clean, prefix)
	}
	clean = strings.TrimSpace(clean)
	if clean == "" {
		return "", nil, false
	}
	fields := strings.Fields(clean)
	return strings.ToLower(fields[0]), fields[1:], true
}

// SafeBroadcast mengirim pesan ke daftar obrolan dengan jeda berkala anti-banned
func SafeBroadcast(c *protocol.Client, ctx context.Context, targets []protocol.JID, text string, delayPerChat time.Duration) (sent int, failed int) {
	if delayPerChat <= 0 {
		delayPerChat = 1200 * time.Millisecond
	}
	for _, t := range targets {
		select {
		case <-ctx.Done():
			return sent, failed
		default:
		}
		err := SendText(c, ctx, t, text)
		if err != nil {
			failed++
		} else {
			sent++
		}
		time.Sleep(delayPerChat)
	}
	return sent, failed
}

// FormatUptime memformat durasi waktu menjadi string hari, jam, menit, detik
func FormatUptime(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	mins := int(d.Minutes()) % 60
	secs := int(d.Seconds()) % 60
	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, mins)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, mins, secs)
	}
	return fmt.Sprintf("%dm %ds", mins, secs)
}
