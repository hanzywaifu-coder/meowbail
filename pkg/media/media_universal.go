package media

import (
	"context"
	"fmt"
	protocol "github.com/hanzywaifu-coder/dongtube-meowbail/internal/protocol"

	"go.mau.fi/whatsmeow/proto/waE2E"
)

// UnwrapMessage membongkar semua jenis wrapper pesan WhatsApp (Ephemeral, ViewOnce v1/v2/ext, DocWithCaption)
func UnwrapMessage(msg *waE2E.Message) *waE2E.Message {
	if msg == nil {
		return nil
	}
	targetMsg := msg
	for {
		changed := false
		if targetMsg.EphemeralMessage != nil && targetMsg.EphemeralMessage.Message != nil {
			targetMsg = targetMsg.EphemeralMessage.Message
			changed = true
		}
		if targetMsg.ViewOnceMessage != nil && targetMsg.ViewOnceMessage.Message != nil {
			targetMsg = targetMsg.ViewOnceMessage.Message
			changed = true
		}
		if targetMsg.ViewOnceMessageV2 != nil && targetMsg.ViewOnceMessageV2.Message != nil {
			targetMsg = targetMsg.ViewOnceMessageV2.Message
			changed = true
		}
		if targetMsg.ViewOnceMessageV2Extension != nil && targetMsg.ViewOnceMessageV2Extension.Message != nil {
			targetMsg = targetMsg.ViewOnceMessageV2Extension.Message
			changed = true
		}
		if targetMsg.DocumentWithCaptionMessage != nil && targetMsg.DocumentWithCaptionMessage.Message != nil {
			targetMsg = targetMsg.DocumentWithCaptionMessage.Message
			changed = true
		}
		if !changed {
			break
		}
	}
	return targetMsg
}

// DownloadAnyMedia otomatis mendeteksi dan mengunduh media dari waE2E.Message (mendukung ViewOnce, Quoted, Image, Video, Audio, Sticker, Doc)
func DownloadAnyMedia(c *protocol.Client, ctx context.Context, msg *waE2E.Message) ([]byte, string, string, error) {
	if msg == nil {
		return nil, "", "", fmt.Errorf("pesan kosong")
	}

	// 1. Unwrap Ephemeral / ViewOnce jika ada
	targetMsg := UnwrapMessage(msg)

	// 2. Cek tipe media langsung
	if targetMsg.ImageMessage != nil {
		data, err := c.Client.Download(ctx, targetMsg.ImageMessage)
		return data, "image/jpeg", "image.jpg", err
	}
	if targetMsg.VideoMessage != nil {
		data, err := c.Client.Download(ctx, targetMsg.VideoMessage)
		return data, "video/mp4", "video.mp4", err
	}
	if targetMsg.AudioMessage != nil {
		data, err := c.Client.Download(ctx, targetMsg.AudioMessage)
		return data, targetMsg.AudioMessage.GetMimetype(), "audio.mp3", err
	}
	if targetMsg.StickerMessage != nil {
		data, err := c.Client.Download(ctx, targetMsg.StickerMessage)
		return data, "image/webp", "sticker.webp", err
	}
	if targetMsg.DocumentMessage != nil {
		data, err := c.Client.Download(ctx, targetMsg.DocumentMessage)
		filename := targetMsg.DocumentMessage.GetFileName()
		if filename == "" {
			filename = "document.bin"
		}
		return data, targetMsg.DocumentMessage.GetMimetype(), filename, err
	}

	// 3. Cek quoted/reply media jika pesan utama tidak memiliki media langsung
	var ctxInfo *waE2E.ContextInfo
	if targetMsg.ExtendedTextMessage != nil && targetMsg.ExtendedTextMessage.ContextInfo != nil {
		ctxInfo = targetMsg.ExtendedTextMessage.ContextInfo
	}

	if ctxInfo == nil && targetMsg.ImageMessage != nil {
		ctxInfo = targetMsg.ImageMessage.ContextInfo
	}
	if ctxInfo == nil && targetMsg.VideoMessage != nil {
		ctxInfo = targetMsg.VideoMessage.ContextInfo
	}
	if ctxInfo == nil && targetMsg.AudioMessage != nil {
		ctxInfo = targetMsg.AudioMessage.ContextInfo
	}
	if ctxInfo == nil && targetMsg.StickerMessage != nil {
		ctxInfo = targetMsg.StickerMessage.ContextInfo
	}
	if ctxInfo == nil && targetMsg.DocumentMessage != nil {
		ctxInfo = targetMsg.DocumentMessage.ContextInfo
	}

	if ctxInfo != nil && ctxInfo.QuotedMessage != nil {
		qm := UnwrapMessage(ctxInfo.QuotedMessage)

		if qm.ImageMessage != nil {
			data, err := c.Client.Download(ctx, qm.ImageMessage)
			return data, "image/jpeg", "image.jpg", err
		}
		if qm.VideoMessage != nil {
			data, err := c.Client.Download(ctx, qm.VideoMessage)
			return data, "video/mp4", "video.mp4", err
		}
		if qm.AudioMessage != nil {
			data, err := c.Client.Download(ctx, qm.AudioMessage)
			return data, qm.AudioMessage.GetMimetype(), "audio.mp3", err
		}
		if qm.StickerMessage != nil {
			data, err := c.Client.Download(ctx, qm.StickerMessage)
			return data, "image/webp", "sticker.webp", err
		}
		if qm.DocumentMessage != nil {
			data, err := c.Client.Download(ctx, qm.DocumentMessage)
			filename := qm.DocumentMessage.GetFileName()
			if filename == "" {
				filename = "document.bin"
			}
			return data, qm.DocumentMessage.GetMimetype(), filename, err
		}
	}

	return nil, "", "", fmt.Errorf("tidak ada media yang dapat diunduh dalam pesan")
}
