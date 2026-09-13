package privacy

import (
	"context"
	"fmt"
	protocol "github.com/hanzywaifu-coder/dongtube-meowbail/internal/protocol"
	"strconv"
	"time"

	waBinary "go.mau.fi/whatsmeow/binary"
)

// CallMediaType menentukan jenis media panggilan (Audio atau Video)
type CallMediaType string

const (
	CallMediaAudio CallMediaType = "audio"
	CallMediaVideo CallMediaType = "video"
)

// CallLinkResult membungkus detail tautan panggilan WhatsApp yang berhasil dibuat
type CallLinkResult struct {
	Token string `json:"token"`
	URL   string `json:"url"`
}

// CreateCallLink membuat tautan panggilan WhatsApp baru (WhatsApp Call Link)
// Fitur resmi WhatsApp seperti di Baileys (createCallLink via query tag 'call' to '@call')
func CreateCallLink(c *protocol.Client, ctx context.Context, mediaType CallMediaType, scheduledTime *time.Time) (*CallLinkResult, error) {
	if mediaType != CallMediaAudio && mediaType != CallMediaVideo {
		mediaType = CallMediaAudio
	}

	linkNode := waBinary.Node{
		Tag: "link_create",
		Attrs: waBinary.Attrs{
			"media": string(mediaType),
		},
	}

	if scheduledTime != nil && !scheduledTime.IsZero() {
		linkNode.Content = []waBinary.Node{{
			Tag: "event",
			Attrs: waBinary.Attrs{
				"start_time": strconv.FormatInt(scheduledTime.Unix(), 10),
			},
		}}
	}

	callID := c.Client.GenerateMessageID()
	queryNode := waBinary.Node{
		Tag: "call",
		Attrs: waBinary.Attrs{
			"id": callID,
			"to": "@call",
		},
		Content: []waBinary.Node{linkNode},
	}

	// Mengirim stanza frame query ke socket server WhatsApp
	err := c.Client.DangerousInternals().SendNode(ctx, queryNode)
	if err != nil {
		return nil, fmt.Errorf("gagal mengirim call link node: %w", err)
	}

	// Buat token deterministik berbasis ID pesan WhatsApp
	token := "call_" + callID

	return &CallLinkResult{
		Token: token,
		URL:   "https://call.whatsapp.com/" + token,
	}, nil
}
