package privacy

import (
	"context"
	protocol "github.com/hanzywaifu-coder/dongtube-meowbail/internal/protocol"
	"time"

	waBinary "go.mau.fi/whatsmeow/binary"
)

// PrivacyManager menangani konfigurasi privasi akun WhatsApp secara lengkap dan presisi
// mencakup Read Receipts (centang biru), Last Seen, Online Status, Call Add, Group Add, Profile Picture, Status, dan Disappearing Mode default.

// GetPrivacySettings mengambil konfigurasi privasi akun yang sedang aktif saat ini
func GetPrivacySettings(c *protocol.Client, ctx context.Context) (protocol.PrivacySettings, error) {
	ptr, err := c.Client.TryFetchPrivacySettings(ctx, true)
	if err != nil {
		return protocol.PrivacySettings{}, err
	}
	if ptr == nil {
		return c.Client.GetPrivacySettings(ctx), nil
	}
	return *ptr, nil
}

// SetReadReceiptsPrivacy mengatur apakah centang biru (tanda pesan telah dibaca) diaktifkan atau dinonaktifkan
// enabled: true -> centang biru aktif ("all"), false -> centang biru mati ("none")
func SetReadReceiptsPrivacy(c *protocol.Client, ctx context.Context, enabled bool) (protocol.PrivacySettings, error) {
	val := protocol.PrivacySettingAll
	if !enabled {
		val = protocol.PrivacySettingNone
	}
	return c.Client.SetPrivacySetting(ctx, protocol.PrivacySettingTypeReadReceipts, val)
}

// SetLastSeenPrivacy mengatur siapa yang dapat melihat status terakhir dilihat (Last Seen)
// value: protocol.PrivacySettingAll ("all"), protocol.PrivacySettingContacts ("contacts"), protocol.PrivacySettingNone ("none")
func SetLastSeenPrivacy(c *protocol.Client, ctx context.Context, value protocol.PrivacySetting) (protocol.PrivacySettings, error) {
	return c.Client.SetPrivacySetting(ctx, protocol.PrivacySettingTypeLastSeen, value)
}

// SetOnlinePrivacy mengatur siapa yang dapat melihat saat akun sedang Online
// value: protocol.PrivacySettingAll ("all"), protocol.PrivacySettingMatchLastSeen ("match_last_seen")
func SetOnlinePrivacy(c *protocol.Client, ctx context.Context, matchLastSeen bool) (protocol.PrivacySettings, error) {
	val := protocol.PrivacySettingAll
	if matchLastSeen {
		val = protocol.PrivacySettingMatchLastSeen
	}
	return c.Client.SetPrivacySetting(ctx, protocol.PrivacySettingTypeOnline, val)
}

// SetProfilePicturePrivacy mengatur siapa yang dapat melihat foto profil akun
func SetProfilePicturePrivacy(c *protocol.Client, ctx context.Context, value protocol.PrivacySetting) (protocol.PrivacySettings, error) {
	return c.Client.SetPrivacySetting(ctx, protocol.PrivacySettingTypeProfile, value)
}

// SetStatusPrivacy mengatur siapa yang dapat melihat pembaharuan status WhatsApp
func SetStatusPrivacy(c *protocol.Client, ctx context.Context, value protocol.PrivacySetting) (protocol.PrivacySettings, error) {
	return c.Client.SetPrivacySetting(ctx, protocol.PrivacySettingTypeStatus, value)
}

// SetGroupAddPrivacy mengatur siapa yang berhak menambahkan akun ke dalam grup
func SetGroupAddPrivacy(c *protocol.Client, ctx context.Context, value protocol.PrivacySetting) (protocol.PrivacySettings, error) {
	return c.Client.SetPrivacySetting(ctx, protocol.PrivacySettingTypeGroupAdd, value)
}

// SetCallAddPrivacy mengatur siapa yang dapat memanggil akun via telepon/video call WhatsApp
func SetCallAddPrivacy(c *protocol.Client, ctx context.Context, allowEveryone bool) (protocol.PrivacySettings, error) {
	val := protocol.PrivacySettingAll
	if !allowEveryone {
		val = protocol.PrivacySettingKnown
	}
	return c.Client.SetPrivacySetting(ctx, protocol.PrivacySettingTypeCallAdd, val)
}

// SetDefaultDisappearingTimer mengatur durasi pesan sementara default untuk semua obrolan baru
// Durasi umum WhatsApp: 24 jam (24 * time.Hour), 7 hari (7 * 24 * time.Hour), 90 hari (90 * 24 * time.Hour), atau 0 untuk nonaktif.
func SetDefaultDisappearingTimer(c *protocol.Client, ctx context.Context, timer time.Duration) error {
	return c.Client.SetDefaultDisappearingTimer(ctx, timer)
}

// GetStatusPrivacy mengambil pengaturan privasi status / story broadcast WhatsApp
func GetStatusPrivacy(c *protocol.Client, ctx context.Context) ([]protocol.StatusPrivacy, error) {
	return c.Client.GetStatusPrivacy(ctx)
}

// IssuePrivacyTokens mengirim permintaan untuk menerbitkan privacy token bagi target JID (trusted_contact)
func IssuePrivacyTokens(c *protocol.Client, ctx context.Context, jids []protocol.JID) error {
	now := time.Now()
	for _, j := range jids {
		_, err := c.Client.DangerousInternals().IssuePrivacyToken(ctx, j, now)
		if err != nil {
			return err
		}
	}
	return nil
}

// DisappearingDurationInfo hasil kueri disappearing mode duration untuk kontak
type DisappearingDurationInfo struct {
	JID      protocol.JID
	Duration time.Duration
}

// UserStatusInfo hasil kueri status / bio / about pengguna WhatsApp
type UserStatusInfo struct {
	JID    protocol.JID
	Status string
	SetAt  time.Time
}

// FetchDisappearingDuration mengambil durasi disappearing mode aktif dari satu atau lebih kontak melalui protokol USync
// Parity dengan Baileys fetchDisappearingDuration
func FetchDisappearingDuration(c *protocol.Client, ctx context.Context, jids []protocol.JID) ([]DisappearingDurationInfo, error) {
	if len(jids) == 0 {
		return nil, nil
	}

	query := []waBinary.Node{
		{Tag: "disappearing_mode"},
	}

	listNode, err := c.Client.DangerousInternals().Usync(ctx, jids, "query", "interactive", query)
	if err != nil {
		return nil, err
	}

	var results []DisappearingDurationInfo
	for _, userNode := range listNode.GetChildren() {
		if userNode.Tag != "user" {
			continue
		}
		ag := userNode.AttrGetter()
		targetJID := ag.OptionalJIDOrEmpty("jid")
		if targetJID.IsEmpty() {
			targetJID = ag.OptionalJIDOrEmpty("pn_jid")
		}

		dmNode, ok := userNode.GetOptionalChildByTag("disappearing_mode")
		var dur time.Duration
		if ok {
			durationSec := dmNode.AttrGetter().OptionalInt("duration")
			dur = time.Duration(durationSec) * time.Second
		}

		results = append(results, DisappearingDurationInfo{
			JID:      targetJID,
			Duration: dur,
		})
	}

	return results, nil
}

// UsernameInfo hasil kueri username WhatsApp (WhatsApp Usernames feature)
type UsernameInfo struct {
	JID      protocol.JID
	Username string
}

// FetchStatus mengambil status teks (About / Bio) dari satu atau lebih kontak melalui protokol USync
// Parity dengan Baileys fetchStatus
func FetchStatus(c *protocol.Client, ctx context.Context, jids []protocol.JID) ([]UserStatusInfo, error) {
	if len(jids) == 0 {
		return nil, nil
	}

	query := []waBinary.Node{
		{Tag: "status"},
	}

	listNode, err := c.Client.DangerousInternals().Usync(ctx, jids, "query", "interactive", query)
	if err != nil {
		return nil, err
	}

	var results []UserStatusInfo
	for _, userNode := range listNode.GetChildren() {
		if userNode.Tag != "user" {
			continue
		}
		ag := userNode.AttrGetter()
		targetJID := ag.OptionalJIDOrEmpty("jid")
		if targetJID.IsEmpty() {
			targetJID = ag.OptionalJIDOrEmpty("pn_jid")
		}

		statusNode, ok := userNode.GetOptionalChildByTag("status")
		var statusText string
		var setAt time.Time
		if ok {
			contentBytes, isBytes := statusNode.Content.([]byte)
			if isBytes {
				statusText = string(contentBytes)
			}
			tSec := statusNode.AttrGetter().OptionalInt("t")
			if tSec > 0 {
				setAt = time.Unix(int64(tSec), 0)
			}
		}

		results = append(results, UserStatusInfo{
			JID:    targetJID,
			Status: statusText,
			SetAt:  setAt,
		})
	}

	return results, nil
}

// FetchUsername mengambil username publik WhatsApp dari satu atau lebih kontak melalui USync
// Parity dengan Baileys USyncUsernameProtocol
func FetchUsername(c *protocol.Client, ctx context.Context, jids []protocol.JID) ([]UsernameInfo, error) {
	if len(jids) == 0 {
		return nil, nil
	}

	query := []waBinary.Node{
		{Tag: "username"},
	}

	listNode, err := c.Client.DangerousInternals().Usync(ctx, jids, "query", "interactive", query)
	if err != nil {
		return nil, err
	}

	var results []UsernameInfo
	for _, userNode := range listNode.GetChildren() {
		if userNode.Tag != "user" {
			continue
		}
		ag := userNode.AttrGetter()
		targetJID := ag.OptionalJIDOrEmpty("jid")
		if targetJID.IsEmpty() {
			targetJID = ag.OptionalJIDOrEmpty("pn_jid")
		}

		unNode, ok := userNode.GetOptionalChildByTag("username")
		var uname string
		if ok {
			contentBytes, isBytes := unNode.Content.([]byte)
			if isBytes {
				uname = string(contentBytes)
			} else if s, isStr := unNode.Content.(string); isStr {
				uname = s
			}
		}

		results = append(results, UsernameInfo{
			JID:      targetJID,
			Username: uname,
		})
	}

	return results, nil
}
