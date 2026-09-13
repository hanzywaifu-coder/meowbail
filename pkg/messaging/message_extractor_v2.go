package messaging

import (
	"context"
	"fmt"
	protocol "github.com/hanzywaifu-coder/dongtube-meowbail/internal/protocol"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
	"strings"
)

// ExtractPureText mengekstrak string pesan murni dari berbagai struktur payload WhatsApp
func ExtractPureText(msg *waE2E.Message) string {
	if msg == nil {
		return ""
	}
	if msg.EphemeralMessage != nil && msg.EphemeralMessage.Message != nil {
		msg = msg.EphemeralMessage.Message
	}
	if msg.ViewOnceMessage != nil && msg.ViewOnceMessage.Message != nil {
		msg = msg.ViewOnceMessage.Message
	}
	if msg.ViewOnceMessageV2 != nil && msg.ViewOnceMessageV2.Message != nil {
		msg = msg.ViewOnceMessageV2.Message
	}
	if msg.DocumentWithCaptionMessage != nil && msg.DocumentWithCaptionMessage.Message != nil {
		msg = msg.DocumentWithCaptionMessage.Message
	}
	if msg.Conversation != nil && *msg.Conversation != "" {
		return *msg.Conversation
	}
	if msg.ExtendedTextMessage != nil && msg.ExtendedTextMessage.Text != nil {
		return *msg.ExtendedTextMessage.Text
	}
	if msg.ImageMessage != nil && msg.ImageMessage.Caption != nil {
		return *msg.ImageMessage.Caption
	}
	if msg.VideoMessage != nil && msg.VideoMessage.Caption != nil {
		return *msg.VideoMessage.Caption
	}
	if msg.DocumentMessage != nil && msg.DocumentMessage.Caption != nil {
		return *msg.DocumentMessage.Caption
	}
	return ""
}

// ExtractQuotedMessage mengambil QuotedMessage dari berbagai lapisan ContextInfo
func ExtractQuotedMessage(msg *waE2E.Message) (*waE2E.Message, *waE2E.ContextInfo) {
	if msg == nil {
		return nil, nil
	}
	target := msg
	if target.EphemeralMessage != nil && target.EphemeralMessage.Message != nil {
		target = target.EphemeralMessage.Message
	}
	if target.ViewOnceMessage != nil && target.ViewOnceMessage.Message != nil {
		target = target.ViewOnceMessage.Message
	}
	if target.ViewOnceMessageV2 != nil && target.ViewOnceMessageV2.Message != nil {
		target = target.ViewOnceMessageV2.Message
	}
	if target.DocumentWithCaptionMessage != nil && target.DocumentWithCaptionMessage.Message != nil {
		target = target.DocumentWithCaptionMessage.Message
	}
	var ctxInfo *waE2E.ContextInfo
	if target.ExtendedTextMessage != nil {
		ctxInfo = target.ExtendedTextMessage.ContextInfo
	} else if target.ImageMessage != nil {
		ctxInfo = target.ImageMessage.ContextInfo
	} else if target.VideoMessage != nil {
		ctxInfo = target.VideoMessage.ContextInfo
	} else if target.AudioMessage != nil {
		ctxInfo = target.AudioMessage.ContextInfo
	} else if target.StickerMessage != nil {
		ctxInfo = target.StickerMessage.ContextInfo
	} else if target.DocumentMessage != nil {
		ctxInfo = target.DocumentMessage.ContextInfo
	}
	if ctxInfo != nil && ctxInfo.QuotedMessage != nil {
		qm := ctxInfo.QuotedMessage
		if qm.EphemeralMessage != nil && qm.EphemeralMessage.Message != nil {
			qm = qm.EphemeralMessage.Message
		}
		if qm.ViewOnceMessage != nil && qm.ViewOnceMessage.Message != nil {
			qm = qm.ViewOnceMessage.Message
		}
		if qm.ViewOnceMessageV2 != nil && qm.ViewOnceMessageV2.Message != nil {
			qm = qm.ViewOnceMessageV2.Message
		}
		return qm, ctxInfo
	}
	return nil, nil
}

// SendMention mengirim teks dengan mention otomatis ke sejumlah protocol.JID dan mewarisi fake reply default
func SendMention(c *protocol.Client, ctx context.Context, chat protocol.JID, text string, mentions []protocol.JID, quotedContext ...*waE2E.ContextInfo) error {
	var mentionStrings []string
	for _, m := range mentions {
		norm := m.ToNonAD()
		if norm.Server == protocol.HiddenUserServer || norm.Server == "lid" {
			if c.LIDResolver != nil {
				resolved := c.LIDResolver.ResolveToPN(norm)
				if resolved.Server == protocol.DefaultUserServer {
					norm = resolved
				}
			}
		}
		clean := strings.TrimSuffix(norm.User, "@s.whatsapp.net")
		if clean != "" {
			mentionStrings = append(mentionStrings, norm.String())
		}
	}
	var contextInfo *waE2E.ContextInfo
	if len(quotedContext) > 0 && quotedContext[0] != nil {
		contextInfo = proto.Clone(quotedContext[0]).(*waE2E.ContextInfo)
	} else if c.Config() != nil && c.Config().DefaultFakeReply != nil {
		contextInfo = proto.Clone(c.Config().DefaultFakeReply).(*waE2E.ContextInfo)
	} else {
		contextInfo = &waE2E.ContextInfo{}
	}
	contextInfo.MentionedJID = mentionStrings
	msg := &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text:        proto.String(text),
			ContextInfo: contextInfo,
		},
	}
	_, err := c.Client.SendMessage(ctx, chat, msg)
	return err
}

// ExtractMentions mengurai @nomor dari teks dan mengembalikannya sebagai slice protocol.JID
func ExtractMentions(text string) []protocol.JID {
	var results []protocol.JID
	seen := make(map[string]bool)
	fields := strings.Fields(text)
	for _, word := range fields {
		if strings.HasPrefix(word, "@") {
			rawNum := strings.TrimPrefix(word, "@")
			rawNum = strings.Trim(rawNum, ".,!?~:;()[]{}")
			if len(rawNum) >= 6 && len(rawNum) <= 16 {
				allDigits := true
				for _, r := range rawNum {
					if r < '0' || r > '9' {
						allDigits = false
						break
					}
				}
				if allDigits && !seen[rawNum] {
					seen[rawNum] = true
					results = append(results, protocol.NewJID(rawNum, protocol.DefaultUserServer))
				}
			}
		}
	}
	return results
}

// ReplyQuick mengirim teks reply cepat yang me-referensi pesan asli pengguna
func ReplyQuick(c *protocol.Client, ctx context.Context, chat protocol.JID, text string, evt *protocol.MessageEvent) error {
	if evt == nil || evt.Message == nil {
		return SendText(c, ctx, chat, text)
	}
	ctxInfo := &waE2E.ContextInfo{
		StanzaID:      proto.String(evt.Message.Info.ID),
		Participant:   proto.String(evt.Message.Info.Sender.ToNonAD().String()),
		QuotedMessage: evt.Message.Message,
	}
	msg := &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text:        proto.String(text),
			ContextInfo: ctxInfo,
		},
	}
	_, err := c.Client.SendMessage(ctx, chat, msg)
	if err != nil {
		return fmt.Errorf("reply quick failed: %w", err)
	}
	return nil
}
