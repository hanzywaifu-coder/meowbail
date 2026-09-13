package messaging

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	protocol "github.com/hanzywaifu-coder/dongtube-meowbail/internal/protocol"
	"time"

	groups "github.com/hanzywaifu-coder/dongtube-meowbail/pkg/groups"
	media "github.com/hanzywaifu-coder/dongtube-meowbail/pkg/media"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

// SendRawCDNMenuWithContent mengirim menu native flow A2UI persis CDN Dongtube dengan body teks konten custom
func SendRawCDNMenuWithContent(c *protocol.Client, ctx context.Context, chat protocol.JID, thumbBytes []byte, headerTitle, customBodyText string, customCtx ...*waE2E.ContextInfo) error {
	secretBytes, _ := base64.StdEncoding.DecodeString("IXM8c2x6FioXJZLibSUFkhheds8R4KQtoWqKWvhwIkY=")

	catParamsJSON := `{"title":"\u0000","sections":[{"title":"ᴋᴀᴛᴇɢᴏʀɪ","highlight_label":"ᴍᴏʀᴇʟᴀ ᴍᴇɴᴜ","rows":[{"title":"ᴀɪ","description":"3 ᴄᴏᴍᴍᴀɴᴅ","id":".menu_ai"},{"title":"ᴅᴏᴡɴʟᴏᴀᴅᴇʀ","description":"12 ᴄᴏᴍᴍᴀɴᴅ","id":".menu_downloader"},{"title":"ꜱᴛɪᴄᴋᴇʀ","description":"13 ᴄᴏᴍᴍᴀɴᴅ","id":".menu_sticker"},{"title":"ᴍᴀᴋᴇʀ","description":"3 ᴄᴏᴍᴍᴀɴᴅ","id":".menu_maker"},{"title":"ᴛᴏᴏʟꜱ","description":"26 ᴄᴏᴍᴍᴀɴᴅ","id":".menu_tools"},{"title":"ɢᴀᴍᴇꜱ","description":"72 ᴄᴏᴍᴍᴀɴᴅ","id":".menu_games"},{"title":"ɴꜱꜰᴡ","description":"40 ᴄᴏᴍᴍᴀɴᴅ","id":".menu_nsfw"},{"title":"ɪɴꜰᴏ","description":"6 ᴄᴏᴍᴍᴀɴᴅ","id":".menu_info"},{"title":"ᴀᴅᴍɪɴ","description":"24 ᴄᴏᴍᴍᴀɴᴅ","id":".menu_admin"},{"title":"ᴏᴡɴᴇʀ","description":"41 ᴄᴏᴍᴍᴀɴᴅ","id":".menu_owner"},{"title":"ꜰᴜɴ","description":"3 ᴄᴏᴍᴍᴀɴᴅ","id":".menu_fun"}]}],"icon":"DEFAULT"}`

	infoParamsJSON := `{"title":"\u0000","sections":[{"title":"ɪɴꜰᴏʀᴍᴀꜱɪ","highlight_label":"ɪɴꜰᴏʀᴍᴀꜱɪ","rows":[{"title":"ᴘɪɴɢ","id":".ping"},{"title":"ᴏᴡɴᴇʀ","id":".menu_owner"}]}],"icon":"REVIEW"}`

	ctaParamsJSON := `{"display_text":"\u0000","url":"https://wa.me/6283143961588","merchant_url":"https://wa.me/6283143961588","icon":"PROMOTION"}`

	// Ambil waktu saat ini + 30 hari dalam millisecond agar limited_time_offer selalu aktif & fresh
	expTime := time.Now().Add(30 * 24 * time.Hour).UnixMilli()
	offerMap := map[string]interface{}{
		"limited_time_offer": map[string]interface{}{
			"text":            "",
			"url":             "https://wa.me/6283143961588",
			"copy_code":       "Dongtube",
			"expiration_time": expTime,
		},
	}
	offerParamsBytes, _ := json.Marshal(offerMap)
	offerParamsJSON := string(offerParamsBytes)

	escapedBody, _ := json.Marshal(customBodyText)
	// Buat A2UI Bloks Widget dengan teks isi kategori
	bloksDataRaw := fmt.Sprintf(`{"version":"v0.9","createSurface":{"surfaceId":"starcore-widget-cat","catalogId":"https://a2ui.org/specification/v0_9/catalogs/basic/catalog.json","components":[{"id":"root","component":"Column","children":["card_2","button_8"]},{"id":"text_0","component":"Text","text":%s,"variant":"body"},{"id":"column_1","component":"Column","children":["text_0"]},{"id":"card_2","component":"Card","child":"column_1"},{"id":"text_7","component":"Text","text":"ᴏᴡɴᴇʀ","variant":"body"},{"id":"button_8","component":"Button","child":"text_7","variant":"primary","action":{"call":"openUrl","args":{"url":"https://wa.me/6283143961588"}}}]}}`, string(escapedBody))

	header := &waE2E.InteractiveMessage_Header{
		HasMediaAttachment: proto.Bool(true),
	}
	if len(thumbBytes) > 0 {
		uploaded, err := media.UploadMedia(c, ctx, thumbBytes, whatsmeow.MediaImage)
		if err == nil && uploaded != nil && uploaded.DirectPath != "" {
			header.Media = &waE2E.InteractiveMessage_Header_ImageMessage{
				ImageMessage: &waE2E.ImageMessage{
					URL:           proto.String(uploaded.URL),
					DirectPath:    proto.String(uploaded.DirectPath),
					MediaKey:      uploaded.MediaKey,
					Mimetype:      proto.String("image/jpeg"),
					FileEncSHA256: uploaded.FileEncSHA256,
					FileSHA256:    uploaded.FileSHA256,
					FileLength:    proto.Uint64(uploaded.FileLength),
					JPEGThumbnail: thumbBytes,
				},
			}
		}
	}

	if header.Media == nil {
		header.Media = &waE2E.InteractiveMessage_Header_LocationMessage{
			LocationMessage: &waE2E.LocationMessage{
				DegreesLatitude:  proto.Float64(0),
				DegreesLongitude: proto.Float64(0),
				Name:             proto.String(headerTitle),
				Address:          proto.String(""),
				JPEGThumbnail:    thumbBytes,
			},
		}
	}

	interactiveMsg := &waE2E.InteractiveMessage{
		Header: header,
		Body: &waE2E.InteractiveMessage_Body{
			Text: proto.String("\u0000"),
		},
		Footer: &waE2E.InteractiveMessage_Footer{
			Text: proto.String("Dongtube"),
		},
		InteractiveMessage: &waE2E.InteractiveMessage_NativeFlowMessage_{
			NativeFlowMessage: &waE2E.InteractiveMessage_NativeFlowMessage{
				Buttons: []*waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
					{Name: proto.String("")},
					{Name: proto.String("single_select"), ButtonParamsJSON: proto.String(catParamsJSON)},
					{Name: proto.String("single_select"), ButtonParamsJSON: proto.String(infoParamsJSON)},
					{Name: proto.String("cta_url"), ButtonParamsJSON: proto.String(ctaParamsJSON)},
				},
				MessageParamsJSON: proto.String(offerParamsJSON),
			},
		},
		BloksWidget: &waE2E.InteractiveMessage_BloksWidget{
			Uuid: proto.String("b4e9d374-935e-4f69-9329-73a548d05b67"),
			Data: proto.String(bloksDataRaw),
			Type: proto.String("im_a2ui"),
		},
		ContextInfo: func() *waE2E.ContextInfo {
			if len(customCtx) > 0 && customCtx[0] != nil {
				return customCtx[0]
			}
			return BuildNewsletterContext(c.Config())
		}(),
	}

	// Pastikan InteractiveMessage selalu membawa fake reply & newsletter
	if c.Config() != nil && c.Config().DefaultFakeReply != nil {
		interactiveMsg.ContextInfo = proto.Clone(c.Config().DefaultFakeReply).(*waE2E.ContextInfo)
		if len(customCtx) > 0 && customCtx[0] != nil {
			if customCtx[0].QuotedMessage != nil {
				interactiveMsg.ContextInfo.QuotedMessage = customCtx[0].QuotedMessage
			}
		}
	}

	msg := &waE2E.Message{
		MessageContextInfo: &waE2E.MessageContextInfo{
			ThreadID:      make([]*waE2E.ThreadID, 0),
			MessageSecret: secretBytes,
		},
		ViewOnceMessage: &waE2E.FutureProofMessage{
			Message: &waE2E.Message{
				InteractiveMessage: interactiveMsg,
			},
		},
	}

	bizNodes := groups.BuildBizAdditionalNodes()
	_, err := c.SendMessageWithRetry(ctx, chat, msg, 3, whatsmeow.SendRequestExtra{
		AdditionalNodes: &bizNodes,
	})
	return err
}
