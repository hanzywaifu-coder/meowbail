package business

import (
	"context"
	protocol "github.com/hanzywaifu-coder/dongtube-meowbail/internal/protocol"
	"strconv"
	"time"

	"go.mau.fi/whatsmeow/appstate"
	"go.mau.fi/whatsmeow/proto/waSyncAction"
)

// LabelChat memberikan atau mencabut label bisnis pada obrolan (WhatsApp Business Labels)
func LabelChat(c *protocol.Client, ctx context.Context, chat protocol.JID, labelID string, labeled bool) error {
	patch := appstate.BuildLabelChat(chat, labelID, labeled)
	return c.Client.SendAppState(ctx, patch)
}

// LabelMessage memberikan atau mencabut label bisnis pada pesan spesifik
func LabelMessage(c *protocol.Client, ctx context.Context, chat protocol.JID, labelID, messageID string, labeled bool) error {
	patch := appstate.BuildLabelMessage(chat, labelID, messageID, labeled)
	return c.Client.SendAppState(ctx, patch)
}

// EditLabel membuat, mengubah, atau menghapus label bisnis (nama & warna ID 0-19)
func EditLabel(c *protocol.Client, ctx context.Context, labelID, labelName string, colorID int32, deleted bool) error {
	patch := appstate.BuildLabelEdit(labelID, labelName, colorID, deleted)
	return c.Client.SendAppState(ctx, patch)
}

// AddOrEditQuickReply membuat atau memperbarui Quick Reply bisnis (WhatsApp Business App State)
func AddOrEditQuickReply(c *protocol.Client, ctx context.Context, shortcut, message string, keywords []string, timestamp string, deleted bool) error {
	if timestamp == "" {
		timestamp = strconv.FormatInt(time.Now().Unix(), 10)
	}
	count := int32(0)
	patch := appstate.PatchInfo{
		Type: appstate.WAPatchRegular,
		Mutations: []appstate.MutationInfo{
			{
				Index:   []string{"quick_reply", timestamp},
				Version: 2,
				Value: &waSyncAction.SyncActionValue{
					QuickReplyAction: &waSyncAction.QuickReplyAction{
						Shortcut: &shortcut,
						Message:  &message,
						Keywords: keywords,
						Count:    &count,
						Deleted:  &deleted,
					},
				},
			},
		},
	}
	return c.Client.SendAppState(ctx, patch)
}

// RemoveQuickReply menghapus Quick Reply bisnis berdasarkan timestamp ID
func RemoveQuickReply(c *protocol.Client, ctx context.Context, timestamp string) error {
	return AddOrEditQuickReply(c, ctx, "", "", nil, timestamp, true)
}
