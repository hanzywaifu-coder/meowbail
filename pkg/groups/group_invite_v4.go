package groups

import (
	"context"
	"fmt"
	protocol "github.com/hanzywaifu-coder/dongtube-meowbail/internal/protocol"
	"strconv"

	waBinary "go.mau.fi/whatsmeow/binary"
	"go.mau.fi/whatsmeow/proto/waE2E"
)

// GroupAcceptInviteV4 menerima undangan masuk grup v4 (GroupInviteMessage)
// Mengikuti implementasi Baileys groupAcceptInviteV4
func GroupAcceptInviteV4(c *protocol.Client, ctx context.Context, groupJID protocol.JID, inviter protocol.JID, inviteCode string, inviteExpiration int64) (protocol.JID, error) {
	if inviteCode == "" {
		return protocol.EmptyJID, fmt.Errorf("invite code tidak boleh kosong")
	}

	attrs := waBinary.Attrs{
		"code":  inviteCode,
		"admin": inviter.String(),
	}
	if inviteExpiration > 0 {
		attrs["expiration"] = strconv.FormatInt(inviteExpiration, 10)
	}

	queryNode := waBinary.Node{
		Tag:   "accept",
		Attrs: attrs,
	}

	resp, err := c.Client.DangerousInternals().SendGroupIQ(ctx, "set", groupJID, queryNode)
	if err != nil {
		return protocol.EmptyJID, err
	}

	if resp != nil {
		return groupJID, nil
	}

	return groupJID, nil
}

// GroupAcceptInviteV4FromMessage mem-parse langsung waE2E.GroupInviteMessage dan mengeksekusi accept invite
func GroupAcceptInviteV4FromMessage(c *protocol.Client, ctx context.Context, inviter protocol.JID, msg *waE2E.GroupInviteMessage) (protocol.JID, error) {
	if msg == nil {
		return protocol.EmptyJID, fmt.Errorf("pesan group invite kosong")
	}

	groupJID := protocol.ParseJID(msg.GetGroupJID())
	if groupJID.IsEmpty() {
		return protocol.EmptyJID, fmt.Errorf("invalid group JID: %q", msg.GetGroupJID())
	}

	return GroupAcceptInviteV4(c, ctx, groupJID, inviter, msg.GetInviteCode(), msg.GetInviteExpiration())
}

// GroupRevokeInviteV4 mencabut undangan V4 grup untuk partisipan tertentu (Baileys groupRevokeInviteV4 parity)
func GroupRevokeInviteV4(c *protocol.Client, ctx context.Context, groupJID protocol.JID, invitedJID protocol.JID) error {
	queryNode := waBinary.Node{
		Tag: "revoke",
		Content: []waBinary.Node{
			{
				Tag: "participant",
				Attrs: waBinary.Attrs{
					"jid": invitedJID.String(),
				},
			},
		},
	}

	_, err := c.Client.DangerousInternals().SendGroupIQ(ctx, "set", groupJID, queryNode)
	return err
}
