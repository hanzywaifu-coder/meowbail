package messaging

import (
	"context"
	"encoding/json"
	protocol "github.com/hanzywaifu-coder/dongtube-meowbail/internal/protocol"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

// InteractiveMenuButton mendefinisikan tombol interaktif
type InteractiveMenuButton struct {
	Name             string `json:"name"`
	ButtonParamsJSON string `json:"buttonParamsJson"`
}

// SendInteractiveMenu mengirim menu persis Baileys modern / dongtube-meowbail:
// viewOnceMessage -> interactiveMessage (Document header + body + footer + nativeFlowMessage buttons)
func SendInteractiveMenu(c *protocol.Client, ctx context.Context, chat protocol.JID, docData []byte, thumbData []byte, footerText string, sections []protocol.Section, ctaText, ctaURL string, channelInfo *protocol.Config) error {
	var docMsg *waE2E.DocumentMessage
	if len(docData) > 0 {
		resp, err := c.Client.Upload(ctx, docData, whatsmeow.MediaDocument)
		if err == nil {
			docLen := uint64(len(docData))
			docMsg = &waE2E.DocumentMessage{
				URL:           &resp.URL,
				Mimetype:      proto.String("image/png"),
				FileName:      proto.String("Dongtube"),
				FileLength:    &docLen,
				PageCount:     proto.Uint32(100),
				FileEncSHA256: resp.FileEncSHA256,
				FileSHA256:    resp.FileSHA256,
				DirectPath:    &resp.DirectPath,
				MediaKey:      resp.MediaKey,
				JPEGThumbnail: thumbData,
			}
		}
	}

	var waSections []map[string]interface{}
	for _, sec := range sections {
		var rows []map[string]interface{}
		for _, r := range sec.Rows {
			rows = append(rows, map[string]interface{}{
				"title":       r.Title,
				"description": r.Description,
				"id":          r.ID,
			})
		}
		waSections = append(waSections, map[string]interface{}{
			"title": sec.Title,
			"rows":  rows,
		})
	}

	catParams, _ := json.Marshal(map[string]interface{}{
		"title":    "Selection",
		"sections": waSections,
	})

	ctaParams, _ := json.Marshal(map[string]interface{}{
		"display_text": ctaText,
		"url":          ctaURL,
		"merchant_url": ctaURL,
	})

	nativeButtons := []*waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
		{
			Name:             proto.String("single_select"),
			ButtonParamsJSON: proto.String(string(catParams)),
		},
	}

	if ctaURL != "" {
		nativeButtons = append(nativeButtons, &waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
			Name:             proto.String("cta_url"),
			ButtonParamsJSON: proto.String(string(ctaParams)),
		})
	}

	header := &waE2E.InteractiveMessage_Header{
		Title:              proto.String(""),
		HasMediaAttachment: proto.Bool(docMsg != nil),
	}
	if docMsg != nil {
		header.Media = &waE2E.InteractiveMessage_Header_DocumentMessage{
			DocumentMessage: docMsg,
		}
	}

	ctxInfo := &waE2E.ContextInfo{
		IsForwarded:     proto.Bool(false),
		ForwardingScore: proto.Uint32(0),
	}
	if channelInfo != nil && channelInfo.NewsletterJID != "" {
		ctxInfo.ForwardedNewsletterMessageInfo = &waE2E.ContextInfo_ForwardedNewsletterMessageInfo{
			NewsletterJID:  proto.String(channelInfo.NewsletterJID),
			NewsletterName: proto.String(channelInfo.NewsletterName),
		}
	}

	interactiveMsg := &waE2E.InteractiveMessage{
		Header: header,
		Body: &waE2E.InteractiveMessage_Body{
			Text: proto.String(""),
		},
		Footer: &waE2E.InteractiveMessage_Footer{
			Text: proto.String(footerText),
		},
		InteractiveMessage: &waE2E.InteractiveMessage_NativeFlowMessage_{
			NativeFlowMessage: &waE2E.InteractiveMessage_NativeFlowMessage{
				Buttons: nativeButtons,
			},
		},
		ContextInfo: ctxInfo,
	}

	// Bungkus ke dalam ViewOnceMessageV2 / ViewOnceMessage persis Baileys
	msg := &waE2E.Message{
		ViewOnceMessage: &waE2E.FutureProofMessage{
			Message: &waE2E.Message{
				InteractiveMessage: interactiveMsg,
			},
		},
	}

	_, err := c.Client.SendMessage(ctx, chat, msg)
	return err
}
