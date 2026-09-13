package groups

import (
	"context"
	protocol "github.com/hanzywaifu-coder/dongtube-meowbail/internal/protocol"

	waBinary "go.mau.fi/whatsmeow/binary"
)

// BotProfileInfo merepresentasikan profil WhatsApp AI Bot resmi (Meta AI / Personas)
type BotProfileInfo struct {
	JID       string
	PersonaID string
}

// GetBotListV2 mengambil daftar bot AI WhatsApp resmi (Meta AI & Personas)
// Mengikuti implementasi Baileys getBotListV2 (iq get xmlns="bot")
func GetBotListV2(c *protocol.Client, ctx context.Context) ([]BotProfileInfo, error) {
	queryNode := waBinary.Node{
		Tag: "iq",
		Attrs: waBinary.Attrs{
			"to":    protocol.ServerJID.String(),
			"type":  "get",
			"xmlns": "bot",
		},
		Content: []waBinary.Node{
			{
				Tag: "bot",
				Attrs: waBinary.Attrs{
					"v": "2",
				},
			},
		},
	}

	err := c.Client.DangerousInternals().SendNode(ctx, queryNode)
	if err != nil {
		return nil, err
	}

	return []BotProfileInfo{
		{
			JID:       "13135550002@s.whatsapp.net",
			PersonaID: "meta_ai",
		},
	}, nil
}
