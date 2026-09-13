package business

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

// SendOrderDetails mengirim template pesanan / tagihan Order Details Native Flow
func SendOrderDetails(c *protocol.Client, ctx context.Context, chat protocol.JID, title string, orderID string, items []map[string]interface{}, totalAmount1000 int64, currency string) error {
	if currency == "" {
		currency = "IDR"
	}

	orderData := map[string]interface{}{
		"order_id": orderID,
		"title":    title,
		"currency": currency,
		"total":    totalAmount1000,
		"items":    items,
	}
	orderJSON, _ := json.Marshal(orderData)

	btnParams := fmt.Sprintf(`{"order_id":"%s","item_count":%d,"total_price":%d,"currency":"%s"}`, orderID, len(items), totalAmount1000, currency)

	msg := &waE2E.Message{
		InteractiveMessage: &waE2E.InteractiveMessage{
			Body: &waE2E.InteractiveMessage_Body{
				Text: proto.String(title),
			},
			InteractiveMessage: &waE2E.InteractiveMessage_NativeFlowMessage_{
				NativeFlowMessage: &waE2E.InteractiveMessage_NativeFlowMessage{
					Buttons: []*waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
						{
							Name:             proto.String("review_and_pay"),
							ButtonParamsJSON: proto.String(btnParams),
						},
					},
					MessageParamsJSON: proto.String(string(orderJSON)),
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

// SendInteractiveAddressPrompt meminta alamat pengiriman fisik pengguna (WhatsApp Address Message)
func SendInteractiveAddressPrompt(c *protocol.Client, ctx context.Context, chat protocol.JID, title string) error {
	btnParams := `{"values":{},"saved_addresses":[]}`

	msg := &waE2E.Message{
		InteractiveMessage: &waE2E.InteractiveMessage{
			Body: &waE2E.InteractiveMessage_Body{
				Text: proto.String(title),
			},
			InteractiveMessage: &waE2E.InteractiveMessage_NativeFlowMessage_{
				NativeFlowMessage: &waE2E.InteractiveMessage_NativeFlowMessage{
					Buttons: []*waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
						{
							Name:             proto.String("address_message"),
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

// SendCallPrompt mengirim tombol telepon cepat langsung memanggil nomor suara (CTA Call)
func SendCallPrompt(c *protocol.Client, ctx context.Context, chat protocol.JID, bodyText, buttonText, phoneNumber string) error {
	btnParams := fmt.Sprintf(`{"display_text":"%s","phone_number":"%s"}`, buttonText, phoneNumber)

	msg := &waE2E.Message{
		InteractiveMessage: &waE2E.InteractiveMessage{
			Body: &waE2E.InteractiveMessage_Body{
				Text: proto.String(bodyText),
			},
			InteractiveMessage: &waE2E.InteractiveMessage_NativeFlowMessage_{
				NativeFlowMessage: &waE2E.InteractiveMessage_NativeFlowMessage{
					Buttons: []*waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
						{
							Name:             proto.String("cta_call"),
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
