package business

import (
	"context"
	"encoding/json"
	protocol "github.com/hanzywaifu-coder/dongtube-meowbail/internal/protocol"
	groups "github.com/hanzywaifu-coder/dongtube-meowbail/pkg/groups"
	media "github.com/hanzywaifu-coder/dongtube-meowbail/pkg/media"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

// SendCopyCodeButton mengirim pesan interaktif dengan tombol Copy Code voucher/promo
func SendCopyCodeButton(c *protocol.Client, ctx context.Context, chat protocol.JID, bodyText, buttonText, promoCode string) error {
	btnParams := `{"display_text":"` + buttonText + `","code":"` + promoCode + `"}`

	msg := &waE2E.Message{
		InteractiveMessage: &waE2E.InteractiveMessage{
			Body: &waE2E.InteractiveMessage_Body{
				Text: proto.String(bodyText),
			},
			InteractiveMessage: &waE2E.InteractiveMessage_NativeFlowMessage_{
				NativeFlowMessage: &waE2E.InteractiveMessage_NativeFlowMessage{
					Buttons: []*waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
						{
							Name:             proto.String("cta_copy"),
							ButtonParamsJSON: proto.String(btnParams),
						},
					},
				},
			},
		},
	}

	bizNodes := groups.BuildBizAdditionalNodes()
	_, err := c.Client.SendMessage(ctx, chat, msg, whatsmeow.SendRequestExtra{
		AdditionalNodes: &bizNodes,
	})
	return err
}

// SendReminderButton mengirim tombol pengingat event / kalender native
func SendReminderButton(c *protocol.Client, ctx context.Context, chat protocol.JID, bodyText, buttonText, title string, timestamp int64) error {
	btnParams := `{"display_text":"` + buttonText + `","title":"` + title + `"}`

	msg := &waE2E.Message{
		InteractiveMessage: &waE2E.InteractiveMessage{
			Body: &waE2E.InteractiveMessage_Body{
				Text: proto.String(bodyText),
			},
			InteractiveMessage: &waE2E.InteractiveMessage_NativeFlowMessage_{
				NativeFlowMessage: &waE2E.InteractiveMessage_NativeFlowMessage{
					Buttons: []*waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
						{
							Name:             proto.String("cta_reminder"),
							ButtonParamsJSON: proto.String(btnParams),
						},
					},
				},
			},
		},
	}

	bizNodes := groups.BuildBizAdditionalNodes()
	_, err := c.Client.SendMessage(ctx, chat, msg, whatsmeow.SendRequestExtra{
		AdditionalNodes: &bizNodes,
	})
	return err
}

// SendInteractiveListWithImage mengirim menu single_select dengan Header Gambar
func SendInteractiveListWithImage(c *protocol.Client, ctx context.Context, chat protocol.JID, imageData []byte, bodyText, footerText string, buttonTitle string, sections []protocol.Section) error {
	uploaded, err := media.UploadMedia(c, ctx, imageData, whatsmeow.MediaImage)
	if err != nil {
		return err
	}

	waSections := make([]map[string]interface{}, 0, len(sections))
	for _, sec := range sections {
		rows := make([]map[string]interface{}, 0, len(sec.Rows))
		for _, r := range sec.Rows {
			rowMap := map[string]interface{}{
				"title": r.Title,
				"id":    r.ID,
			}
			if r.Description != "" {
				rowMap["description"] = r.Description
			}
			rows = append(rows, rowMap)
		}
		waSections = append(waSections, map[string]interface{}{
			"title": sec.Title,
			"rows":  rows,
		})
	}

	selectJSONBytes, _ := json.Marshal(map[string]interface{}{
		"title":    buttonTitle,
		"sections": waSections,
	})

	msg := &waE2E.Message{
		InteractiveMessage: &waE2E.InteractiveMessage{
			Header: &waE2E.InteractiveMessage_Header{
				HasMediaAttachment: proto.Bool(true),
				Media: &waE2E.InteractiveMessage_Header_ImageMessage{
					ImageMessage: &waE2E.ImageMessage{
						URL:           proto.String(uploaded.URL),
						DirectPath:    proto.String(uploaded.DirectPath),
						MediaKey:      uploaded.MediaKey,
						Mimetype:      proto.String("image/jpeg"),
						FileEncSHA256: uploaded.FileEncSHA256,
						FileSHA256:    uploaded.FileSHA256,
						FileLength:    proto.Uint64(uploaded.FileLength),
					},
				},
			},
			Body: &waE2E.InteractiveMessage_Body{
				Text: proto.String(bodyText),
			},
			Footer: &waE2E.InteractiveMessage_Footer{
				Text: proto.String(footerText),
			},
			InteractiveMessage: &waE2E.InteractiveMessage_NativeFlowMessage_{
				NativeFlowMessage: &waE2E.InteractiveMessage_NativeFlowMessage{
					Buttons: []*waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
						{
							Name:             proto.String("single_select"),
							ButtonParamsJSON: proto.String(string(selectJSONBytes)),
						},
					},
				},
			},
		},
	}

	bizNodes := groups.BuildBizAdditionalNodes()
	_, err = c.Client.SendMessage(ctx, chat, msg, whatsmeow.SendRequestExtra{
		AdditionalNodes: &bizNodes,
	})
	return err
}
