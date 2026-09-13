package buttons

import (
	"context"
	"encoding/json"
	"fmt"
	protocol "github.com/hanzywaifu-coder/dongtube-meowbail/internal/protocol"
	groups "github.com/hanzywaifu-coder/dongtube-meowbail/pkg/groups"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

// SendButtonV2 mengirim format ButtonsMessage (header location/image + viewOnce: true + biz node)
// Ini adalah format paling stabil yang dirender oleh aplikasi WhatsApp Android/iOS resmi
func SendButtonV2(c *protocol.Client, ctx context.Context, chat protocol.JID, title, bodyText, footerText string, thumbData []byte, buttons []protocol.Button) error {
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
					Name:       proto.String("cta_copy"),
					ParamsJSON: proto.String(string(params)),
				},
			})
		}
	}

	headerType := waE2E.ButtonsMessage_LOCATION
	buttonsMsg := &waE2E.ButtonsMessage{
		HeaderType: &headerType,
		Header: &waE2E.ButtonsMessage_LocationMessage{
			LocationMessage: &waE2E.LocationMessage{
				DegreesLatitude:  proto.Float64(0),
				DegreesLongitude: proto.Float64(0),
				Name:             proto.String(title),
				Address:          proto.String(""),
				JPEGThumbnail:    thumbData,
			},
		},
		ContentText: proto.String(bodyText),
		FooterText:  proto.String(footerText),
		Buttons:     waButtons,
		ContextInfo: newsletterContext(c.GetConfig()),
	}

	msg := &waE2E.Message{
		ViewOnceMessage: &waE2E.FutureProofMessage{
			Message: &waE2E.Message{
				ButtonsMessage: buttonsMsg,
			},
		},
	}

	bizNodes := groups.BuildBizAdditionalNodes()
	_, err := c.Client.SendMessage(ctx, chat, msg, whatsmeow.SendRequestExtra{
		AdditionalNodes: &bizNodes,
	})
	return err
}

// SendInteractiveCard mengirim InteractiveMessage tanpa bloksWidget yang sering di-drop
// tetapi menggunakan ViewOnceMessage + InteractiveMessage + NativeFlow + BizNode
func SendInteractiveCard(c *protocol.Client, ctx context.Context, chat protocol.JID, thumbData []byte, bodyText, footerText string, sections []protocol.Section, ctaURL, ctaText string) error {
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

	catParamsJSON, _ := json.Marshal(map[string]interface{}{
		"title":    "Pilih Menu",
		"sections": waSections,
	})

	nativeButtons := []*waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
		{
			Name:             proto.String("single_select"),
			ButtonParamsJSON: proto.String(string(catParamsJSON)),
		},
	}

	if ctaURL != "" {
		if ctaText == "" {
			ctaText = "Official Channel"
		}
		ctaParamsJSON, _ := json.Marshal(map[string]interface{}{
			"display_text": ctaText,
			"url":          ctaURL,
			"merchant_url": ctaURL,
		})
		nativeButtons = append(nativeButtons, &waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
			Name:             proto.String("cta_url"),
			ButtonParamsJSON: proto.String(string(ctaParamsJSON)),
		})
	}

	header := &waE2E.InteractiveMessage_Header{
		HasMediaAttachment: proto.Bool(len(thumbData) > 0),
		Media: &waE2E.InteractiveMessage_Header_LocationMessage{
			LocationMessage: &waE2E.LocationMessage{
				DegreesLatitude:  proto.Float64(0),
				DegreesLongitude: proto.Float64(0),
				Name:             proto.String(""),
				Address:          proto.String(""),
				JPEGThumbnail:    thumbData,
			},
		},
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
				Buttons: nativeButtons,
			},
		},
		ContextInfo: newsletterContext(c.GetConfig()),
	}

	msg := &waE2E.Message{
		ViewOnceMessage: &waE2E.FutureProofMessage{
			Message: &waE2E.Message{
				InteractiveMessage: interactiveMsg,
			},
		},
	}

	bizNodes := groups.BuildBizAdditionalNodes()
	_, err := c.Client.SendMessage(ctx, chat, msg, whatsmeow.SendRequestExtra{
		AdditionalNodes: &bizNodes,
	})
	return err
}
