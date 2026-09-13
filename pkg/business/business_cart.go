package business

import (
	"context"
	protocol "github.com/hanzywaifu-coder/dongtube-meowbail/internal/protocol"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

// SendOrderReceipt mengirim struk belanja resmi pesanan (Order Receipt Message)
func SendOrderReceipt(c *protocol.Client, ctx context.Context, chat protocol.JID, orderID, orderTitle string, itemCount int32, thumb []byte, sellerJID string) error {
	if sellerJID == "" {
		sellerJID = "0@s.whatsapp.net"
	}

	status := waE2E.OrderMessage_OrderStatus(1)
	surface := waE2E.OrderMessage_OrderSurface(1)

	msg := &waE2E.Message{
		OrderMessage: &waE2E.OrderMessage{
			OrderID:    proto.String(orderID),
			OrderTitle: proto.String(orderTitle),
			ItemCount:  proto.Int32(itemCount),
			Status:     &status,
			Surface:    &surface,
			SellerJID:  proto.String(sellerJID),
			Thumbnail:  thumb,
		},
	}

	_, err := c.Client.SendMessage(ctx, chat, msg)
	return err
}

// SendPaymentInvite mengirim ajakan transfer pembayaran (Payment Invite Message)
func SendPaymentInvite(c *protocol.Client, ctx context.Context, chat protocol.JID, serviceType waE2E.PaymentInviteMessage_ServiceType, expiry int64) error {
	msg := &waE2E.Message{
		PaymentInviteMessage: &waE2E.PaymentInviteMessage{
			ServiceType:     serviceType.Enum(),
			ExpiryTimestamp: proto.Int64(expiry),
		},
	}
	_, err := c.Client.SendMessage(ctx, chat, msg)
	return err
}

// SendForwardedNewsletterPost meneruskan postingan dari channel WhatsApp ke chat/grup
func SendForwardedNewsletterPost(c *protocol.Client, ctx context.Context, chat protocol.JID, text string, channelJID protocol.JID, channelName string, serverMessageID int64) error {
	msg := &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: proto.String(text),
			ContextInfo: &waE2E.ContextInfo{
				IsForwarded:     proto.Bool(false),
				ForwardingScore: proto.Uint32(0),
				ForwardedNewsletterMessageInfo: &waE2E.ContextInfo_ForwardedNewsletterMessageInfo{
					NewsletterJID:   proto.String(channelJID.String()),
					NewsletterName:  proto.String(channelName),
					ServerMessageID: proto.Int32(int32(serverMessageID)),
				},
			},
		},
	}
	_, err := c.Client.SendMessage(ctx, chat, msg)
	return err
}
