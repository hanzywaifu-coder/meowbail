package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	protocol "github.com/hanzywaifu-coder/dongtube-meowbail/internal/protocol"

	media "github.com/hanzywaifu-coder/dongtube-meowbail/pkg/media"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waCommon"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

// SendText sends a plain text message
func SendText(c *protocol.Client, ctx context.Context, chat protocol.JID, text string, opts ...*protocol.MessageOptions) error {
	c.Throttle(chat)
	_, err := SendTextMessage(c, ctx, chat, text, opts...)
	return err
}

// SendTextMessage sends a plain text message and returns whatsmeow SendResponse
func SendTextMessage(c *protocol.Client, ctx context.Context, chat protocol.JID, text string, opts ...*protocol.MessageOptions) (whatsmeow.SendResponse, error) {
	c.Throttle(chat)
	var ctxInfo *waE2E.ContextInfo
	if c.Config() != nil && c.Config().DefaultFakeReply != nil {
		ctxInfo = proto.Clone(c.Config().DefaultFakeReply).(*waE2E.ContextInfo)
	}

	var msg *waE2E.Message
	if ctxInfo != nil {
		if len(ctxInfo.MentionedJID) == 0 {
			mentions := ExtractMentions(text)
			if len(mentions) > 0 {
				var mList []string
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
					mList = append(mList, norm.String())
				}
				ctxInfo.MentionedJID = mList
			}
		}
		msg = &waE2E.Message{
			ExtendedTextMessage: &waE2E.ExtendedTextMessage{
				Text:        proto.String(text),
				ContextInfo: ctxInfo,
			},
		}
	} else {
		msg = &waE2E.Message{
			Conversation: &text,
		}
	}

	if len(opts) > 0 && opts[0] != nil {
		msg = applyOptions(msg, opts[0])
	}

	return c.Client.SendMessage(ctx, chat, msg)
}

// SendMentionAll mengirim pesan teks yang me-mention seluruh anggota grup sekaligus
// menggunakan field NonJIDMentions (protokol Baileys v7 mentionAll: 1)
func SendMentionAll(c *protocol.Client, ctx context.Context, chat protocol.JID, text string) error {
	c.Throttle(chat)
	ctxInfo := BuildNewsletterContext(c.Config())
	if ctxInfo == nil {
		ctxInfo = &waE2E.ContextInfo{}
	}
	ctxInfo.NonJIDMentions = proto.Uint32(1)

	msg := &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text:        proto.String(text),
			ContextInfo: ctxInfo,
		},
	}
	_, err := c.Client.SendMessage(ctx, chat, msg)
	return err
}
func SendTextWithNewsletter(c *protocol.Client, ctx context.Context, chat protocol.JID, text string) error {
	c.Throttle(chat)
	msg := &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text:        proto.String(text),
			ContextInfo: BuildNewsletterContext(c.Config()),
		},
	}

	_, err := c.Client.SendMessage(ctx, chat, msg)
	return err
}

// WrapDocument sends a document message with thumbnail + caption + newsletter context
// This matches Baileys wrapDocument() behavior from the Node.js Dongtube Bot
func WrapDocument(c *protocol.Client, ctx context.Context, chat protocol.JID, caption string, docData []byte, fileName string, thumbData []byte) error {
	c.Throttle(chat)
	resp, err := c.Client.Upload(ctx, docData, whatsmeow.MediaDocument)
	if err != nil {
		// Fallback: send as text with newsletter
		return SendTextWithNewsletter(c, ctx, chat, caption)
	}

	ctxInfo := BuildNewsletterContext(c.Config())
	if ctxInfo == nil {
		ctxInfo = &waE2E.ContextInfo{}
	}
	if len(ctxInfo.MentionedJID) == 0 {
		mentions := ExtractMentions(caption)
		if len(mentions) > 0 {
			var mList []string
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
				mList = append(mList, norm.String())
			}
			ctxInfo.MentionedJID = mList
		}
	}

	docLen := uint64(len(docData))
	msg := &waE2E.Message{
		DocumentMessage: &waE2E.DocumentMessage{
			URL:           &resp.URL,
			Mimetype:      proto.String("image/png"),
			FileName:      proto.String(fileName),
			Caption:       proto.String(caption),
			FileLength:    &docLen,
			PageCount:     proto.Uint32(100),
			FileEncSHA256: resp.FileEncSHA256,
			FileSHA256:    resp.FileSHA256,
			DirectPath:    &resp.DirectPath,
			MediaKey:      resp.MediaKey,
			JPEGThumbnail: thumbData,
			ContextInfo:   ctxInfo,
		},
	}

	_, err = c.Client.SendMessage(ctx, chat, msg)
	return err
}

// SendButtons sends a message with quick_reply buttons
func SendButtons(c *protocol.Client, ctx context.Context, chat protocol.JID, text string, buttons []protocol.Button, opts ...*protocol.MessageOptions) error {
	c.Throttle(chat)
	if len(buttons) == 0 {
		return SendText(c, ctx, chat, text, opts...)
	}

	var waButtons []*waE2E.ButtonsMessage_Button
	for i, btn := range buttons {
		switch btn.Type {
		case protocol.ButtonQuickReply:
			waButtons = append(waButtons, &waE2E.ButtonsMessage_Button{
				ButtonID: proto.String(btn.ID),
				ButtonText: &waE2E.ButtonsMessage_Button_ButtonText{
					DisplayText: proto.String(btn.Text),
				},
				Type: waE2E.ButtonsMessage_Button_RESPONSE.Enum(),
			})
		case protocol.ButtonCTAURL:
			params, _ := json.Marshal(map[string]interface{}{
				"display_text": btn.DisplayText,
				"url":          btn.URL,
				"merchant_url": btn.URL,
			})
			waButtons = append(waButtons, &waE2E.ButtonsMessage_Button{
				ButtonID: proto.String(fmt.Sprintf("cta_%d", i)),
				ButtonText: &waE2E.ButtonsMessage_Button_ButtonText{
					DisplayText: proto.String(btn.Text),
				},
				Type: waE2E.ButtonsMessage_Button_NATIVE_FLOW.Enum(),
				NativeFlowInfo: &waE2E.ButtonsMessage_Button_NativeFlowInfo{
					Name:       proto.String("cta_url"),
					ParamsJSON: proto.String(string(params)),
				},
			})
		case protocol.ButtonPhoneNumber:
			params, _ := json.Marshal(map[string]interface{}{
				"display_text": btn.DisplayText,
				"phone_number": btn.Phone,
			})
			waButtons = append(waButtons, &waE2E.ButtonsMessage_Button{
				ButtonID: proto.String(fmt.Sprintf("phone_%d", i)),
				ButtonText: &waE2E.ButtonsMessage_Button_ButtonText{
					DisplayText: proto.String(btn.Text),
				},
				Type: waE2E.ButtonsMessage_Button_NATIVE_FLOW.Enum(),
				NativeFlowInfo: &waE2E.ButtonsMessage_Button_NativeFlowInfo{
					Name:       proto.String("quick_reply"),
					ParamsJSON: proto.String(string(params)),
				},
			})
		case protocol.ButtonCopyText:
			params, _ := json.Marshal(map[string]interface{}{
				"display_text": btn.DisplayText,
				"copy_code":    btn.ID,
			})
			waButtons = append(waButtons, &waE2E.ButtonsMessage_Button{
				ButtonID: proto.String(fmt.Sprintf("copy_%d", i)),
				ButtonText: &waE2E.ButtonsMessage_Button_ButtonText{
					DisplayText: proto.String(btn.Text),
				},
				Type: waE2E.ButtonsMessage_Button_NATIVE_FLOW.Enum(),
				NativeFlowInfo: &waE2E.ButtonsMessage_Button_NativeFlowInfo{
					Name:       proto.String("quick_reply"),
					ParamsJSON: proto.String(string(params)),
				},
			})
		}
	}

	msg := &waE2E.Message{
		ButtonsMessage: &waE2E.ButtonsMessage{
			HeaderType: waE2E.ButtonsMessage_TEXT.Enum(),
			Header: &waE2E.ButtonsMessage_Text{
				Text: "",
			},
			ContentText: proto.String(text),
			FooterText:  proto.String(""),
			Buttons:     waButtons,
		},
	}

	_, err := c.Client.SendMessage(ctx, chat, msg)
	return err
}

// SendList sends a list/dropdown menu message
func SendList(c *protocol.Client, ctx context.Context, chat protocol.JID, title, description, buttonText string, sections []protocol.Section, opts ...*protocol.MessageOptions) error {
	c.Throttle(chat)
	var waSections []*waE2E.ListMessage_Section
	for _, sec := range sections {
		var rows []*waE2E.ListMessage_Row
		for _, row := range sec.Rows {
			rows = append(rows, &waE2E.ListMessage_Row{
				Title:       proto.String(row.Title),
				Description: proto.String(row.Description),
				RowID:       proto.String(row.ID),
			})
		}
		waSections = append(waSections, &waE2E.ListMessage_Section{
			Title: proto.String(sec.Title),
			Rows:  rows,
		})
	}

	msg := &waE2E.Message{
		ListMessage: &waE2E.ListMessage{
			Title:       proto.String(title),
			Description: proto.String(description),
			ButtonText:  proto.String(buttonText),
			ListType:    waE2E.ListMessage_SINGLE_SELECT.Enum(),
			Sections:    waSections,
			FooterText:  proto.String(""),
			ContextInfo: BuildNewsletterContext(c.Config()),
		},
	}

	_, err := c.Client.SendMessage(ctx, chat, msg)
	return err
}

// SendImage sends an image message
func SendImage(c *protocol.Client, ctx context.Context, chat protocol.JID, data []byte, caption string, opts ...*protocol.MessageOptions) error {
	c.Throttle(chat)
	resp, err := media.UploadMedia(c, ctx, data, whatsmeow.MediaImage)
	if err != nil {
		return err
	}

	var ctxInfo *waE2E.ContextInfo
	if c.Config() != nil && c.Config().DefaultFakeReply != nil {
		ctxInfo = proto.Clone(c.Config().DefaultFakeReply).(*waE2E.ContextInfo)
	} else {
		ctxInfo = BuildNewsletterContext(c.Config())
	}

	if len(opts) > 0 && opts[0] != nil && len(opts[0].Mentions) > 0 {
		if ctxInfo == nil {
			ctxInfo = &waE2E.ContextInfo{}
		}
		ctxInfo.MentionedJID = opts[0].Mentions
	}

	msg := &waE2E.Message{
		ImageMessage: &waE2E.ImageMessage{
			URL:           &resp.URL,
			Mimetype:      proto.String("image/jpeg"),
			Caption:       proto.String(caption),
			FileEncSHA256: resp.FileEncSHA256,
			FileSHA256:    resp.FileSHA256,
			FileLength:    &resp.FileLength,
			DirectPath:    &resp.DirectPath,
			MediaKey:      resp.MediaKey,
			ContextInfo:   ctxInfo,
		},
	}

	_, err = c.Client.SendMessage(ctx, chat, msg)
	return err
}

// SendDocument sends a document message
func SendDocument(c *protocol.Client, ctx context.Context, chat protocol.JID, data []byte, filename, mimetype, caption string, opts ...*protocol.MessageOptions) error {
	c.Throttle(chat)
	resp, err := media.UploadMedia(c, ctx, data, whatsmeow.MediaDocument)
	if err != nil {
		return err
	}

	var ctxInfo *waE2E.ContextInfo
	if c.Config() != nil && c.Config().DefaultFakeReply != nil {
		ctxInfo = proto.Clone(c.Config().DefaultFakeReply).(*waE2E.ContextInfo)
	} else {
		ctxInfo = BuildNewsletterContext(c.Config())
	}

	msg := &waE2E.Message{
		DocumentMessage: &waE2E.DocumentMessage{
			URL:           &resp.URL,
			Mimetype:      proto.String(mimetype),
			FileName:      proto.String(filename),
			Caption:       proto.String(caption),
			FileEncSHA256: resp.FileEncSHA256,
			FileSHA256:    resp.FileSHA256,
			FileLength:    &resp.FileLength,
			DirectPath:    &resp.DirectPath,
			MediaKey:      resp.MediaKey,
			ContextInfo:   ctxInfo,
		},
	}

	_, err = c.Client.SendMessage(ctx, chat, msg)
	return err
}

// SendVideo sends a video message
func SendVideo(c *protocol.Client, ctx context.Context, chat protocol.JID, data []byte, caption string, opts ...*protocol.MessageOptions) error {
	c.Throttle(chat)
	resp, err := media.UploadMedia(c, ctx, data, whatsmeow.MediaVideo)
	if err != nil {
		return err
	}

	var ctxInfo *waE2E.ContextInfo
	if c.Config() != nil && c.Config().DefaultFakeReply != nil {
		ctxInfo = proto.Clone(c.Config().DefaultFakeReply).(*waE2E.ContextInfo)
	} else {
		ctxInfo = BuildNewsletterContext(c.Config())
	}

	msg := &waE2E.Message{
		VideoMessage: &waE2E.VideoMessage{
			URL:           &resp.URL,
			Mimetype:      proto.String("video/mp4"),
			Caption:       proto.String(caption),
			FileEncSHA256: resp.FileEncSHA256,
			FileSHA256:    resp.FileSHA256,
			FileLength:    &resp.FileLength,
			DirectPath:    &resp.DirectPath,
			MediaKey:      resp.MediaKey,
			ContextInfo:   ctxInfo,
		},
	}

	_, err = c.Client.SendMessage(ctx, chat, msg)
	return err
}

// SendAudio sends an audio message
func SendAudio(c *protocol.Client, ctx context.Context, chat protocol.JID, data []byte, opts ...*protocol.MessageOptions) error {
	c.Throttle(chat)
	resp, err := media.UploadMedia(c, ctx, data, whatsmeow.MediaAudio)
	if err != nil {
		return err
	}

	msg := &waE2E.Message{
		AudioMessage: &waE2E.AudioMessage{
			URL:           &resp.URL,
			Mimetype:      proto.String("audio/mp4"),
			FileEncSHA256: resp.FileEncSHA256,
			FileSHA256:    resp.FileSHA256,
			FileLength:    &resp.FileLength,
			DirectPath:    &resp.DirectPath,
			MediaKey:      resp.MediaKey,
			ContextInfo:   BuildNewsletterContext(c.Config()),
		},
	}

	_, err = c.Client.SendMessage(ctx, chat, msg)
	return err
}

// SendSticker sends a sticker message
func SendSticker(c *protocol.Client, ctx context.Context, chat protocol.JID, data []byte, opts ...*protocol.MessageOptions) error {
	c.Throttle(chat)
	resp, err := media.UploadMedia(c, ctx, data, whatsmeow.MediaImage)
	if err != nil {
		return err
	}

	msg := &waE2E.Message{
		StickerMessage: &waE2E.StickerMessage{
			URL:           &resp.URL,
			Mimetype:      proto.String("image/webp"),
			FileEncSHA256: resp.FileEncSHA256,
			FileSHA256:    resp.FileSHA256,
			FileLength:    &resp.FileLength,
			DirectPath:    &resp.DirectPath,
			MediaKey:      resp.MediaKey,
		},
	}

	_, err = c.Client.SendMessage(ctx, chat, msg)
	return err
}

// SendLocation sends a location message
func SendLocation(c *protocol.Client, ctx context.Context, chat protocol.JID, lat, lng float64, name, address string) error {
	c.Throttle(chat)
	msg := &waE2E.Message{
		LocationMessage: &waE2E.LocationMessage{
			DegreesLatitude:  proto.Float64(lat),
			DegreesLongitude: proto.Float64(lng),
			Name:             proto.String(name),
			Address:          proto.String(address),
		},
	}

	_, err := c.Client.SendMessage(ctx, chat, msg)
	return err
}

// SendContact sends a contact/vcard message
func SendContact(c *protocol.Client, ctx context.Context, chat protocol.JID, name, phone string) error {
	c.Throttle(chat)
	vcard := fmt.Sprintf("BEGIN:VCARD\nVERSION:3.0\nFN:%s\nTEL;type=CELL;type=VOICE;waid=%s:+%s\nEND:VCARD", name, phone, phone)

	msg := &waE2E.Message{
		ContactMessage: &waE2E.ContactMessage{
			DisplayName: proto.String(name),
			Vcard:       proto.String(vcard),
		},
	}

	_, err := c.Client.SendMessage(ctx, chat, msg)
	return err
}

// SendReaction sends a reaction to a message
func SendReaction(c *protocol.Client, ctx context.Context, chat protocol.JID, msgID protocol.MessageID, emoji string, sender ...types.JID) error {
	c.Throttle(chat)
	key := &waCommon.MessageKey{
		RemoteJID: proto.String(chat.String()),
		ID:        proto.String(string(msgID)),
	}
	if len(sender) > 0 && !sender[0].IsEmpty() {
		fromMe := false
		if c.Client != nil && c.Client.Store != nil && c.Client.Store.ID != nil {
			fromMe = sender[0].ToNonAD().User == c.Client.Store.ID.ToNonAD().User
		}
		key.FromMe = proto.Bool(fromMe)
		if chat.Server == protocol.GroupServer {
			key.Participant = proto.String(sender[0].String())
		}
	}

	reaction := &waE2E.ReactionMessage{
		Key:  key,
		Text: proto.String(emoji),
	}

	_, err := c.Client.SendMessage(ctx, chat, &waE2E.Message{
		ReactionMessage: reaction,
	})
	return err
}

// SendPoll sends a poll message (auto-routes to PollCreationMessageV3 if single-select)
func SendPoll(c *protocol.Client, ctx context.Context, chat protocol.JID, name string, options []string, selectableCount int) error {
	c.Throttle(chat)
	var pollOptions []*waE2E.PollCreationMessage_Option
	for _, opt := range options {
		pollOptions = append(pollOptions, &waE2E.PollCreationMessage_Option{
			OptionName: proto.String(opt),
		})
	}

	poll := &waE2E.PollCreationMessage{
		Name:                   proto.String(name),
		Options:                pollOptions,
		SelectableOptionsCount: proto.Uint32(uint32(selectableCount)),
	}

	msg := &waE2E.Message{}
	if selectableCount == 1 {
		msg.PollCreationMessageV3 = poll
	} else {
		msg.PollCreationMessage = poll
	}

	_, err := c.Client.SendMessage(ctx, chat, msg)
	return err
}

// SendQuizPoll sends a quiz poll with a designated correct answer
func SendQuizPoll(c *protocol.Client, ctx context.Context, chat protocol.JID, name string, options []string, correctAnswer string) error {
	c.Throttle(chat)
	var pollOptions []*waE2E.PollCreationMessage_Option
	for _, opt := range options {
		pollOptions = append(pollOptions, &waE2E.PollCreationMessage_Option{
			OptionName: proto.String(opt),
		})
	}

	quizType := waE2E.PollType_QUIZ
	poll := &waE2E.PollCreationMessage{
		Name:                   proto.String(name),
		Options:                pollOptions,
		SelectableOptionsCount: proto.Uint32(1),
		PollType:               &quizType,
		CorrectAnswer: &waE2E.PollCreationMessage_Option{
			OptionName: proto.String(correctAnswer),
		},
	}

	msg := &waE2E.Message{
		PollCreationMessageV3: poll,
	}

	_, err := c.Client.SendMessage(ctx, chat, msg)
	return err
}

// HandleButtonResponse checks if a message event is a button response
// (Deprecated: implemented with nativeFlow in button_response.go)

// Helper functions

func BuildNewsletterContext(cfg *protocol.Config) *waE2E.ContextInfo {
	if cfg != nil && cfg.DefaultFakeReply != nil {
		return proto.Clone(cfg.DefaultFakeReply).(*waE2E.ContextInfo)
	}

	if cfg == nil || cfg.NewsletterJID == "" {
		return &waE2E.ContextInfo{}
	}

	return &waE2E.ContextInfo{
		IsForwarded:     proto.Bool(false),
		ForwardingScore: proto.Uint32(0),
		BusinessMessageForwardInfo: &waE2E.ContextInfo_BusinessMessageForwardInfo{
			BusinessOwnerJID: proto.String(cfg.BusinessOwnerJID),
		},
		ForwardedNewsletterMessageInfo: &waE2E.ContextInfo_ForwardedNewsletterMessageInfo{
			NewsletterJID:  proto.String(cfg.NewsletterJID),
			NewsletterName: proto.String(cfg.NewsletterName),
		},
	}
}

func applyOptions(msg *waE2E.Message, opts *protocol.MessageOptions) *waE2E.Message {
	if opts == nil {
		return msg
	}

	if len(opts.Mentions) > 0 {
		setMentions(msg, opts.Mentions)
	}

	return msg
}

func setMentions(msg *waE2E.Message, mentions []string) {
	if msg.ExtendedTextMessage != nil {
		if msg.ExtendedTextMessage.ContextInfo == nil {
			msg.ExtendedTextMessage.ContextInfo = &waE2E.ContextInfo{}
		}
		msg.ExtendedTextMessage.ContextInfo.MentionedJID = mentions
	} else if msg.Conversation != nil {
		text := *msg.Conversation
		msg.ExtendedTextMessage = &waE2E.ExtendedTextMessage{
			Text: proto.String(text),
			ContextInfo: &waE2E.ContextInfo{
				MentionedJID: mentions,
			},
		}
		msg.Conversation = nil
	}
}
