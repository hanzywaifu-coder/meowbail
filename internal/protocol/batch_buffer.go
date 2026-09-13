package protocol

import (
	"sync"
	"time"
)

// BatchBuffer mengumpulkan item / event dalam batch untuk mengurangi overhead I/O dan rate-limit
type BatchBuffer[T any] struct {
	mu          sync.Mutex
	items       []T
	batchSize   int
	flushPeriod time.Duration
	flushFunc   func([]T)
	timer       *time.Timer
	closed      bool
}

// NewBatchBuffer membuat batch buffer thread-safe dengan ukuran batas dan interval auto-flush
func NewBatchBuffer[T any](batchSize int, flushPeriod time.Duration, flushFunc func([]T)) *BatchBuffer[T] {
	if batchSize <= 0 {
		batchSize = 50
	}
	if flushPeriod <= 0 {
		flushPeriod = 1 * time.Second
	}
	b := &BatchBuffer[T]{
		batchSize:   batchSize,
		flushPeriod: flushPeriod,
		flushFunc:   flushFunc,
	}
	return b
}

// Add memasukkan item ke dalam buffer, otomatis flush jika kapasitas batch tercapai
func (b *BatchBuffer[T]) Add(item T) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return
	}
	b.items = append(b.items, item)
	if len(b.items) >= b.batchSize {
		b.flushLocked()
		return
	}
	if b.timer == nil {
		b.timer = time.AfterFunc(b.flushPeriod, func() {
			b.mu.Lock()
			defer b.mu.Unlock()
			b.flushLocked()
		})
	}
}

// Flush menguras semua antrean item yang tersisa
func (b *BatchBuffer[T]) Flush() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.flushLocked()
}
func (b *BatchBuffer[T]) flushLocked() {
	if b.timer != nil {
		b.timer.Stop()
		b.timer = nil
	}
	if len(b.items) == 0 {
		return
	}
	batch := b.items
	b.items = nil
	if b.flushFunc != nil {
		go b.flushFunc(batch)
	}
}

// Close menutup buffer dan melakukan flush terakhir
func (b *BatchBuffer[T]) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.closed = true
	b.flushLocked()
}
