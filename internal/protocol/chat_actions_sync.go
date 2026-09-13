package protocol

import (
	"context"
	"go.mau.fi/whatsmeow/appstate"
	"go.mau.fi/whatsmeow/types"
	"time"
)

// StarMessage membintangi (star) atau menghapus bintang (unstar) pesan tertentu (App State Sync)
// seperti pada Baileys sock.star(jid, [{id, fromMe}], star)
func (c *Client) StarMessage(ctx context.Context, chat types.JID, sender types.JID, messageID types.MessageID, fromMe, star bool) error {
	if sender.IsEmpty() {
		sender = chat
	}
	patch := appstate.BuildStar(chat, sender, messageID, fromMe, star)
	return c.Client.SendAppState(ctx, patch)
}

// DeleteChatForMe menghapus riwayat obrolan untuk akun ini (Delete Chat For Me via App State Sync)
func (c *Client) DeleteChatForMe(ctx context.Context, chat types.JID, deleteMedia bool) error {
	patch := appstate.BuildDeleteChat(chat, time.Now(), nil, deleteMedia)
	return c.Client.SendAppState(ctx, patch)
}

// UpdatePushNameSetting memperbarui pushname akun secara global melalui sinkronisasi App State
func (c *Client) UpdatePushNameSetting(ctx context.Context, pushName string) error {
	patch := appstate.BuildSettingPushName(pushName)
	return c.Client.SendAppState(ctx, patch)
}
