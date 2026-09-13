package protocol

import (
	"context"
	waBinary "go.mau.fi/whatsmeow/binary"
	"go.mau.fi/whatsmeow/types"
)

// CleanDirtyBits membersihkan bit sinkronisasi kotor (dirty bits) untuk sinkronisasi akun atau grup
// Mengikuti Baileys cleanDirtyBits (iq set xmlns="urn:xmpp:whatsapp:dirty")
func (c *Client) CleanDirtyBits(ctx context.Context, dirtyType string) error {
	cleanNode := waBinary.Node{
		Tag: "clean",
		Attrs: waBinary.Attrs{
			"type": dirtyType,
		},
	}
	queryNode := waBinary.Node{
		Tag: "iq",
		Attrs: waBinary.Attrs{
			"to":    types.ServerJID.String(),
			"type":  "set",
			"xmlns": "urn:xmpp:whatsapp:dirty",
		},
		Content: []waBinary.Node{cleanNode},
	}
	return c.Client.DangerousInternals().SendNode(ctx, queryNode)
}

// GetServerPreKeyCount mengambil jumlah sisa pre-keys yang masih tersedia di server WhatsApp
// Parity dengan Baileys getAvailablePreKeysOnServer (iq xmlns="encrypt" type="get" <count/>)
func (c *Client) GetServerPreKeyCount(ctx context.Context) (int, error) {
	return c.Client.DangerousInternals().GetServerPreKeyCount(ctx)
}

// UploadPreKeys mengunggah pre-keys baru ke server WhatsApp
func (c *Client) UploadPreKeys(ctx context.Context, initialUpload bool) {
	c.Client.DangerousInternals().UploadPreKeys(ctx, initialUpload)
}
