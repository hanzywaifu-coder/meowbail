package protocol

import (
	"context"
	"fmt"
	"go.mau.fi/whatsmeow/types"
	"sync"
	"time"
)

// SocketHealthMonitor mengawasi stabilitas koneksi WebSocket WhatsApp,
// mengukur latency ping-pong round-trip time (RTT), dan mendeteksi silent-disconnect.
type SocketHealthMonitor struct {
	client      *Client
	mu          sync.RWMutex
	lastPing    time.Time
	lastPongRTT time.Duration
	failCount   int
	stopChan    chan struct{}
}

// NewSocketHealthMonitor membuat instans baru health monitor
func (c *Client) NewSocketHealthMonitor() *SocketHealthMonitor {
	return &SocketHealthMonitor{
		client:   c,
		stopChan: make(chan struct{}),
	}
}

// Ping mengukur round-trip time (RTT) koneksi ke server WhatsApp
func (m *SocketHealthMonitor) Ping(ctx context.Context) (time.Duration, error) {
	start := time.Now()
	// Kirim stanza keepalive bawaan whatsmeow
	isSuccess, _ := m.client.Client.DangerousInternals().SendKeepAlive(ctx)
	if !isSuccess {
		m.mu.Lock()
		m.failCount++
		m.mu.Unlock()
		return 0, fmt.Errorf("ping timeout / server whatsapp tidak merespons")
	}
	rtt := time.Since(start)
	m.mu.Lock()
	m.lastPing = time.Now()
	m.lastPongRTT = rtt
	m.failCount = 0
	m.mu.Unlock()
	return rtt, nil
}

// GetStats mengembalikan metrik kesehatan socket terkini
func (m *SocketHealthMonitor) GetStats() (lastPing time.Time, rtt time.Duration, fails int) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastPing, m.lastPongRTT, m.failCount
}

// PingServer memanggil uji ping langsung dari Client MeowBail
func (c *Client) PingServer(ctx context.Context) (time.Duration, error) {
	start := time.Now()
	ok, _ := c.Client.DangerousInternals().SendKeepAlive(ctx)
	if !ok {
		return 0, fmt.Errorf("ping gagal: no response from server")
	}
	return time.Since(start), nil
}

// PrewarmMediaConn menghangatkan cache koneksi MMS/media WhatsApp
// Menghindari delay 1-2 detik saat bot pertama kali mengunduh/mengunggah media
func (c *Client) PrewarmMediaConn(ctx context.Context) error {
	_, err := c.Client.DangerousInternals().RefreshMediaConn(ctx, false)
	return err
}

// CheckNumberOnWhatsApp memeriksa secara efisien apakah satu nomor telepon terdaftar di WhatsApp
func (c *Client) CheckNumberOnWhatsApp(ctx context.Context, phone string) (types.IsOnWhatsAppResponse, error) {
	res, err := c.Client.IsOnWhatsApp(ctx, []string{phone})
	if err != nil {
		return types.IsOnWhatsAppResponse{}, err
	}
	if len(res) == 0 {
		return types.IsOnWhatsAppResponse{}, fmt.Errorf("nomor tidak ditemukan")
	}
	return res[0], nil
}
