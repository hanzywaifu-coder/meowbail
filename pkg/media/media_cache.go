package media

import (
	"container/list"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"go.mau.fi/whatsmeow"
)

// MediaUploadCacheEntry menyimpan metadata hasil upload WhatsApp
type MediaUploadCacheEntry struct {
	Response  whatsmeow.UploadResponse
	CreatedAt time.Time
	MediaType whatsmeow.MediaType
}

// MediaUploadLRUCache mencegah re-upload media berulang (misal thumbnail menu, audio, stiker, gambar yang sama)
// Menghemat hingga 95% kuota data dan latensi network ke server WhatsApp CDN
type MediaUploadLRUCache struct {
	capacity  int
	ttl       time.Duration
	items     map[string]*list.Element
	evictList *list.List
	mu        sync.RWMutex
}

type lruItem struct {
	key   string
	entry MediaUploadCacheEntry
}

var (
	globalUploadCache     *MediaUploadLRUCache
	globalUploadCacheOnce sync.Once
)

// GetGlobalUploadCache mengembalikan instance upload cache global
func GetGlobalUploadCache() *MediaUploadLRUCache {
	globalUploadCacheOnce.Do(func() {
		// Default: kapasitas 1000 item media terunggah, TTL 12 jam (sesuai lifecycle WhatsApp media server token)
		globalUploadCache = NewMediaUploadLRUCache(1000, 12*time.Hour)
	})
	return globalUploadCache
}

func NewMediaUploadLRUCache(capacity int, ttl time.Duration) *MediaUploadLRUCache {
	if capacity <= 0 {
		capacity = 500
	}
	if ttl <= 0 {
		ttl = 12 * time.Hour
	}
	return &MediaUploadLRUCache{
		capacity:  capacity,
		ttl:       ttl,
		items:     make(map[string]*list.Element),
		evictList: list.New(),
	}
}

func (c *MediaUploadLRUCache) cacheKey(mediaType whatsmeow.MediaType, data []byte) string {
	h := sha256.New()
	h.Write(data)
	return fmt.Sprintf("%s:%s", mediaType, hex.EncodeToString(h.Sum(nil)))
}

// Get mengambil hasil upload ter-cache jika masih valid dalam rentang TTL
func (c *MediaUploadLRUCache) Get(mediaType whatsmeow.MediaType, data []byte) (whatsmeow.UploadResponse, bool) {
	if c == nil || len(data) == 0 {
		return whatsmeow.UploadResponse{}, false
	}
	key := c.cacheKey(mediaType, data)

	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		item := elem.Value.(*lruItem)
		if time.Since(item.entry.CreatedAt) > c.ttl {
			c.evictList.Remove(elem)
			delete(c.items, key)
			return whatsmeow.UploadResponse{}, false
		}
		c.evictList.MoveToFront(elem)
		return item.entry.Response, true
	}
	return whatsmeow.UploadResponse{}, false
}

// Put menyimpan hasil upload ke cache
func (c *MediaUploadLRUCache) Put(mediaType whatsmeow.MediaType, data []byte, resp whatsmeow.UploadResponse) {
	if c == nil || len(data) == 0 || resp.DirectPath == "" {
		return
	}
	key := c.cacheKey(mediaType, data)

	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		c.evictList.MoveToFront(elem)
		item := elem.Value.(*lruItem)
		item.entry.Response = resp
		item.entry.CreatedAt = time.Now()
		return
	}

	elem := c.evictList.PushFront(&lruItem{
		key: key,
		entry: MediaUploadCacheEntry{
			Response:  resp,
			CreatedAt: time.Now(),
			MediaType: mediaType,
		},
	})
	c.items[key] = elem

	for c.evictList.Len() > c.capacity {
		oldest := c.evictList.Back()
		if oldest != nil {
			c.evictList.Remove(oldest)
			kv := oldest.Value.(*lruItem)
			delete(c.items, kv.key)
		}
	}
}

// Stats mengembalikan jumlah item ter-cache
func (c *MediaUploadLRUCache) Stats() int {
	if c == nil {
		return 0
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}
