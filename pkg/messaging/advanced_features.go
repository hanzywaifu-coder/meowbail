package messaging

import (
	"context"
	"fmt"
	protocol "github.com/hanzywaifu-coder/dongtube-meowbail/internal/protocol"
	"time"

	groups "github.com/hanzywaifu-coder/dongtube-meowbail/pkg/groups"
	media "github.com/hanzywaifu-coder/dongtube-meowbail/pkg/media"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

// NewsletterReaction mengirim reaksi emoji ke pesan dalam Saluran/Newsletter WhatsApp
func NewsletterReaction(c *protocol.Client, ctx context.Context, channelJID protocol.JID, serverID protocol.MessageServerID, emoji string) error {
	return c.Client.NewsletterSendReaction(ctx, channelJID, serverID, emoji, "")
}

// LiveLocation mengirim lokasi real-time dengan update koordinat berkala
func SendLiveLocation(c *protocol.Client, ctx context.Context, chat protocol.JID, lat, lng float64, accuracy uint32, speed float32, caption string, duration time.Duration) error {
	if duration == 0 {
		duration = 300 * time.Second
	}
	seqNumber := int64(1)
	msg := &waE2E.Message{
		LiveLocationMessage: &waE2E.LiveLocationMessage{
			DegreesLatitude:                   proto.Float64(lat),
			DegreesLongitude:                  proto.Float64(lng),
			AccuracyInMeters:                  proto.Uint32(accuracy),
			SpeedInMps:                        proto.Float32(speed),
			DegreesClockwiseFromMagneticNorth: proto.Uint32(0),
			Caption:                           proto.String(caption),
			SequenceNumber:                    proto.Int64(seqNumber),
			TimeOffset:                        proto.Uint32(uint32(duration.Seconds())),
		},
	}
	_, err := c.Client.SendMessage(ctx, chat, msg)
	return err
}

// ForwardMessage meneruskan pesan yang ada ke obrolan lain dengan forwarding context
func ForwardMessage(c *protocol.Client, ctx context.Context, toChat protocol.JID, rawMsg *waE2E.Message, isMultiForward bool) error {
	if rawMsg == nil {
		return fmt.Errorf("pesan kosong")
	}

	clone := proto.Clone(rawMsg).(*waE2E.Message)
	_ = isMultiForward

	ctxInfo := &waE2E.ContextInfo{
		IsForwarded:     proto.Bool(false),
		ForwardingScore: proto.Uint32(0),
	}

	// Sisipkan context info ke jenis pesan yang sesuai
	switch {
	case clone.ExtendedTextMessage != nil:
		clone.ExtendedTextMessage.ContextInfo = ctxInfo
	case clone.ImageMessage != nil:
		clone.ImageMessage.ContextInfo = ctxInfo
	case clone.VideoMessage != nil:
		clone.VideoMessage.ContextInfo = ctxInfo
	case clone.DocumentMessage != nil:
		clone.DocumentMessage.ContextInfo = ctxInfo
	case clone.AudioMessage != nil:
		clone.AudioMessage.ContextInfo = ctxInfo
	case clone.StickerMessage != nil:
		clone.StickerMessage.ContextInfo = ctxInfo
	case clone.Conversation != nil:
		clone = &waE2E.Message{
			ExtendedTextMessage: &waE2E.ExtendedTextMessage{
				Text:        clone.Conversation,
				ContextInfo: ctxInfo,
			},
		}
	}

	_, err := c.Client.SendMessage(ctx, toChat, clone)
	return err
}

// SendPTTVoiceNote mengirim audio rekaman suara PTT (Push To Talk waveform)
func SendPTTVoiceNote(c *protocol.Client, ctx context.Context, chat protocol.JID, oggOpusData []byte, waveform []byte) error {
	uploaded, err := media.UploadMedia(c, ctx, oggOpusData, whatsmeow.MediaAudio)
	if err != nil {
		return err
	}

	if len(waveform) == 0 {
		// Generate 64-sample realistic waveform jika tidak disediakan
		waveform = make([]byte, 64)
		for i := range waveform {
			waveform[i] = byte(20 + (i*7)%60)
		}
	}

	ptt := true
	msg := &waE2E.Message{
		AudioMessage: &waE2E.AudioMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			Mimetype:      proto.String("audio/ogg; codecs=opus"),
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uploaded.FileLength),
			PTT:           &ptt,
			Waveform:      waveform,
		},
	}

	_, err = c.Client.SendMessage(ctx, chat, msg)
	return err
}

// RequestPhoneNumber mengirim prompt tombol Native Flow request phone number WhatsApp
func RequestPhoneNumber(c *protocol.Client, ctx context.Context, chat protocol.JID, bodyText string) error {
	btnName := "cta_phone_number"
	btnParams := `{"display_text":"Bagikan Nomor Saya"}`

	msg := &waE2E.Message{
		InteractiveMessage: &waE2E.InteractiveMessage{
			Body: &waE2E.InteractiveMessage_Body{
				Text: proto.String(bodyText),
			},
			InteractiveMessage: &waE2E.InteractiveMessage_NativeFlowMessage_{
				NativeFlowMessage: &waE2E.InteractiveMessage_NativeFlowMessage{
					Buttons: []*waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
						{
							Name:             proto.String(btnName),
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

// RequestLiveLocationPrompt meminta user mengirim live location secara native flow
func RequestLiveLocationPrompt(c *protocol.Client, ctx context.Context, chat protocol.JID, bodyText string) error {
	btnName := "send_location"
	btnParams := `{}`

	msg := &waE2E.Message{
		InteractiveMessage: &waE2E.InteractiveMessage{
			Body: &waE2E.InteractiveMessage_Body{
				Text: proto.String(bodyText),
			},
			InteractiveMessage: &waE2E.InteractiveMessage_NativeFlowMessage_{
				NativeFlowMessage: &waE2E.InteractiveMessage_NativeFlowMessage{
					Buttons: []*waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
						{
							Name:             proto.String(btnName),
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

// SetDisappearingChat mengubah durasi disappearing timer obrolan (24 Jam, 7 Hari, 90 Hari, atau Mati)
func SetDisappearingChat(c *protocol.Client, ctx context.Context, chat protocol.JID, timer time.Duration) error {
	msg := &waE2E.Message{
		ProtocolMessage: &waE2E.ProtocolMessage{
			Type:                      waE2E.ProtocolMessage_EPHEMERAL_SETTING.Enum(),
			EphemeralExpiration:       proto.Uint32(uint32(timer.Seconds())),
			EphemeralSettingTimestamp: proto.Int64(time.Now().Unix()),
		},
	}
	_, err := c.Client.SendMessage(ctx, chat, msg)
	return err
}

// RejectCall menolak panggilan masuk WhatsApp otomatis
func RejectCall(c *protocol.Client, ctx context.Context, callFrom protocol.JID, callID string) error {
	return c.Client.RejectCall(ctx, callFrom, callID)
}

// GetContactProfilePicture mengambil link/buffer foto profil kontak atau grup
func GetContactProfilePicture(c *protocol.Client, ctx context.Context, target protocol.JID, isCommunity bool) (*protocol.ProfilePictureInfo, error) {
	var extra *whatsmeow.GetProfilePictureParams
	if isCommunity {
		extra = &whatsmeow.GetProfilePictureParams{
			IsCommunity: true,
		}
	}
	return c.Client.GetProfilePictureInfo(ctx, target, extra)
}

// DownloadContactProfilePicture mengambil dan mengunduh byte gambar langsung dari avatar pengguna atau grup
func DownloadContactProfilePicture(c *protocol.Client, ctx context.Context, target protocol.JID, isCommunity bool) ([]byte, error) {
	info, err := GetContactProfilePicture(c, ctx, target, isCommunity)
	if err != nil {
		return nil, err
	}
	if info == nil || info.URL == "" {
		return nil, fmt.Errorf("avatar tidak ditemukan")
	}
	return media.FetchURL(info.URL)
}
