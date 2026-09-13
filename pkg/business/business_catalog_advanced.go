package business

import (
	"context"
	protocol "github.com/hanzywaifu-coder/dongtube-meowbail/internal/protocol"
	groups "github.com/hanzywaifu-coder/dongtube-meowbail/pkg/groups"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

// CatalogItem mendefinisikan item produk di katalog
type CatalogItem struct {
	ID          string
	Title       string
	Description string
	Price       int64
	Currency    string
	ImageBytes  []byte
}

// SendInteractiveProductList mengirim daftar produk Native Flow (WhatsApp Multi-Product Sections)
func SendInteractiveProductList(c *protocol.Client, ctx context.Context, chat protocol.JID, businessOwnerJID protocol.JID, headerText, bodyText, footerText string, catalogTitle string, items []CatalogItem) error {
	var productSections []*waE2E.ListMessage_ProductSection

	var sectionProducts []*waE2E.ListMessage_Product
	for _, it := range items {
		sectionProducts = append(sectionProducts, &waE2E.ListMessage_Product{
			ProductID: proto.String(it.ID),
		})
	}

	productSections = append(productSections, &waE2E.ListMessage_ProductSection{
		Title:    proto.String(catalogTitle),
		Products: sectionProducts,
	})

	msg := &waE2E.Message{
		ListMessage: &waE2E.ListMessage{
			Title:       proto.String(headerText),
			Description: proto.String(bodyText),
			FooterText:  proto.String(footerText),
			ListType:    waE2E.ListMessage_PRODUCT_LIST.Enum(),
			ProductListInfo: &waE2E.ListMessage_ProductListInfo{
				ProductSections:  productSections,
				BusinessOwnerJID: proto.String(businessOwnerJID.String()),
			},
		},
	}

	bizNodes := groups.BuildBizAdditionalNodes()
	_, err := c.Client.SendMessage(ctx, chat, msg, whatsmeow.SendRequestExtra{
		AdditionalNodes: &bizNodes,
	})
	return err
}

// SendVoiceCallPrompt mengirim tombol aksi langsung panggil telepon suara (Voice Call protocol.Button)
func SendVoiceCallPrompt(c *protocol.Client, ctx context.Context, chat protocol.JID, bodyText, callText, phoneNumber string) error {
	btnParams := `{"display_text":"` + callText + `","phone_number":"` + phoneNumber + `"}`

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

// SendCallConfirmation mengirim pesan template panggilan terjadwal (Scheduled Call)
func SendCallConfirmation(c *protocol.Client, ctx context.Context, chat protocol.JID, title string, callType waE2E.ScheduledCallCreationMessage_CallType, startTime int64) error {
	msg := &waE2E.Message{
		ScheduledCallCreationMessage: &waE2E.ScheduledCallCreationMessage{
			ScheduledTimestampMS: proto.Int64(startTime),
			CallType:             callType.Enum(),
			Title:                proto.String(title),
		},
	}
	_, err := c.Client.SendMessage(ctx, chat, msg)
	return err
}
