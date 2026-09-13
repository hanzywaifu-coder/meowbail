package messaging

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	protocol "github.com/hanzywaifu-coder/dongtube-meowbail/internal/protocol"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

func randHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// FakeQuotedBuilder penyedia aneka fake reply bergaya dongtube-meowbail
type FakeQuotedBuilder struct{}

var FakeReply = &FakeQuotedBuilder{}

// Text membuat fake reply teks status broadcast (qtext / qtext2)
func (f *FakeQuotedBuilder) Text(text string) *waE2E.ContextInfo {
	return &waE2E.ContextInfo{
		StanzaID:    proto.String("DONGTUBE" + randHex(8)),
		Participant: proto.String("0@s.whatsapp.net"),
		RemoteJID:   proto.String("status@broadcast"),
		QuotedMessage: &waE2E.Message{
			ExtendedTextMessage: &waE2E.ExtendedTextMessage{
				Text: proto.String(text),
			},
		},
	}
}

// MetaAI membuat fake reply pesan seolah dijawab oleh Meta AI resmi WhatsApp
func (f *FakeQuotedBuilder) MetaAI(promptText string) *waE2E.ContextInfo {
	return &waE2E.ContextInfo{
		StanzaID:    proto.String("DONGTUBE" + randHex(8)),
		Participant: proto.String("13135550002@s.whatsapp.net"),
		RemoteJID:   proto.String("status@broadcast"),
		QuotedMessage: &waE2E.Message{
			ExtendedTextMessage: &waE2E.ExtendedTextMessage{
				Text: proto.String(promptText),
			},
		},
	}
}

// Verified membuat fake reply akun verified / WhatsApp Official Support
func (f *FakeQuotedBuilder) Verified(titleText string) *waE2E.ContextInfo {
	return &waE2E.ContextInfo{
		StanzaID:    proto.String("DONGTUBE" + randHex(8)),
		Participant: proto.String("0@s.whatsapp.net"),
		RemoteJID:   proto.String("status@broadcast"),
		QuotedMessage: &waE2E.Message{
			ExtendedTextMessage: &waE2E.ExtendedTextMessage{
				Text: proto.String(titleText),
			},
		},
	}
}

// Audio membuat fake reply audio/voice note rekaman
func (f *FakeQuotedBuilder) Audio(durationSec uint32) *waE2E.ContextInfo {
	if durationSec == 0 {
		durationSec = 3600
	}
	ptt := true
	return &waE2E.ContextInfo{
		StanzaID:    proto.String("DONGTUBE" + randHex(8)),
		Participant: proto.String("0@s.whatsapp.net"),
		RemoteJID:   proto.String("status@broadcast"),
		QuotedMessage: &waE2E.Message{
			AudioMessage: &waE2E.AudioMessage{
				Seconds:  proto.Uint32(durationSec),
				PTT:      &ptt,
				Mimetype: proto.String("audio/ogg; codecs=opus"),
			},
		},
	}
}

// Video membuat fake reply video status/story
func (f *FakeQuotedBuilder) Video(caption string, durationSec uint32, thumb []byte) *waE2E.ContextInfo {
	if durationSec == 0 {
		durationSec = 120
	}
	return &waE2E.ContextInfo{
		StanzaID:    proto.String("DONGTUBE" + randHex(8)),
		Participant: proto.String("0@s.whatsapp.net"),
		RemoteJID:   proto.String("status@broadcast"),
		QuotedMessage: &waE2E.Message{
			VideoMessage: &waE2E.VideoMessage{
				Seconds:       proto.Uint32(durationSec),
				Caption:       proto.String(caption),
				JPEGThumbnail: thumb,
			},
		},
	}
}

// Document membuat fake reply dokumen lampiran (qdoc)
func (f *FakeQuotedBuilder) Document(fileName, pageCount uint32, thumb []byte) *waE2E.ContextInfo {
	return &waE2E.ContextInfo{
		StanzaID:    proto.String("DONGTUBE" + randHex(8)),
		Participant: proto.String("0@s.whatsapp.net"),
		RemoteJID:   proto.String("status@broadcast"),
		QuotedMessage: &waE2E.Message{
			DocumentMessage: &waE2E.DocumentMessage{
				Title:         proto.String("Document"),
				FileName:      proto.String(fmt.Sprintf("%d.pdf", fileName)),
				PageCount:     proto.Uint32(pageCount),
				JPEGThumbnail: thumb,
				Mimetype:      proto.String("application/pdf"),
			},
		},
	}
}

// Saluran membuat fake reply berasal dari Saluran/Newsletter resmi
func (f *FakeQuotedBuilder) Saluran(channelJID, channelName, postText string) *waE2E.ContextInfo {
	return &waE2E.ContextInfo{
		StanzaID:        proto.String("DONGTUBE" + randHex(8)),
		Participant:     proto.String("0@s.whatsapp.net"),
		RemoteJID:       proto.String("status@broadcast"),
		IsForwarded:     proto.Bool(false),
		ForwardingScore: proto.Uint32(0),
		ForwardedNewsletterMessageInfo: &waE2E.ContextInfo_ForwardedNewsletterMessageInfo{
			NewsletterJID:   proto.String(channelJID),
			NewsletterName:  proto.String(channelName),
			ServerMessageID: proto.Int32(100),
		},
		QuotedMessage: &waE2E.Message{
			ExtendedTextMessage: &waE2E.ExtendedTextMessage{
				Text: proto.String(postText),
			},
		},
	}
}

// Polling membuat fake reply polling pilihan interaktif
func (f *FakeQuotedBuilder) Polling(question string, options []string) *waE2E.ContextInfo {
	var pollOpts []*waE2E.PollCreationMessage_Option
	for _, opt := range options {
		pollOpts = append(pollOpts, &waE2E.PollCreationMessage_Option{
			OptionName: proto.String(opt),
		})
	}
	return &waE2E.ContextInfo{
		StanzaID:    proto.String("DONGTUBE" + randHex(8)),
		Participant: proto.String("0@s.whatsapp.net"),
		RemoteJID:   proto.String("status@broadcast"),
		QuotedMessage: &waE2E.Message{
			PollCreationMessage: &waE2E.PollCreationMessage{
				Name:                   proto.String(question),
				Options:                pollOpts,
				SelectableOptionsCount: proto.Uint32(1),
			},
		},
	}
}

// Location membuat fake reply lokasi broadcast (qlocJpm / qlocPush)
func (f *FakeQuotedBuilder) Location(name string, thumb []byte) *waE2E.ContextInfo {
	return &waE2E.ContextInfo{
		StanzaID:    proto.String("DONGTUBE" + randHex(8)),
		Participant: proto.String("0@s.whatsapp.net"),
		RemoteJID:   proto.String("status@broadcast"),
		QuotedMessage: &waE2E.Message{
			LocationMessage: &waE2E.LocationMessage{
				DegreesLatitude:  proto.Float64(0),
				DegreesLongitude: proto.Float64(0),
				Name:             proto.String(name),
				JPEGThumbnail:    thumb,
			},
		},
	}
}

// LiveLocation membuat fake reply live location broadcast (qlive)
func (f *FakeQuotedBuilder) LiveLocation(caption string, thumb []byte) *waE2E.ContextInfo {
	return &waE2E.ContextInfo{
		StanzaID:    proto.String("DONGTUBE" + randHex(8)),
		Participant: proto.String("0@s.whatsapp.net"),
		RemoteJID:   proto.String("status@broadcast"),
		QuotedMessage: &waE2E.Message{
			LiveLocationMessage: &waE2E.LiveLocationMessage{
				Caption:       proto.String(caption),
				JPEGThumbnail: thumb,
			},
		},
	}
}

// Payment membuat fake reply payment request (qpayment)
func (f *FakeQuotedBuilder) Payment(botName string, amount1000 uint64, currency string) *waE2E.ContextInfo {
	if currency == "" {
		currency = "USD"
	}
	if amount1000 == 0 {
		amount1000 = 999999999
	}
	return &waE2E.ContextInfo{
		StanzaID:    proto.String("ownername"),
		Participant: proto.String("0@s.whatsapp.net"),
		RemoteJID:   proto.String("0@s.whatsapp.net"),
		QuotedMessage: &waE2E.Message{
			RequestPaymentMessage: &waE2E.RequestPaymentMessage{
				CurrencyCodeIso4217: proto.String(currency),
				Amount1000:          proto.Uint64(amount1000),
				RequestFrom:         proto.String("0@s.whatsapp.net"),
				NoteMessage: &waE2E.Message{
					ExtendedTextMessage: &waE2E.ExtendedTextMessage{
						Text: proto.String(botName),
					},
				},
				ExpiryTimestamp: proto.Int64(999999999),
			},
		},
	}
}

// Toko membuat fake reply produk katalog marketplace (qtoko)
func (f *FakeQuotedBuilder) Toko(title, retailerId string, thumb []byte) *waE2E.ContextInfo {
	return &waE2E.ContextInfo{
		StanzaID:    proto.String("DONGTUBE" + randHex(8)),
		Participant: proto.String("0@s.whatsapp.net"),
		RemoteJID:   proto.String("status@broadcast"),
		QuotedMessage: &waE2E.Message{
			ProductMessage: &waE2E.ProductMessage{
				Product: &waE2E.ProductMessage_ProductSnapshot{
					ProductImage: &waE2E.ImageMessage{
						Mimetype:      proto.String("image/jpeg"),
						JPEGThumbnail: thumb,
					},
					Title:             proto.String(title),
					CurrencyCode:      proto.String("IDR"),
					PriceAmount1000:   proto.Int64(999999999999999),
					RetailerID:        proto.String(retailerId),
					ProductImageCount: proto.Uint32(1),
				},
				BusinessOwnerJID: proto.String("0@s.whatsapp.net"),
			},
		},
	}
}

// Troli membuat fake reply troli belanja pengguna (troli)
func (f *FakeQuotedBuilder) Troli(senderJID protocol.JID, title, runtimeText string, thumb []byte) *waE2E.ContextInfo {
	status := waE2E.OrderMessage_INQUIRY
	surface := waE2E.OrderMessage_CATALOG
	orderID := "DONGTUBE" + randHex(8)
	return &waE2E.ContextInfo{
		StanzaID:    proto.String(orderID),
		Participant: proto.String("0@s.whatsapp.net"),
		RemoteJID:   proto.String("status@broadcast"),
		QuotedMessage: &waE2E.Message{
			OrderMessage: &waE2E.OrderMessage{
				OrderID:           proto.String(orderID),
				ItemCount:         proto.Int32(2026),
				Status:            &status,
				Surface:           &surface,
				Message:           proto.String(runtimeText),
				OrderTitle:        proto.String(title),
				SellerJID:         proto.String("0@s.whatsapp.net"),
				Thumbnail:         thumb,
				Token:             proto.String("AR6xBKmme9otv9WMZ4O6L9p968T2v99A"),
				TotalAmount1000:   proto.Int64(50000000000),
				TotalCurrencyCode: proto.String("IDR"),
				MessageVersion:    proto.Int32(1),
			},
		},
	}
}

// Kontak membuat fake reply kartu nama kontak developer (qkontak)
func (f *FakeQuotedBuilder) Kontak(displayName, phone string) *waE2E.ContextInfo {
	vcard := "BEGIN:VCARD\nVERSION:3.0\nFN:" + displayName + "\nORG:Developer\nTEL;type=CELL;type=VOICE;waid=" + phone + ":+" + phone + "\nEND:VCARD"
	return &waE2E.ContextInfo{
		StanzaID:    proto.String("DONGTUBE" + randHex(8)),
		Participant: proto.String("0@s.whatsapp.net"),
		RemoteJID:   proto.String("status@broadcast"),
		QuotedMessage: &waE2E.Message{
			ContactMessage: &waE2E.ContactMessage{
				DisplayName: proto.String(displayName),
				Vcard:       proto.String(vcard),
			},
		},
	}
}

// CustomQuote membuat custom fake reply untuk target user apa pun
func (f *FakeQuotedBuilder) CustomQuote(remoteJID, participantJID string, msg *waE2E.Message) *waE2E.ContextInfo {
	return &waE2E.ContextInfo{
		StanzaID:      proto.String("DONGTUBE" + randHex(8)),
		Participant:   proto.String(participantJID),
		RemoteJID:     proto.String(remoteJID),
		QuotedMessage: msg,
	}
}

// SendTextWithFakeReply mengirim pesan teks menggunakan fake reply context
func SendTextWithFakeReply(c *protocol.Client, ctx context.Context, chat protocol.JID, text string, fakeContext *waE2E.ContextInfo) error {
	if fakeContext == nil {
		fakeContext = &waE2E.ContextInfo{}
	}

	// Auto-extract @nomor dari teks jika fakeContext belum memiliki MentionedJID
	if len(fakeContext.MentionedJID) == 0 {
		mentions := ExtractMentions(text)
		if len(mentions) > 0 {
			var mList []string
			for _, m := range mentions {
				norm := m.ToNonAD()
				if norm.Server == protocol.HiddenUserServer || norm.Server == "lid" {
					if c.LIDResolver != nil {
						resolved := c.LIDResolver.ResolveToPN(norm)
						if resolved.Server == protocol.DefaultUserServer {
							norm = resolved
						}
					}
				}
				mList = append(mList, norm.String())
			}
			fakeContext.MentionedJID = mList
		}
	}

	// Gabungkan newsletter context jika ada
	if c.Config() != nil && c.Config().NewsletterJID != "" {
		fakeContext.IsForwarded = proto.Bool(false)
		fakeContext.ForwardingScore = proto.Uint32(0)
		fakeContext.BusinessMessageForwardInfo = &waE2E.ContextInfo_BusinessMessageForwardInfo{
			BusinessOwnerJID: proto.String(c.Config().BusinessOwnerJID),
		}
		fakeContext.ForwardedNewsletterMessageInfo = &waE2E.ContextInfo_ForwardedNewsletterMessageInfo{
			NewsletterJID:  proto.String(c.Config().NewsletterJID),
			NewsletterName: proto.String(c.Config().NewsletterName),
		}
	}

	msg := &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text:        proto.String(text),
			ContextInfo: fakeContext,
		},
	}

	_, err := c.Client.SendMessage(ctx, chat, msg)
	return err
}
