package channels

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	protocol "github.com/hanzywaifu-coder/dongtube-meowbail/internal/protocol"
	"time"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

// NewsletterHeaderType mendefinisikan tipe media yang diposting ke Saluran
type NewsletterMediaType string

const (
	NewsletterText  NewsletterMediaType = "text"
	NewsletterImage NewsletterMediaType = "image"
	NewsletterVideo NewsletterMediaType = "video"
)

// BuildNewsletterMessage membungkus pesan menjadi format resmi siaran Saluran WhatsApp (Newsletter)
func BuildNewsletterMessage(text string, mediaType NewsletterMediaType, mediaURL string) *waE2E.Message {
	switch mediaType {
	case NewsletterImage:
		return &waE2E.Message{
			ImageMessage: &waE2E.ImageMessage{
				URL:      &mediaURL,
				Mimetype: proto.String("image/jpeg"),
				Caption:  proto.String(text),
			},
		}
	case NewsletterVideo:
		return &waE2E.Message{
			VideoMessage: &waE2E.VideoMessage{
				URL:      &mediaURL,
				Mimetype: proto.String("video/mp4"),
				Caption:  proto.String(text),
			},
		}
	default:
		return &waE2E.Message{
			ExtendedTextMessage: &waE2E.ExtendedTextMessage{
				Text: proto.String(text),
			},
		}
	}
}

// GenerateMessageSecret membuat kunci rahasia acak untuk penandatanganan messageContextInfo
func GenerateMessageSecret(key []byte, msgID protocol.MessageID) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(msgID))
	ts := make([]byte, 8)
	binary.BigEndian.PutUint64(ts, uint64(time.Now().Unix()))
	h.Write(ts)
	return h.Sum(nil)
}

// BuildOrderReviewMessage membuat pesan konfirmasi pembayaran atau invoice (Native Flow Order)
func BuildOrderReviewMessage(title, currency, amount string, orderID string) (*waE2E.Message, error) {
	params := fmt.Sprintf(`{"type":"review_and_pay","currency":"%s","total_amount":{"value":"%s","offset":100},"order_id":"%s"}`, currency, amount, orderID)

	interactive := &waE2E.InteractiveMessage{
		Header: &waE2E.InteractiveMessage_Header{
			Title: proto.String(title),
		},
		Body: &waE2E.InteractiveMessage_Body{
			Text: proto.String(fmt.Sprintf("Total Tagihan: %s %s\nID Pesanan: %s", currency, amount, orderID)),
		},
		Footer: &waE2E.InteractiveMessage_Footer{
			Text: proto.String("Dongtube Payment System"),
		},
		InteractiveMessage: &waE2E.InteractiveMessage_NativeFlowMessage_{
			NativeFlowMessage: &waE2E.InteractiveMessage_NativeFlowMessage{
				Buttons: []*waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
					{
						Name:             proto.String("review_and_pay"),
						ButtonParamsJSON: proto.String(params),
					},
				},
			},
		},
	}

	return &waE2E.Message{
		ViewOnceMessage: &waE2E.FutureProofMessage{
			Message: &waE2E.Message{
				InteractiveMessage: interactive,
			},
		},
	}, nil
}
