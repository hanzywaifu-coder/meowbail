package messaging

import (
	"context"
	"fmt"
	protocol "github.com/hanzywaifu-coder/dongtube-meowbail/internal/protocol"
	"time"

	media "github.com/hanzywaifu-coder/dongtube-meowbail/pkg/media"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waCommon"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

// ViewOnceType tipe media view once
type ViewOnceType string

const (
	ViewOnceImage ViewOnceType = "image"
	ViewOnceVideo ViewOnceType = "video"
	ViewOnceAudio ViewOnceType = "audio"
)

// SendViewOnce mengirim media satu kali lihat (ViewOnce V2)
func SendViewOnce(c *protocol.Client, ctx context.Context, chat protocol.JID, data []byte, mediaType ViewOnceType, caption string) error {
	var innerMsg *waE2E.Message

	switch mediaType {
	case ViewOnceImage:
		uploaded, err := media.UploadMedia(c, ctx, data, whatsmeow.MediaImage)
		if err != nil {
			return fmt.Errorf("upload image: %w", err)
		}
		innerMsg = &waE2E.Message{
			ImageMessage: &waE2E.ImageMessage{
				URL:           proto.String(uploaded.URL),
				DirectPath:    proto.String(uploaded.DirectPath),
				MediaKey:      uploaded.MediaKey,
				Mimetype:      proto.String("image/jpeg"),
				FileEncSHA256: uploaded.FileEncSHA256,
				FileSHA256:    uploaded.FileSHA256,
				FileLength:    proto.Uint64(uploaded.FileLength),
				Caption:       proto.String(caption),
			},
		}
	case ViewOnceVideo:
		uploaded, err := media.UploadMedia(c, ctx, data, whatsmeow.MediaVideo)
		if err != nil {
			return fmt.Errorf("upload video: %w", err)
		}
		innerMsg = &waE2E.Message{
			VideoMessage: &waE2E.VideoMessage{
				URL:           proto.String(uploaded.URL),
				DirectPath:    proto.String(uploaded.DirectPath),
				MediaKey:      uploaded.MediaKey,
				Mimetype:      proto.String("video/mp4"),
				FileEncSHA256: uploaded.FileEncSHA256,
				FileSHA256:    uploaded.FileSHA256,
				FileLength:    proto.Uint64(uploaded.FileLength),
				Caption:       proto.String(caption),
			},
		}
	case ViewOnceAudio:
		uploaded, err := media.UploadMedia(c, ctx, data, whatsmeow.MediaAudio)
		if err != nil {
			return fmt.Errorf("upload audio: %w", err)
		}
		ptt := true
		innerMsg = &waE2E.Message{
			AudioMessage: &waE2E.AudioMessage{
				URL:           proto.String(uploaded.URL),
				DirectPath:    proto.String(uploaded.DirectPath),
				MediaKey:      uploaded.MediaKey,
				Mimetype:      proto.String("audio/ogg; codecs=opus"),
				FileEncSHA256: uploaded.FileEncSHA256,
				FileSHA256:    uploaded.FileSHA256,
				FileLength:    proto.Uint64(uploaded.FileLength),
				PTT:           &ptt,
			},
		}
	default:
		return fmt.Errorf("unsupported view once type: %s", mediaType)
	}

	msg := &waE2E.Message{
		ViewOnceMessage: &waE2E.FutureProofMessage{
			Message: innerMsg,
		},
	}

	_, err := c.Client.SendMessage(ctx, chat, msg)
	return err
}

// EditMessage mengubah pesan teks yang telah dikirim sebelumnya (WhatsApp Message Edit Protocol)
func EditMessage(c *protocol.Client, ctx context.Context, chat protocol.JID, targetMsgID protocol.MessageID, newText string) error {
	editMsg := &waE2E.Message{
		ProtocolMessage: &waE2E.ProtocolMessage{
			Key: &waCommon.MessageKey{
				RemoteJID: proto.String(chat.String()),
				FromMe:    proto.Bool(true),
				ID:        proto.String(string(targetMsgID)),
			},
			Type: waE2E.ProtocolMessage_MESSAGE_EDIT.Enum(),
			EditedMessage: &waE2E.Message{
				Conversation: proto.String(newText),
			},
			TimestampMS: proto.Int64(time.Now().UnixMilli()),
		},
	}

	_, err := c.Client.SendMessage(ctx, chat, editMsg)
	return err
}

// RevokeMessage menghapus pesan untuk semua orang (Delete for Everyone)
func RevokeMessage(c *protocol.Client, ctx context.Context, chat protocol.JID, targetMsgID protocol.MessageID, fromMe bool, sender protocol.JID) error {
	key := &waCommon.MessageKey{
		RemoteJID: proto.String(chat.String()),
		FromMe:    proto.Bool(fromMe),
		ID:        proto.String(string(targetMsgID)),
	}
	if !fromMe && sender.User != "" {
		key.Participant = proto.String(sender.ToNonAD().String())
	}

	revokeMsg := &waE2E.Message{
		ProtocolMessage: &waE2E.ProtocolMessage{
			Key:  key,
			Type: waE2E.ProtocolMessage_REVOKE.Enum(),
		},
	}

	_, err := c.Client.SendMessage(ctx, chat, revokeMsg)
	return err
}

// SendPaymentRequest mengirim invoice / payment request Native Flow (WhatsApp Pay format)
func SendPaymentRequest(c *protocol.Client, ctx context.Context, chat protocol.JID, amountINR int64, currency string, note string, expiry time.Duration) error {
	if currency == "" {
		currency = "IDR"
	}
	if expiry == 0 {
		expiry = 24 * time.Hour
	}

	msg := &waE2E.Message{
		RequestPaymentMessage: &waE2E.RequestPaymentMessage{
			NoteMessage: &waE2E.Message{
				ExtendedTextMessage: &waE2E.ExtendedTextMessage{
					Text: proto.String(note),
				},
			},
			CurrencyCodeIso4217: proto.String(currency),
			Amount1000:          proto.Uint64(uint64(amountINR * 1000)),
			ExpiryTimestamp:     proto.Int64(time.Now().Add(expiry).Unix()),
		},
	}

	_, err := c.Client.SendMessage(ctx, chat, msg)
	return err
}
