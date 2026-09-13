package buttons

import (
	"encoding/json"
	"fmt"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

// NativeFlowButton mewakili format tombol native flow WhatsApp
type NativeFlowButton struct {
	Name             string `json:"name"`
	ButtonParamsJSON string `json:"buttonParamsJson"`
}

// SingleSelectSection mewakili bagian menu dropdown
type SingleSelectSection struct {
	Title          string            `json:"title"`
	HighlightLabel string            `json:"highlight_label,omitempty"`
	Rows           []SingleSelectRow `json:"rows"`
}

// SingleSelectRow mewakili pilihan dalam dropdown
type SingleSelectRow struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	ID          string `json:"id"`
}

// BuildNativeFlowButtons membuat slice tombol native flow (single_select, cta_url, copy_text)
func BuildNativeFlowButtons(sections []SingleSelectSection, ctaText, ctaURL string, copyTextTitle, copyCode string) ([]*waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton, error) {
	var buttons []*waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton

	// 1. Single Select Dropdown
	if len(sections) > 0 {
		params, err := json.Marshal(map[string]interface{}{
			"title":    "Selection",
			"sections": sections,
		})
		if err != nil {
			return nil, fmt.Errorf("marshal single_select failed: %w", err)
		}
		buttons = append(buttons, &waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
			Name:             proto.String("single_select"),
			ButtonParamsJSON: proto.String(string(params)),
		})
	}

	// 2. Call To Action (URL)
	if ctaURL != "" {
		if ctaText == "" {
			ctaText = "Visit Website"
		}
		params, err := json.Marshal(map[string]interface{}{
			"display_text": ctaText,
			"url":          ctaURL,
			"merchant_url": ctaURL,
		})
		if err != nil {
			return nil, fmt.Errorf("marshal cta_url failed: %w", err)
		}
		buttons = append(buttons, &waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
			Name:             proto.String("cta_url"),
			ButtonParamsJSON: proto.String(string(params)),
		})
	}

	// 3. Quick Copy Code
	if copyCode != "" {
		if copyTextTitle == "" {
			copyTextTitle = "Salin Kode"
		}
		params, err := json.Marshal(map[string]interface{}{
			"display_text": copyTextTitle,
			"copy_code":    copyCode,
		})
		if err != nil {
			return nil, fmt.Errorf("marshal copy_text failed: %w", err)
		}
		buttons = append(buttons, &waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
			Name:             proto.String("cta_copy"),
			ButtonParamsJSON: proto.String(string(params)),
		})
	}

	return buttons, nil
}

// BuildInteractiveMessage membungkus tombol native flow ke dalam interactive message WhatsApp
func BuildInteractiveMessage(bodyText, footerText string, headerDoc *waE2E.DocumentMessage, buttons []*waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton, newsletterJID, newsletterName string) *waE2E.Message {
	header := &waE2E.InteractiveMessage_Header{
		HasMediaAttachment: proto.Bool(headerDoc != nil),
	}
	if headerDoc != nil {
		header.Media = &waE2E.InteractiveMessage_Header_DocumentMessage{
			DocumentMessage: headerDoc,
		}
	}

	ctxInfo := &waE2E.ContextInfo{
		IsForwarded:     proto.Bool(false),
		ForwardingScore: proto.Uint32(0),
	}
	if newsletterJID != "" {
		ctxInfo.ForwardedNewsletterMessageInfo = &waE2E.ContextInfo_ForwardedNewsletterMessageInfo{
			NewsletterJID:  proto.String(newsletterJID),
			NewsletterName: proto.String(newsletterName),
		}
	}

	interactiveMsg := &waE2E.InteractiveMessage{
		Header: header,
		Body: &waE2E.InteractiveMessage_Body{
			Text: proto.String(bodyText),
		},
		Footer: &waE2E.InteractiveMessage_Footer{
			Text: proto.String(footerText),
		},
		InteractiveMessage: &waE2E.InteractiveMessage_NativeFlowMessage_{
			NativeFlowMessage: &waE2E.InteractiveMessage_NativeFlowMessage{
				Buttons: buttons,
			},
		},
		ContextInfo: ctxInfo,
	}

	// Bungkus ke ViewOnceMessage agar tombol muncul stabil di WhatsApp Web & Mobile
	return &waE2E.Message{
		ViewOnceMessage: &waE2E.FutureProofMessage{
			Message: &waE2E.Message{
				InteractiveMessage: interactiveMsg,
			},
		},
	}
}
