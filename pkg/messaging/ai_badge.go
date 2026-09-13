package messaging

import (
	"context"
	protocol "github.com/hanzywaifu-coder/dongtube-meowbail/internal/protocol"

	"go.mau.fi/whatsmeow/proto/waAICommon"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

// AIBadgeOptions opsi penyertaan badge Meta AI resmi pada pesan
type AIBadgeOptions struct {
	PersonaID string
	ModelName string
}

// SendTextWithAIBadge mengirim teks dengan badge AI / Meta AI resmi di WhatsApp
func SendTextWithAIBadge(c *protocol.Client, ctx context.Context, chat protocol.JID, text string, opt ...AIBadgeOptions) error {
	persona := "867051314767696$760019659443059"
	if len(opt) > 0 && opt[0].PersonaID != "" {
		persona = opt[0].PersonaID
	}

	botMetadata := &waAICommon.BotMetadata{
		PersonaID: proto.String(persona),
	}

	modelType := waAICommon.BotModelMetadata_LLAMA_PROD
	botMetadata.ModelMetadata = &waAICommon.BotModelMetadata{
		ModelType: &modelType,
	}

	msg := &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: proto.String(text),
		},
		MessageContextInfo: &waE2E.MessageContextInfo{
			BotMetadata: botMetadata,
		},
	}

	_, err := c.Client.SendMessage(ctx, chat, msg)
	return err
}
