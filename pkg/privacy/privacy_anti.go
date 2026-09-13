package privacy

import (
	"context"
	"fmt"
	protocol "github.com/hanzywaifu-coder/dongtube-meowbail/internal/protocol"
	messaging "github.com/hanzywaifu-coder/dongtube-meowbail/pkg/messaging"
	"time"

	media "github.com/hanzywaifu-coder/dongtube-meowbail/pkg/media"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

// ExtractQuotedMedia mengekstrak konten media dari pesan yang di-reply/quote
func ExtractQuotedMedia(c *protocol.Client, msg *waE2E.Message) (mediaData []byte, mediaType string, err error) {
	if msg == nil {
		return nil, "", fmt.Errorf("pesan kosong")
	}

	var qm *waE2E.Message
	if msg.ExtendedTextMessage != nil && msg.ExtendedTextMessage.ContextInfo != nil {
		qm = msg.ExtendedTextMessage.ContextInfo.QuotedMessage
	}

	if qm == nil {
		return nil, "", fmt.Errorf("tidak ada pesan quoted/reply")
	}

	if qm.ImageMessage != nil {
		data, err := media.DownloadMedia(c, context.Background(), qm.ImageMessage)
		return data, "image", err
	}
	if qm.VideoMessage != nil {
		data, err := media.DownloadMedia(c, context.Background(), qm.VideoMessage)
		return data, "video", err
	}
	if qm.AudioMessage != nil {
		data, err := media.DownloadMedia(c, context.Background(), qm.AudioMessage)
		return data, "audio", err
	}
	if qm.StickerMessage != nil {
		data, err := media.DownloadMedia(c, context.Background(), qm.StickerMessage)
		return data, "sticker", err
	}
	if qm.DocumentMessage != nil {
		data, err := media.DownloadMedia(c, context.Background(), qm.DocumentMessage)
		return data, "document", err
	}

	return nil, "", fmt.Errorf("tipe media tidak dikenali pada quoted message")
}

// ResendViewOnce meneruskan media yang dikirim sebagai View-Once ke chat/grup tujuan dalam bentuk media biasa
func ResendViewOnce(c *protocol.Client, ctx context.Context, targetChat protocol.JID, viewOnceMsg *waE2E.FutureProofMessage, caption string) error {
	if viewOnceMsg == nil || viewOnceMsg.Message == nil {
		return fmt.Errorf("pesan view-once kosong")
	}

	inner := viewOnceMsg.Message
	if inner.ImageMessage != nil {
		data, err := media.DownloadMedia(c, ctx, inner.ImageMessage)
		if err != nil {
			return err
		}
		return messaging.SendImage(c, ctx, targetChat, data, caption)
	}

	if inner.VideoMessage != nil {
		data, err := media.DownloadMedia(c, ctx, inner.VideoMessage)
		if err != nil {
			return err
		}
		return messaging.SendVideo(c, ctx, targetChat, data, caption)
	}

	if inner.AudioMessage != nil {
		data, err := media.DownloadMedia(c, ctx, inner.AudioMessage)
		if err != nil {
			return err
		}
		return messaging.SendAudio(c, ctx, targetChat, data)
	}

	return fmt.Errorf("tipe media view-once tidak didukung")
}

// SendGhostMention mengirim pesan teks yang men-tag seluruh pengguna secara tidak kasat mata (Ghost/Invisible Mention)
func SendGhostMention(c *protocol.Client, ctx context.Context, chat protocol.JID, text string, targetJIDs []string) error {
	msg := &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: proto.String(text),
			ContextInfo: &waE2E.ContextInfo{
				MentionedJID: targetJIDs,
			},
		},
	}
	_, err := c.Client.SendMessage(ctx, chat, msg)
	return err
}

// SendAutoDeleteText mengirim pesan teks yang otomatis hilang (self-destructing) setelah durasi tertentu
func SendAutoDeleteText(c *protocol.Client, ctx context.Context, chat protocol.JID, text string, duration time.Duration) error {
	resp, err := c.Client.SendMessage(ctx, chat, &waE2E.Message{
		Conversation: proto.String(text),
	})
	if err != nil {
		return err
	}

	go func(targetID protocol.MessageID) {
		time.Sleep(duration)
		_ = messaging.RevokeMessage(c, context.Background(), chat, targetID, true, protocol.EmptyJID)
	}(resp.ID)

	return nil
}

// FormatJID membersihkan string nomor telepon menjadi format WhatsApp JID standar
func FormatJID(phoneOrJID string) protocol.JID {
	if phoneOrJID == "" {
		return protocol.EmptyJID
	}
	clean := phoneOrJID
	for _, ch := range []string{"+", "-", " ", "@s.whatsapp.net", "@c.us", "@g.us"} {
		for {
			idx := len(clean) - len(ch)
			if idx >= 0 && clean[idx:] == ch {
				clean = clean[:idx]
			} else {
				break
			}
		}
	}
	return protocol.NewJID(clean, protocol.DefaultUserServer)
}
