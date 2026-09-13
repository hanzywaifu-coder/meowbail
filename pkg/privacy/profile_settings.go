package privacy

import (
	"context"
	protocol "github.com/hanzywaifu-coder/dongtube-meowbail/internal/protocol"
)

// SetGroupPhoto mengubah foto/ikon grup (avatar)
func SetGroupPhoto(c *protocol.Client, ctx context.Context, groupJID protocol.JID, jpegData []byte) (string, error) {
	return c.Client.SetGroupPhoto(ctx, groupJID, jpegData)
}

// RemoveGroupPhoto menghapus foto/ikon grup
func RemoveGroupPhoto(c *protocol.Client, ctx context.Context, groupJID protocol.JID) error {
	_, err := c.Client.SetGroupPhoto(ctx, groupJID, nil)
	return err
}

// SetGroupJoinApprovalMode mengatur persetujuan admin untuk member baru yang mau bergabung (approve mode)
func SetGroupJoinApprovalMode(c *protocol.Client, ctx context.Context, groupJID protocol.JID, requireApproval bool) error {
	return c.Client.SetGroupJoinApprovalMode(ctx, groupJID, requireApproval)
}

// SetGroupMemberAddMode mengatur siapa saja yang berhak menambahkan member ke grup (admin only atau all members)
func SetGroupMemberAddMode(c *protocol.Client, ctx context.Context, groupJID protocol.JID, adminOnly bool) error {
	mode := protocol.GroupMemberAddModeAllMember
	if adminOnly {
		mode = protocol.GroupMemberAddModeAdmin
	}
	return c.Client.SetGroupMemberAddMode(ctx, groupJID, mode)
}

// SetAboutStatus mengubah status teks bio profil akun WhatsApp ("About")
func SetAboutStatus(c *protocol.Client, ctx context.Context, statusText string) error {
	return c.Client.SetStatusMessage(ctx, protocol.SetStatusInput{
		Text: &statusText,
	})
}

// GetUserProfiles mengambil profil lengkap satu atau beberapa nomor (status bio, avatar ID, verified business name, daftar perangkat)
func GetUserProfiles(c *protocol.Client, ctx context.Context, jids []protocol.JID) (map[protocol.JID]protocol.UserInfo, error) {
	return c.Client.GetUserInfo(ctx, jids)
}

// GetUserProfile mengambil profil tunggal seorang pengguna
func GetUserProfile(c *protocol.Client, ctx context.Context, jid protocol.JID) (*protocol.UserInfo, error) {
	res, err := c.Client.GetUserInfo(ctx, []protocol.JID{jid})
	if err != nil {
		return nil, err
	}
	if info, ok := res[jid]; ok {
		return &info, nil
	}
	return nil, nil
}
