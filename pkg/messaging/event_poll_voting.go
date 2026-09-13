package messaging

import (
	"context"
	protocol "github.com/hanzywaifu-coder/dongtube-meowbail/internal/protocol"
	"time"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

// SendInteractiveEvent mengirim undangan acara kalender / Event Message interaktif
func SendInteractiveEvent(c *protocol.Client, ctx context.Context, chat protocol.JID, name, description, locationName string, startTime, endTime time.Duration, isCanceled bool) error {
	now := time.Now()
	startTS := now.Add(startTime).Unix()
	endTS := now.Add(endTime).Unix()

	msg := &waE2E.Message{
		EventMessage: &waE2E.EventMessage{
			IsCanceled:  proto.Bool(isCanceled),
			Name:        proto.String(name),
			Description: proto.String(description),
			Location: &waE2E.LocationMessage{
				Name: proto.String(locationName),
			},
			StartTime: proto.Int64(startTS),
			EndTime:   proto.Int64(endTS),
		},
	}

	_, err := c.Client.SendMessage(ctx, chat, msg)
	return err
}

// SendReplyWithQuote mengirim balasan dengan mengutip (quote) pesan secara presisi
func SendReplyWithQuote(c *protocol.Client, ctx context.Context, chat protocol.JID, text string, quotedMsgID protocol.MessageID, quotedSender protocol.JID, quotedMessage *waE2E.Message) error {
	if quotedMessage == nil {
		quotedMessage = &waE2E.Message{
			Conversation: proto.String(""),
		}
	}

	participantStr := quotedSender.ToNonAD().String()
	if participantStr == "" {
		participantStr = chat.ToNonAD().String()
	}

	msg := &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: proto.String(text),
			ContextInfo: &waE2E.ContextInfo{
				StanzaID:      proto.String(string(quotedMsgID)),
				Participant:   proto.String(participantStr),
				QuotedMessage: quotedMessage,
			},
		},
	}

	_, err := c.Client.SendMessage(ctx, chat, msg)
	return err
}
