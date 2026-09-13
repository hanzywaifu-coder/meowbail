package protocol

import (
	"context"
	"math"
	"sync"
	"time"
)

// AutoReconnectManager mengelola auto-reconnect cerdas dengan Exponential Backoff + Jitter + Circuit Breaker
type AutoReconnectManager struct {
	client       *Client
	mu           sync.Mutex
	running      bool
	stopChan     chan struct{}
	retryCount   int
	maxRetries   int
	initialDelay time.Duration
	maxDelay     time.Duration
	jitterFactor float64
	// Circuit breaker
	consecutiveFailures int
	circuitOpen         bool
	circuitOpenTime     time.Time
	circuitThreshold    int           // failures before opening circuit
	circuitTimeout      time.Duration // how long to keep circuit open
	lastSuccessTime     time.Time
}

func NewAutoReconnectManager(client *Client) *AutoReconnectManager {
	return &AutoReconnectManager{
		client:           client,
		stopChan:         make(chan struct{}),
		maxRetries:       50,
		initialDelay:     2 * time.Second,
		maxDelay:         60 * time.Second,
		jitterFactor:     0.3,             // 30% jitter
		circuitThreshold: 5,               // open circuit after 5 consecutive failures
		circuitTimeout:   2 * time.Minute, // keep circuit open for 2 minutes
	}
}
func (m *AutoReconnectManager) Start(ctx context.Context) {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return
	}
	m.running = true
	m.mu.Unlock()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// Cegah crash process bila terjadi panic tak terduga dalam loop reconnect
			}
		}()
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-m.stopChan:
				return
			case <-ticker.C:
				if m.client == nil || m.client.Client == nil {
					continue
				}
				// Hanya reconnect jika perangkat sudah terdaftar/logged in tapi koneksi terputus
				if m.client.IsLoggedIn() && !m.client.IsConnected() {
					m.attemptReconnect(ctx)
				} else if m.client.IsConnected() {
					m.onConnectionRestored()
				}
			}
		}
	}()
}
func (m *AutoReconnectManager) onConnectionRestored() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.retryCount = 0
	m.consecutiveFailures = 0
	m.circuitOpen = false
	m.lastSuccessTime = time.Now()
}
func (m *AutoReconnectManager) attemptReconnect(ctx context.Context) {
	m.mu.Lock()
	// Check circuit breaker
	if m.circuitOpen {
		if time.Since(m.circuitOpenTime) > m.circuitTimeout {
			// Half-open: allow one attempt
			m.circuitOpen = false
		} else {
			m.mu.Unlock()
			return // circuit open, skip this attempt
		}
	}
	if m.retryCount >= m.maxRetries {
		m.mu.Unlock()
		return
	}
	m.retryCount++
	attempt := m.retryCount
	m.mu.Unlock()
	// Exponential backoff with jitter: initialDelay * 2^(attempt-1) * (1 ± jitter)
	backoff := float64(m.initialDelay) * math.Pow(1.8, float64(attempt-1))
	jitter := backoff * m.jitterFactor * (2*float64(time.Now().UnixNano()%1000)/1000.0 - 1)
	delay := time.Duration(backoff + jitter)
	if delay > m.maxDelay {
		delay = m.maxDelay
	}
	select {
	case <-time.After(delay):
	case <-ctx.Done():
		return
	case <-m.stopChan:
		return
	}
	if !m.client.IsConnected() {
		err := m.client.Connect(ctx)
		m.recordAttemptResult(err == nil)
	}
}
func (m *AutoReconnectManager) recordAttemptResult(success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if success {
		m.consecutiveFailures = 0
		m.circuitOpen = false
		m.lastSuccessTime = time.Now()
	} else {
		m.consecutiveFailures++
		if m.consecutiveFailures >= m.circuitThreshold && !m.circuitOpen {
			m.circuitOpen = true
			m.circuitOpenTime = time.Now()
		}
	}
}
func (m *AutoReconnectManager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.running {
		return
	}
	m.running = false
	close(m.stopChan)
}

// GetStats mengembalikan statistik reconnect untuk monitoring
func (m *AutoReconnectManager) GetStats() map[string]interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()
	return map[string]interface{}{
		"running":              m.running,
		"retry_count":          m.retryCount,
		"max_retries":          m.maxRetries,
		"consecutive_failures": m.consecutiveFailures,
		"circuit_open":         m.circuitOpen,
		"circuit_open_time":    m.circuitOpenTime,
		"circuit_threshold":    m.circuitThreshold,
		"last_success_time":    m.lastSuccessTime,
	}
}
func (c *Client) EnableAutoReconnect(ctx context.Context) *AutoReconnectManager {
	mgr := NewAutoReconnectManager(c)
	mgr.Start(ctx)
	return mgr
}
func (c *Client) SafeConnect(ctx context.Context) error {
	err := c.Connect(ctx)
	if c.Config() != nil && c.Config().AutoReconnect {
		c.EnableAutoReconnect(ctx)
	}
	return err
}
