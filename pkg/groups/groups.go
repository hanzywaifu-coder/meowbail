package groups

import (
	"context"
	"fmt"
	protocol "github.com/hanzywaifu-coder/dongtube-meowbail/internal/protocol"
	"strings"
	"sync"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

type cachedGroupMeta struct {
	info *protocol.GroupInfo
	exp  time.Time
}

var (
	groupInfoCacheMu sync.RWMutex
	groupInfoCache   = make(map[protocol.JID]*cachedGroupMeta)
)

// InvalidateGroupInfoCache membersihkan cache metadata grup saat ada perubahan peserta
func InvalidateGroupInfoCache(groupJID protocol.JID) {
	groupInfoCacheMu.Lock()
	delete(groupInfoCache, groupJID.ToNonAD())
	groupInfoCacheMu.Unlock()
}

// GetGroupInfo gets group information with intelligent TTL caching (60s) to prevent IQ flood & timeout
func GetGroupInfo(c *protocol.Client, ctx context.Context, groupJID protocol.JID) (*protocol.GroupInfo, error) {
	normJID := groupJID.ToNonAD()

	groupInfoCacheMu.RLock()
	cached, ok := groupInfoCache[normJID]
	if ok && time.Now().Before(cached.exp) {
		infoCopy := cached.info
		groupInfoCacheMu.RUnlock()
		return infoCopy, nil
	}
	groupInfoCacheMu.RUnlock()

	ctxTimeout, cancel := context.WithTimeout(ctx, 12*time.Second)
	defer cancel()

	info, err := c.Client.GetGroupInfo(ctxTimeout, groupJID)
	if err != nil {
		// Jika timeout/gagal tapi kita punya cached versi lama, gunakan sebagai graceful degradation
		groupInfoCacheMu.RLock()
		if stale, okStale := groupInfoCache[normJID]; okStale && stale.info != nil {
			groupInfoCacheMu.RUnlock()
			return stale.info, nil
		}
		groupInfoCacheMu.RUnlock()
		return nil, err
	}

	groupInfoCacheMu.Lock()
	groupInfoCache[normJID] = &cachedGroupMeta{
		info: info,
		exp:  time.Now().Add(60 * time.Second),
	}
	groupInfoCacheMu.Unlock()

	return info, nil
}

// GetJoinedGroups mengambil seluruh daftar grup tempat bot bergabung beserta metadata lengkapnya
func GetJoinedGroups(c *protocol.Client, ctx context.Context) ([]*protocol.GroupInfo, error) {
	return c.Client.GetJoinedGroups(ctx)
}

// GetGroupInviteLink gets group invite link
func GetGroupInviteLink(c *protocol.Client, ctx context.Context, groupJID protocol.JID, reset bool) (string, error) {
	return c.Client.GetGroupInviteLink(ctx, groupJID, reset)
}

// GetGroupInfoFromInviteLink mengambil metadata grup dari link undangan chat.whatsapp.com/xxx
func GetGroupInfoFromInviteLink(c *protocol.Client, ctx context.Context, linkOrCode string) (*protocol.GroupInfo, error) {
	code := linkOrCode
	if idx := strings.Index(code, "chat.whatsapp.com/"); idx != -1 {
		code = code[idx+len("chat.whatsapp.com/"):]
		if slashIdx := strings.Index(code, "/"); slashIdx != -1 {
			code = code[:slashIdx]
		}
		if qIdx := strings.Index(code, "?"); qIdx != -1 {
			code = code[:qIdx]
		}
	}
	code = strings.TrimSpace(code)
	return c.Client.GetGroupInfoFromLink(ctx, code)
}

// JoinGroupWithInviteLink bergabung ke grup menggunakan link undangan
func JoinGroupWithInviteLink(c *protocol.Client, ctx context.Context, linkOrCode string) (protocol.JID, error) {
	code := linkOrCode
	if idx := strings.Index(code, "chat.whatsapp.com/"); idx != -1 {
		code = code[idx+len("chat.whatsapp.com/"):]
		if slashIdx := strings.Index(code, "/"); slashIdx != -1 {
			code = code[:slashIdx]
		}
		if qIdx := strings.Index(code, "?"); qIdx != -1 {
			code = code[:qIdx]
		}
	}
	code = strings.TrimSpace(code)
	return c.Client.JoinGroupWithLink(ctx, code)
}

// SetGroupName sets group name
func SetGroupName(c *protocol.Client, ctx context.Context, groupJID protocol.JID, name string) error {
	return c.Client.SetGroupName(ctx, groupJID, name)
}

// SetGroupDescription sets group description
func SetGroupDescription(c *protocol.Client, ctx context.Context, groupJID protocol.JID, description string) error {
	return c.Client.SetGroupDescription(ctx, groupJID, description)
}

// SetGroupAnnounce sets group announce mode
func SetGroupAnnounce(c *protocol.Client, ctx context.Context, groupJID protocol.JID, announce bool) error {
	return c.Client.SetGroupAnnounce(ctx, groupJID, announce)
}

// SetGroupLocked sets group locked status
func SetGroupLocked(c *protocol.Client, ctx context.Context, groupJID protocol.JID, locked bool) error {
	return c.Client.SetGroupLocked(ctx, groupJID, locked)
}

// SetGroupEphemeral sets group disappearing messages timer
func SetGroupEphemeral(c *protocol.Client, ctx context.Context, groupJID protocol.JID, timer time.Duration) error {
	return c.Client.SetDisappearingTimer(ctx, groupJID, timer, time.Now())
}

// PromoteParticipant promotes a participant to admin
func PromoteParticipant(c *protocol.Client, ctx context.Context, groupJID protocol.JID, participants ...types.JID) error {
	var targets []protocol.JID
	for _, p := range participants {
		targets = append(targets, p.ToNonAD())
	}
	_, err := c.Client.UpdateGroupParticipants(ctx, groupJID, targets, whatsmeow.ParticipantChangePromote)
	if err == nil {
		InvalidateGroupInfoCache(groupJID)
	}
	return err
}

// DemoteParticipant demotes an admin to regular member
func DemoteParticipant(c *protocol.Client, ctx context.Context, groupJID protocol.JID, participants ...types.JID) error {
	var targets []protocol.JID
	for _, p := range participants {
		targets = append(targets, p.ToNonAD())
	}
	_, err := c.Client.UpdateGroupParticipants(ctx, groupJID, targets, whatsmeow.ParticipantChangeDemote)
	if err == nil {
		InvalidateGroupInfoCache(groupJID)
	}
	return err
}

// RemoveParticipant removes a participant from group
func RemoveParticipant(c *protocol.Client, ctx context.Context, groupJID protocol.JID, participants ...types.JID) error {
	var targets []protocol.JID
	for _, p := range participants {
		targets = append(targets, p.ToNonAD())
	}
	_, err := c.Client.UpdateGroupParticipants(ctx, groupJID, targets, whatsmeow.ParticipantChangeRemove)
	if err == nil {
		InvalidateGroupInfoCache(groupJID)
	}
	return err
}

// GetGroupRequestParticipants mengambil daftar permintaan bergabung (membership approval requests)
func GetGroupRequestParticipants(c *protocol.Client, ctx context.Context, groupJID protocol.JID) ([]protocol.GroupParticipantRequest, error) {
	return c.Client.GetGroupRequestParticipants(ctx, groupJID)
}

// ApproveGroupRequests menyetujui calon anggota yang meminta masuk ke grup
func ApproveGroupRequests(c *protocol.Client, ctx context.Context, groupJID protocol.JID, participants ...types.JID) ([]protocol.GroupParticipant, error) {
	res, err := c.Client.UpdateGroupRequestParticipants(ctx, groupJID, participants, whatsmeow.ParticipantChangeApprove)
	if err == nil {
		InvalidateGroupInfoCache(groupJID)
	}
	return res, err
}

// RejectGroupRequests menolak calon anggota yang meminta masuk ke grup
func RejectGroupRequests(c *protocol.Client, ctx context.Context, groupJID protocol.JID, participants ...types.JID) ([]protocol.GroupParticipant, error) {
	return c.Client.UpdateGroupRequestParticipants(ctx, groupJID, participants, whatsmeow.ParticipantChangeReject)
}

// AddParticipant adds a participant to group
func AddParticipant(c *protocol.Client, ctx context.Context, groupJID protocol.JID, participants ...types.JID) error {
	var targets []protocol.JID
	for _, p := range participants {
		targets = append(targets, p.ToNonAD())
	}
	_, err := c.Client.UpdateGroupParticipants(ctx, groupJID, targets, whatsmeow.ParticipantChangeAdd)
	if err == nil {
		InvalidateGroupInfoCache(groupJID)
	}
	return err
}

// LeaveGroup makes the bot leave the group
func LeaveGroup(c *protocol.Client, ctx context.Context, groupJID protocol.JID) error {
	InvalidateGroupInfoCache(groupJID)
	return c.Client.LeaveGroup(ctx, groupJID)
}

// ResolveParticipantPN menyelesaikan JID user (termasuk LID) ke nomor telepon asli dan info kontak
func ResolveParticipantPN(c *protocol.Client, ctx context.Context, groupJID protocol.JID, userJID protocol.JID) (protocol.JID, string) {
	phoneJID := userJID
	pushName := ""
	if c == nil {
		return phoneJID, pushName
	}

	// 1. Cek melalui protocol.LIDResolver in-memory jika userJID adalah LID
	if c.LIDResolver != nil && (userJID.Server == protocol.HiddenUserServer || userJID.Server == "lid") {
		resolved := c.LIDResolver.ResolveToPN(userJID)
		if resolved.Server == protocol.DefaultUserServer {
			phoneJID = resolved
		}
	}

	// 2. Jika masih bertipe LID, lakukan pencocokan di daftar peserta grup dari GetGroupInfo
	if phoneJID.Server != protocol.DefaultUserServer && groupJID.Server == protocol.GroupServer {
		ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
		groupInfo, err := c.Client.GetGroupInfo(ctxTimeout, groupJID)
		cancel()
		if err == nil && groupInfo != nil {
			normLID := userJID.ToNonAD().User
			for _, p := range groupInfo.Participants {
				if (p.LID.ToNonAD().User == normLID || p.JID.ToNonAD().User == normLID) && p.PhoneNumber.Server == protocol.DefaultUserServer {
					phoneJID = p.PhoneNumber
					if c.LIDResolver != nil {
						c.LIDResolver.RegisterMapping(p.LID.String(), p.PhoneNumber.String())
					}
					break
				} else if p.LID.ToNonAD().User == normLID && p.JID.Server == protocol.DefaultUserServer {
					phoneJID = p.JID
					if c.LIDResolver != nil {
						c.LIDResolver.RegisterMapping(p.LID.String(), p.JID.String())
					}
					break
				}
			}
		}
	}

	// 3. Ambil pushName / nama tersimpan dari local store contacts
	if c.Client != nil && c.Client.Store != nil && c.Client.Store.Contacts != nil {
		if contact, err := c.Client.Store.Contacts.GetContact(ctx, phoneJID); err == nil {
			if contact.PushName != "" {
				pushName = contact.PushName
			} else if contact.FullName != "" {
				pushName = contact.FullName
			} else if contact.BusinessName != "" {
				pushName = contact.BusinessName
			}
		}
	}

	return phoneJID, pushName
}

// IsBotAdmin checks if the bot is admin in the group
func IsBotAdmin(c *protocol.Client, ctx context.Context, groupJID protocol.JID) (bool, error) {
	groupInfo, err := GetGroupInfo(c, ctx, groupJID)
	if err != nil {
		return false, err
	}

	if c.Client == nil || c.Client.Store == nil {
		return false, fmt.Errorf("client store not initialized")
	}

	// Check both ID (phone) and LID (linked device) safely
	var botJID, botLID, botPhone string
	if c.Client.Store.ID != nil {
		botJID = c.Client.Store.ID.ToNonAD().User
		botPhone = botJID
	}
	if c.Client.Store.LID.User != "" {
		botLID = c.Client.Store.LID.ToNonAD().User
	}

	for _, p := range groupInfo.Participants {
		pUser := p.JID.ToNonAD().User
		pLID := p.LID.ToNonAD().User
		pPhone := p.PhoneNumber.ToNonAD().User

		isSelf := false
		if botJID != "" && (pUser == botJID || pLID == botJID || pPhone == botJID) {
			isSelf = true
		}
		if botLID != "" && (pUser == botLID || pLID == botLID || pPhone == botLID) {
			isSelf = true
		}
		if botPhone != "" && (pUser == botPhone || pLID == botPhone || pPhone == botPhone) {
			isSelf = true
		}

		if isSelf {
			if p.IsAdmin || p.IsSuperAdmin {
				return true, nil
			}
		}
	}

	return false, nil
}

// IsUserAdmin checks if a user is admin in the group (supports LID and phone JID matching)
func IsUserAdmin(c *protocol.Client, ctx context.Context, groupJID protocol.JID, userJID protocol.JID) (bool, error) {
	groupInfo, err := GetGroupInfo(c, ctx, groupJID)
	if err != nil {
		return false, err
	}

	normUser := userJID.ToNonAD().User

	// Cek juga mapping nomor asli jika userJID berupa LID atau sebaliknya
	resolvedUser := ""
	if c.LIDResolver != nil {
		pn := c.LIDResolver.ResolveToPN(userJID.ToNonAD())
		if pn.Server == protocol.DefaultUserServer && pn.User != "" {
			resolvedUser = pn.User
		}
	}

	for _, p := range groupInfo.Participants {
		pJID := p.JID.ToNonAD().User
		pLID := p.LID.ToNonAD().User
		pPN := p.PhoneNumber.ToNonAD().User

		matched := (pJID != "" && (pJID == normUser || pJID == resolvedUser)) ||
			(pLID != "" && (pLID == normUser || pLID == resolvedUser)) ||
			(pPN != "" && (pPN == normUser || pPN == resolvedUser))

		if matched && (p.IsAdmin || p.IsSuperAdmin) {
			return true, nil
		}
	}

	return false, nil
}

// GetGroupParticipants returns all participants in the group
func GetGroupParticipants(c *protocol.Client, ctx context.Context, groupJID protocol.JID) ([]protocol.GroupParticipant, error) {
	groupInfo, err := GetGroupInfo(c, ctx, groupJID)
	if err != nil {
		return nil, err
	}

	return groupInfo.Participants, nil
}

// TagAll mentions all participants in the group
func TagAll(c *protocol.Client, ctx context.Context, groupJID protocol.JID, message string) error {
	participants, err := GetGroupParticipants(c, ctx, groupJID)
	if err != nil {
		return err
	}

	var mentionedJids []string
	for _, p := range participants {
		mentionedJids = append(mentionedJids, p.JID.String())
	}

	text := message + "\n"
	for _, p := range participants {
		text += fmt.Sprintf("@%s\n", p.JID.User)
	}

	_, sendErr := c.Client.SendMessage(ctx, groupJID, &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: proto.String(text),
			ContextInfo: &waE2E.ContextInfo{
				MentionedJID: mentionedJids,
			},
		},
	})
	return sendErr
}
