package privacy

import (
	"context"
	protocol "github.com/hanzywaifu-coder/dongtube-meowbail/internal/protocol"
)

// GetContactInfo mengambil informasi nama dan detail kontak dari local store whatsmeow
func GetContactInfo(c *protocol.Client, ctx context.Context, jid protocol.JID) (*protocol.ContactInfo, error) {
	if c.Client == nil || c.Client.Store == nil || c.Client.Store.Contacts == nil {
		return &protocol.ContactInfo{Found: false}, nil
	}
	info, err := c.Client.Store.Contacts.GetContact(ctx, jid)
	return &info, err
}

// GetAllContacts mengambil seluruh daftar kontak yang tersimpan di database lokal
func GetAllContacts(c *protocol.Client, ctx context.Context) (map[protocol.JID]protocol.ContactInfo, error) {
	if c.Client == nil || c.Client.Store == nil || c.Client.Store.Contacts == nil {
		return make(map[protocol.JID]protocol.ContactInfo), nil
	}
	return c.Client.Store.Contacts.GetAllContacts(ctx)
}

// GetUserDevices mengambil daftar perangkat terhubung (multi-device) untuk kumpulan kontak WhatsApp
// Parity dengan Baileys getUSyncDevices / fbid:devices
func GetUserDevices(c *protocol.Client, ctx context.Context, jids []protocol.JID) ([]protocol.JID, error) {
	if len(jids) == 0 {
		return nil, nil
	}
	return c.Client.DangerousInternals().GetFBIDDevices(ctx, jids)
}
