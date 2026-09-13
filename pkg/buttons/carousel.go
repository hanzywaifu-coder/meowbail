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

// CarouselCard mewakili 1 kartu geser dalam Carousel Message
type CarouselCard struct {
	Title      string
	Body       string
	Footer     string
	MediaImage []byte
	Buttons    []protocol.Button
}

// SendCarousel mengirimkan pesan Carousel (slider cards dengan gambar & tombol masing-masing)
func SendCarousel(c *protocol.Client, ctx context.Context, chat protocol.JID, text string, cards []CarouselCard) error {
	if len(cards) == 0 {
		return fmt.Errorf("carousel must have at least one card")
	}

	var protoCards []*waE2E.InteractiveMessage

	for _, card := range cards {
		var imgMsg *waE2E.ImageMessage
		if len(card.MediaImage) > 0 {
			resp, err := c.Client.Upload(ctx, card.MediaImage, whatsmeow.MediaImage)
			if err == nil {
				imgMsg = &waE2E.ImageMessage{
					URL:           &resp.URL,
					Mimetype:      proto.String("image/jpeg"),
					FileEncSHA256: resp.FileEncSHA256,
					FileSHA256:    resp.FileSHA256,
					FileLength:    &resp.FileLength,
					DirectPath:    &resp.DirectPath,
					MediaKey:      resp.MediaKey,
				}
			}
		}

		var cardButtons []*waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton
		for i, btn := range card.Buttons {
			switch btn.Type {
			case protocol.ButtonQuickReply:
				params, _ := json.Marshal(map[string]interface{}{
					"display_text": btn.Text,
					"id":           btn.ID,
				})
				cardButtons = append(cardButtons, &waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
					Name:             proto.String("quick_reply"),
					ButtonParamsJSON: proto.String(string(params)),
				})
			case protocol.ButtonCTAURL:
				displayText := btn.DisplayText
				if displayText == "" {
					displayText = btn.Text
				}
				params, _ := json.Marshal(map[string]interface{}{
					"display_text": displayText,
					"url":          btn.URL,
					"merchant_url": btn.URL,
				})
				cardButtons = append(cardButtons, &waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
					Name:             proto.String("cta_url"),
					ButtonParamsJSON: proto.String(string(params)),
				})
			case protocol.ButtonCopyText:
				params, _ := json.Marshal(map[string]interface{}{
					"display_text": btn.DisplayText,
					"copy_code":    btn.ID,
				})
				cardButtons = append(cardButtons, &waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
					Name:             proto.String("cta_copy"),
					ButtonParamsJSON: proto.String(string(params)),
				})
			default:
				params, _ := json.Marshal(map[string]interface{}{
					"display_text": fmt.Sprintf("Btn %d", i+1),
					"id":           btn.ID,
				})
				cardButtons = append(cardButtons, &waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
					Name:             proto.String("quick_reply"),
					ButtonParamsJSON: proto.String(string(params)),
				})
			}
		}

		header := &waE2E.InteractiveMessage_Header{
			Title:              proto.String(card.Title),
			HasMediaAttachment: proto.Bool(imgMsg != nil),
		}
		if imgMsg != nil {
			header.Media = &waE2E.InteractiveMessage_Header_ImageMessage{
				ImageMessage: imgMsg,
			}
		}

		protoCards = append(protoCards, &waE2E.InteractiveMessage{
			Header: header,
			Body: &waE2E.InteractiveMessage_Body{
				Text: proto.String(card.Body),
			},
			Footer: &waE2E.InteractiveMessage_Footer{
				Text: proto.String(card.Footer),
			},
			InteractiveMessage: &waE2E.InteractiveMessage_NativeFlowMessage_{
				NativeFlowMessage: &waE2E.InteractiveMessage_NativeFlowMessage{
					Buttons: cardButtons,
				},
			},
		})
	}

	carouselMessage := &waE2E.InteractiveMessage_CarouselMessage{
		Cards: protoCards,
	}

	interactiveMsg := &waE2E.InteractiveMessage{
		Header: &waE2E.InteractiveMessage_Header{
			HasMediaAttachment: proto.Bool(false),
		},
		Body: &waE2E.InteractiveMessage_Body{
			Text: proto.String(text),
		},
		Footer: &waE2E.InteractiveMessage_Footer{
			Text: proto.String("Dongtube"),
		},
		InteractiveMessage: &waE2E.InteractiveMessage_CarouselMessage_{
			CarouselMessage: carouselMessage,
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
