package messaging

import (
	"context"
	"fmt"
	protocol "github.com/hanzywaifu-coder/dongtube-meowbail/internal/protocol"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
	"strings"
)

// NormalizeMessage membuka semua wrapper WhatsApp seperti Ephemeral, ViewOnce, Edited, DocumentWithCaption, dan LottieSticker
func NormalizeMessage(msg *waE2E.Message) *waE2E.Message {
	if msg == nil {
		return nil
	}
	for {
		if msg.EphemeralMessage != nil && msg.EphemeralMessage.Message != nil {
			msg = msg.EphemeralMessage.Message
			continue
		}
		if msg.ViewOnceMessage != nil && msg.ViewOnceMessage.Message != nil {
			msg = msg.ViewOnceMessage.Message
			continue
		}
		if msg.ViewOnceMessageV2 != nil && msg.ViewOnceMessageV2.Message != nil {
			msg = msg.ViewOnceMessageV2.Message
			continue
		}
		if msg.ViewOnceMessageV2Extension != nil && msg.ViewOnceMessageV2Extension.Message != nil {
			msg = msg.ViewOnceMessageV2Extension.Message
			continue
		}
		if msg.DocumentWithCaptionMessage != nil && msg.DocumentWithCaptionMessage.Message != nil {
			msg = msg.DocumentWithCaptionMessage.Message
			continue
		}
		if msg.EditedMessage != nil && msg.EditedMessage.Message != nil && msg.EditedMessage.Message.ProtocolMessage != nil && msg.EditedMessage.Message.ProtocolMessage.EditedMessage != nil {
			msg = msg.EditedMessage.Message.ProtocolMessage.EditedMessage
			continue
		}
		if msg.LottieStickerMessage != nil && msg.LottieStickerMessage.Message != nil {
			msg = msg.LottieStickerMessage.Message
			continue
		}
		break
	}
	return msg
}

// ExtractText mengekstrak pesan teks dari berbagai varian payload WhatsApp
func ExtractText(msg *waE2E.Message) string {
	msg = NormalizeMessage(msg)
	if msg == nil {
		return ""
	}
	if msg.Conversation != nil {
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
	if msg.ButtonsResponseMessage != nil && msg.ButtonsResponseMessage.SelectedButtonID != nil {
		return *msg.ButtonsResponseMessage.SelectedButtonID
	}
	if msg.ListResponseMessage != nil && msg.ListResponseMessage.SingleSelectReply != nil && msg.ListResponseMessage.SingleSelectReply.SelectedRowID != nil {
		return *msg.ListResponseMessage.SingleSelectReply.SelectedRowID
	}
	if msg.TemplateButtonReplyMessage != nil && msg.TemplateButtonReplyMessage.SelectedID != nil {
		return *msg.TemplateButtonReplyMessage.SelectedID
	}
	if msg.InteractiveResponseMessage != nil {
		if native := msg.InteractiveResponseMessage.GetNativeFlowResponseMessage(); native != nil && native.ParamsJSON != nil {
			return *native.ParamsJSON
		}
		if body := msg.InteractiveResponseMessage.GetBody(); body != nil && body.Text != nil {
			return *body.Text
		}
	}
	if msg.PollCreationMessage != nil && msg.PollCreationMessage.Name != nil {
		return *msg.PollCreationMessage.Name
	}
	if msg.PollCreationMessageV2 != nil && msg.PollCreationMessageV2.Name != nil {
		return *msg.PollCreationMessageV2.Name
	}
	if msg.PollCreationMessageV3 != nil && msg.PollCreationMessageV3.Name != nil {
		return *msg.PollCreationMessageV3.Name
	}
	if msg.EventMessage != nil && msg.EventMessage.Name != nil {
		return *msg.EventMessage.Name
	}
	if msg.ViewOnceMessage != nil && msg.ViewOnceMessage.Message != nil {
		return ExtractText(msg.ViewOnceMessage.Message)
	}
	if msg.ViewOnceMessageV2 != nil && msg.ViewOnceMessageV2.Message != nil {
		return ExtractText(msg.ViewOnceMessageV2.Message)
	}
	if msg.ViewOnceMessageV2Extension != nil && msg.ViewOnceMessageV2Extension.Message != nil {
		return ExtractText(msg.ViewOnceMessageV2Extension.Message)
	}
	if msg.DocumentWithCaptionMessage != nil && msg.DocumentWithCaptionMessage.Message != nil {
		return ExtractText(msg.DocumentWithCaptionMessage.Message)
	}
	if msg.EditedMessage != nil && msg.EditedMessage.Message != nil {
		return ExtractText(msg.EditedMessage.Message)
	}
	return ""
}

// ExtractSender mengekstrak protocol.JID pengirim secara aman dari grup maupun DM pribadi
func ExtractSender(evt *protocol.MessageEvent) protocol.JID {
	if evt == nil {
		return protocol.EmptyJID
	}
	if evt.IsGroup {
		return evt.Sender.ToNonAD()
	}
	return evt.Chat.ToNonAD()
}

// SendSilentReply mengirim balasan tanpa suara notifikasi pada obrolan (Silent Message Notification)
func SendSilentReply(c *protocol.Client, ctx context.Context, chat protocol.JID, text string) error {
	msg := &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: proto.String(text),
		},
	}
	_, err := c.Client.SendMessage(ctx, chat, msg)
	return err
}

// SendAutoReaction otomatis memberi reaksi emoji ke pesan masuk
func SendAutoReaction(c *protocol.Client, ctx context.Context, chat protocol.JID, targetID protocol.MessageID, emoji string) error {
	return SendReaction(c, ctx, chat, targetID, emoji)
}

// CleanNumber membersihkan format nomor internasional menjadi digit bersih
func CleanNumber(phone string) string {
	var sb strings.Builder
	for _, ch := range phone {
		if ch >= '0' && ch <= '9' {
			sb.WriteRune(ch)
		}
	}
	return sb.String()
}

// FormatMentionList membentuk format string @user1 @user2 dan slice string jid untuk tag massal
func FormatMentionList(jids []protocol.JID) (text string, mentions []string) {
	var tags []string
	for _, j := range jids {
		user := j.User
		if user != "" {
			tags = append(tags, fmt.Sprintf("@%s", user))
			mentions = append(mentions, j.ToNonAD().String())
		}
	}
	return strings.Join(tags, " "), mentions
}

// GetDevice memprediksi tipe perangkat pengirim pesan (ios, android, web, desktop) berdasarkan ID pesan
func GetDevice(id string) string {
	if len(id) >= 21 && (strings.HasPrefix(id, "3EB0") || strings.HasPrefix(id, "3E")) {
		return "web"
	}
	if len(id) == 20 && strings.HasPrefix(id, "3A") {
		return "ios"
	}
	if len(id) == 21 || len(id) == 32 {
		return "android"
	}
	if strings.HasPrefix(id, "3F") || len(id) == 18 {
		return "desktop"
	}
	return "unknown"
}

// IsRealMessage memvalidasi apakah payload merupakan pesan konten nyata yang ditujukan untuk pengguna (bukan protokol sinkronisasi/reaksi semata)
func IsRealMessage(msg *waE2E.Message) bool {
	if msg == nil {
		return false
	}
	if msg.ProtocolMessage != nil || msg.ReactionMessage != nil || msg.EncReactionMessage != nil || msg.PollUpdateMessage != nil {
		return false
	}
	return msg.Conversation != nil ||
		msg.ExtendedTextMessage != nil ||
		msg.ImageMessage != nil ||
		msg.VideoMessage != nil ||
		msg.AudioMessage != nil ||
		msg.DocumentMessage != nil ||
		msg.StickerMessage != nil ||
		msg.ContactMessage != nil ||
		msg.LocationMessage != nil ||
		msg.LiveLocationMessage != nil ||
		msg.ButtonsResponseMessage != nil ||
		msg.ListResponseMessage != nil ||
		msg.TemplateButtonReplyMessage != nil ||
		msg.InteractiveResponseMessage != nil ||
		msg.PollCreationMessage != nil ||
		msg.PollCreationMessageV2 != nil ||
		msg.PollCreationMessageV3 != nil ||
		msg.EventMessage != nil ||
		msg.ViewOnceMessage != nil ||
		msg.ViewOnceMessageV2 != nil ||
		msg.DocumentWithCaptionMessage != nil ||
		msg.EditedMessage != nil
}
