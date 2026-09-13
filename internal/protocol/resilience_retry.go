package protocol

import (
	"context"
	"errors"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"strings"
	"time"
)

// SendMessageWithRetry membungkus SendMessage dengan strategi exponential backoff
// Menangani transient connection drop atau frame header EOF secara otomatis tanpa crash
func (c *Client) SendMessageWithRetry(ctx context.Context, to types.JID, msg *waE2E.Message, maxRetries int, extra ...whatsmeow.SendRequestExtra) (whatsmeow.SendResponse, error) {
	if maxRetries <= 0 {
		maxRetries = 3
	}
	var reqExtra whatsmeow.SendRequestExtra
	if len(extra) > 0 {
		reqExtra = extra[0]
	}
	var lastErr error
	backoff := 250 * time.Millisecond
	for attempt := 1; attempt <= maxRetries; attempt++ {
		resp, err := c.Client.SendMessage(ctx, to, msg, reqExtra)
		if err == nil {
			return resp, nil
		}
		lastErr = err
		errStr := strings.ToLower(err.Error())
		// Deteksi apakah error bersifat transient network/socket issue
		isTransient := strings.Contains(errStr, "eof") ||
			strings.Contains(errStr, "timeout") ||
			strings.Contains(errStr, "connection reset") ||
			strings.Contains(errStr, "broken pipe") ||
			strings.Contains(errStr, "websocket") ||
			strings.Contains(errStr, "server closed")
		if !isTransient || attempt == maxRetries {
			break
		}
		select {
		case <-ctx.Done():
			return whatsmeow.SendResponse{}, ctx.Err()
		case <-time.After(backoff):
			backoff *= 2
		}
	}
	return whatsmeow.SendResponse{}, errors.New("meowbail: failed after retries: " + lastErr.Error())
}
