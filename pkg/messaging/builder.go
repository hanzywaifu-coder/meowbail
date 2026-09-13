package messaging

import (
	"encoding/json"
	"fmt"

	buttons "github.com/hanzywaifu-coder/dongtube-meowbail/pkg/buttons"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

// FluentButtonBuilder mengimplementasikan builder pattern modern untuk interactive message
type FluentButtonBuilder struct {
	title       string
	subtitle    string
	body        string
	footer      string
	headerDoc   *waE2E.DocumentMessage
	headerImg   *waE2E.ImageMessage
	buttons     []*waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton
	contextInfo *waE2E.ContextInfo
}

func NewButtonBuilder() *FluentButtonBuilder {
	return &FluentButtonBuilder{
		buttons: make([]*waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton, 0),
	}
}

func (b *FluentButtonBuilder) SetTitle(title string) *FluentButtonBuilder {
	b.title = title
	return b
}

func (b *FluentButtonBuilder) SetSubtitle(subtitle string) *FluentButtonBuilder {
	b.subtitle = subtitle
	return b
}

func (b *FluentButtonBuilder) SetBody(body string) *FluentButtonBuilder {
	b.body = body
	return b
}

func (b *FluentButtonBuilder) SetFooter(footer string) *FluentButtonBuilder {
	b.footer = footer
	return b
}

func (b *FluentButtonBuilder) SetDocumentHeader(doc *waE2E.DocumentMessage) *FluentButtonBuilder {
	b.headerDoc = doc
	return b
}

func (b *FluentButtonBuilder) SetImageHeader(img *waE2E.ImageMessage) *FluentButtonBuilder {
	b.headerImg = img
	return b
}

func (b *FluentButtonBuilder) SetNewsletter(jid, name string) *FluentButtonBuilder {
	b.contextInfo = &waE2E.ContextInfo{
		IsForwarded:     proto.Bool(false),
		ForwardingScore: proto.Uint32(0),
		ForwardedNewsletterMessageInfo: &waE2E.ContextInfo_ForwardedNewsletterMessageInfo{
			NewsletterJID:  proto.String(jid),
			NewsletterName: proto.String(name),
		},
	}
	return b
}

// AddQuickReply menambahkan tombol respon cepat
func (b *FluentButtonBuilder) AddQuickReply(displayText, id string) *FluentButtonBuilder {
	params, _ := json.Marshal(map[string]interface{}{
		"display_text": displayText,
		"id":           id,
	})
	b.buttons = append(b.buttons, &waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
		Name:             proto.String("quick_reply"),
		ButtonParamsJSON: proto.String(string(params)),
	})
	return b
}

// AddCTAURL menambahkan tombol link web
func (b *FluentButtonBuilder) AddCTAURL(displayText, url string) *FluentButtonBuilder {
	params, _ := json.Marshal(map[string]interface{}{
		"display_text": displayText,
		"url":          url,
		"merchant_url": url,
	})
	b.buttons = append(b.buttons, &waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
		Name:             proto.String("cta_url"),
		ButtonParamsJSON: proto.String(string(params)),
	})
	return b
}

// AddCTACall menambahkan tombol panggilan telepon
func (b *FluentButtonBuilder) AddCTACall(displayText, phoneNumber string) *FluentButtonBuilder {
	params, _ := json.Marshal(map[string]interface{}{
		"display_text": displayText,
		"phone_number": phoneNumber,
	})
	b.buttons = append(b.buttons, &waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
		Name:             proto.String("cta_call"),
		ButtonParamsJSON: proto.String(string(params)),
	})
	return b
}

// AddCopyCode menambahkan tombol salin kode otomatis
func (b *FluentButtonBuilder) AddCopyCode(displayText, code string) *FluentButtonBuilder {
	params, _ := json.Marshal(map[string]interface{}{
		"display_text": displayText,
		"copy_code":    code,
	})
	b.buttons = append(b.buttons, &waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
		Name:             proto.String("cta_copy"),
		ButtonParamsJSON: proto.String(string(params)),
	})
	return b
}

// AddSingleSelect menambahkan menu dropdown dengan sections
func (b *FluentButtonBuilder) AddSingleSelect(title string, sections []buttons.SingleSelectSection) *FluentButtonBuilder {
	params, _ := json.Marshal(map[string]interface{}{
		"title":    title,
		"sections": sections,
	})
	b.buttons = append(b.buttons, &waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
		Name:             proto.String("single_select"),
		ButtonParamsJSON: proto.String(string(params)),
	})
	return b
}

// Build mengompilasi seluruh konfigurasi menjadi message viewOnce yang siap dikirimkan
func (b *FluentButtonBuilder) Build() (*waE2E.Message, error) {
	if len(b.buttons) == 0 {
		return nil, fmt.Errorf("at least one button is required")
	}

	header := &waE2E.InteractiveMessage_Header{
		Title:              proto.String(b.title),
		Subtitle:           proto.String(b.subtitle),
		HasMediaAttachment: proto.Bool(b.headerDoc != nil || b.headerImg != nil),
	}

	if b.headerDoc != nil {
		header.Media = &waE2E.InteractiveMessage_Header_DocumentMessage{
			DocumentMessage: b.headerDoc,
		}
	} else if b.headerImg != nil {
		header.Media = &waE2E.InteractiveMessage_Header_ImageMessage{
			ImageMessage: b.headerImg,
		}
	}

	interactiveMsg := &waE2E.InteractiveMessage{
		Header: header,
		Body: &waE2E.InteractiveMessage_Body{
			Text: proto.String(b.body),
		},
		Footer: &waE2E.InteractiveMessage_Footer{
			Text: proto.String(b.footer),
		},
		InteractiveMessage: &waE2E.InteractiveMessage_NativeFlowMessage_{
			NativeFlowMessage: &waE2E.InteractiveMessage_NativeFlowMessage{
				Buttons: b.buttons,
			},
		},
		ContextInfo: b.contextInfo,
	}

	return &waE2E.Message{
		ViewOnceMessage: &waE2E.FutureProofMessage{
			Message: &waE2E.Message{
				InteractiveMessage: interactiveMsg,
			},
		},
	}, nil
}
