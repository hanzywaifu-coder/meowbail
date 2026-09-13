package privacy

import (
	"context"
	protocol "github.com/hanzywaifu-coder/dongtube-meowbail/internal/protocol"
	"time"
)

// SimulateTyping menyimulasikan bot sedang mengetik teks selama durasi tertentu
func SimulateTyping(c *protocol.Client, ctx context.Context, chat protocol.JID, duration time.Duration) error {
	_ = c.Client.SendChatPresence(ctx, chat, protocol.ChatPresenceComposing, protocol.ChatPresenceMediaText)
	select {
	case <-time.After(duration):
	case <-ctx.Done():
		_ = c.Client.SendChatPresence(context.Background(), chat, protocol.ChatPresencePaused, protocol.ChatPresenceMediaText)
		return ctx.Err()
	}
	return c.Client.SendChatPresence(ctx, chat, protocol.ChatPresencePaused, protocol.ChatPresenceMediaText)
}

// SimulateRecording menyimulasikan bot sedang merekam audio/voice note selama durasi tertentu
func SimulateRecording(c *protocol.Client, ctx context.Context, chat protocol.JID, duration time.Duration) error {
	_ = c.Client.SendChatPresence(ctx, chat, protocol.ChatPresenceComposing, protocol.ChatPresenceMediaAudio)
	select {
	case <-time.After(duration):
	case <-ctx.Done():
		_ = c.Client.SendChatPresence(context.Background(), chat, protocol.ChatPresencePaused, protocol.ChatPresenceMediaAudio)
		return ctx.Err()
	}
	return c.Client.SendChatPresence(ctx, chat, protocol.ChatPresencePaused, protocol.ChatPresenceMediaAudio)
}

// MarkReadSimple menandai pesan-pesan terakhir dalam obrolan sebagai sudah dibaca (Read Receipts)
func MarkReadSimple(c *protocol.Client, ctx context.Context, chat protocol.JID, messageIDs []protocol.MessageID) error {
	return c.Client.MarkRead(ctx, messageIDs, time.Now(), chat, protocol.EmptyJID)
}

// MarkVoicePlayed menandai voice note atau audio yang diterima sebagai sudah didengarkan (mic biru menyala)
// Parity dengan Baileys readMessages([key], 'played')
func MarkVoicePlayed(c *protocol.Client, ctx context.Context, chat protocol.JID, sender protocol.JID, messageIDs []protocol.MessageID) error {
	return c.Client.MarkRead(ctx, messageIDs, time.Now(), chat, sender, protocol.ReceiptTypePlayed)
}

// SetBotPresence mengatur status online bot (Available / Unavailable)
func SetBotPresence(c *protocol.Client, ctx context.Context, isAvailable bool) error {
	state := protocol.PresenceUnavailable
	if isAvailable {
		state = protocol.PresenceAvailable
	}
	return c.Client.SendPresence(ctx, state)
}

// SubscribePresence meminta server WhatsApp mengirimkan update presensi/online/lastseen dari kontak tertentu
// Parity dengan Baileys presenceSubscribe(toJid)
func SubscribePresence(c *protocol.Client, ctx context.Context, targetJID protocol.JID) error {
	return c.Client.SubscribePresence(ctx, targetJID)
}
