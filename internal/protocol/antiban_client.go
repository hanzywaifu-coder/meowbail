package protocol

import (
	"context"
	"go.mau.fi/whatsmeow/types"
	"time"
)

// SendWithAntiBan mengirim pesan dengan simulasi presence (mengetik) dan jeda alami manusia
func (c *Client) SendWithAntiBan(ctx context.Context, chat types.JID, text string, sendFn func() error) error {
	engine := c.AntiBanEngine
	if engine == nil {
		engine = NewAntiBanEngine(DefaultAntiBanConfig())
	}
	// 1. Cek rate limit token bucket
	allowed, wait := engine.AllowSend(chat.String())
	if !allowed && wait > 0 {
		time.Sleep(wait)
	}
	// 2. Simulasi composing / sedang mengetik
	_ = c.Client.SendChatPresence(ctx, chat, types.ChatPresenceComposing, types.ChatPresenceMediaText)
	// 3. Jeda dinamis berdasarkan panjang pesan
	delay := engine.ComputeHumanDelay(len(text))
	time.Sleep(delay)
	// 4. Set presence paused lalu eksekusi pengiriman
	_ = c.Client.SendChatPresence(ctx, chat, types.ChatPresencePaused, types.ChatPresenceMediaText)
	return sendFn()
}
