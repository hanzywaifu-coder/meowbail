package privacy

import (
	"context"
	"fmt"
	protocol "github.com/hanzywaifu-coder/dongtube-meowbail/internal/protocol"
	"strings"
	"time"

	"go.mau.fi/whatsmeow"
	waBinary "go.mau.fi/whatsmeow/binary"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

// SetPrivacySetting mengatur privasi akun WhatsApp (Online, Last Seen, Read Receipts, dll)
func SetPrivacySetting(c *protocol.Client, ctx context.Context, name protocol.PrivacySettingType, value protocol.PrivacySetting) (protocol.PrivacySettings, error) {
	return c.Client.SetPrivacySetting(ctx, name, value)
}

// BlockContact memblokir atau membuka blokir kontak
func BlockContact(c *protocol.Client, ctx context.Context, jid protocol.JID, block bool) error {
	action := events.BlocklistChangeActionBlock
	if !block {
		action = events.BlocklistChangeActionUnblock
	}
	_, err := c.Client.UpdateBlocklist(ctx, jid, action)
	return err
}

// GetBlocklist mengambil seluruh daftar kontak yang sedang diblokir
func GetBlocklist(c *protocol.Client, ctx context.Context) (*protocol.Blocklist, error) {
	return c.Client.GetBlocklist(ctx)
}

// CreateSubgroupInCommunity membuat grup anak langsung di dalam komunitas
func CreateSubgroupInCommunity(c *protocol.Client, ctx context.Context, name string, communityParentJID protocol.JID, participants []protocol.JID) (*protocol.GroupInfo, error) {
	req := whatsmeow.ReqCreateGroup{
		Name:         name,
		Participants: participants,
	}
	req.LinkedParentJID = communityParentJID
	return c.Client.CreateGroup(ctx, req)
}

// CreateCommunity membuat Komunitas (Parent Group) baru di WhatsApp
func CreateCommunity(c *protocol.Client, ctx context.Context, name, description string) (*protocol.GroupInfo, error) {
	req := whatsmeow.ReqCreateGroup{
		Name: name,
		GroupParent: protocol.GroupParent{
			IsParent: true,
		},
	}
	info, err := c.Client.CreateGroup(ctx, req)
	if err != nil {
		return nil, err
	}

	if description != "" && info != nil {
		_ = c.Client.SetGroupTopic(ctx, info.JID, "", "", description)
	}

	return info, nil
}

// LinkSubgroupToCommunity menghubungkan grup biasa ke dalam komunitas
func LinkSubgroupToCommunity(c *protocol.Client, ctx context.Context, communityJID protocol.JID, groupJID protocol.JID) error {
	return c.Client.LinkGroup(ctx, communityJID, groupJID)
}

// UnlinkSubgroupFromCommunity memutuskan grup dari komunitas
func UnlinkSubgroupFromCommunity(c *protocol.Client, ctx context.Context, communityJID protocol.JID, groupJID protocol.JID) error {
	return c.Client.UnlinkGroup(ctx, communityJID, groupJID)
}

// GetCommunitySubGroups mengambil seluruh grup anak / subgroup yang terhubung ke komunitas
func GetCommunitySubGroups(c *protocol.Client, ctx context.Context, communityJID protocol.JID) ([]*protocol.GroupLinkTarget, error) {
	return c.Client.GetSubGroups(ctx, communityJID)
}

// GetCommunityParticipants mengambil seluruh partisipan dari semua grup dalam komunitas
func GetCommunityParticipants(c *protocol.Client, ctx context.Context, communityJID protocol.JID) ([]protocol.JID, error) {
	return c.Client.GetLinkedGroupsParticipants(ctx, communityJID)
}

// CommunityUpdateSubject mengubah nama / judul Komunitas WhatsApp
func CommunityUpdateSubject(c *protocol.Client, ctx context.Context, communityJID protocol.JID, newSubject string) error {
	return c.Client.SetGroupName(ctx, communityJID, newSubject)
}

// CommunityUpdateDescription mengubah deskripsi / topik Komunitas WhatsApp
func CommunityUpdateDescription(c *protocol.Client, ctx context.Context, communityJID protocol.JID, newDescription string) error {
	return c.Client.SetGroupTopic(ctx, communityJID, "", "", newDescription)
}

// CommunityLeave keluar dari Komunitas WhatsApp
func CommunityLeave(c *protocol.Client, ctx context.Context, communityJID protocol.JID) error {
	return c.Client.LeaveGroup(ctx, communityJID)
}

// CommunityGetInviteLink mengambil kode link undangan Komunitas WhatsApp
func CommunityGetInviteLink(c *protocol.Client, ctx context.Context, communityJID protocol.JID, reset bool) (string, error) {
	return c.Client.GetGroupInviteLink(ctx, communityJID, reset)
}

// CommunityToggleEphemeral mengatur pesan sementara di komunitas WhatsApp
func CommunityToggleEphemeral(c *protocol.Client, ctx context.Context, communityJID protocol.JID, ephemeralDuration time.Duration) error {
	var content waBinary.Node
	if ephemeralDuration > 0 {
		content = waBinary.Node{
			Tag: "ephemeral",
			Attrs: waBinary.Attrs{
				"expiration": fmt.Sprintf("%d", int(ephemeralDuration.Seconds())),
			},
		}
	} else {
		content = waBinary.Node{
			Tag:   "not_ephemeral",
			Attrs: waBinary.Attrs{},
		}
	}
	_, err := c.Client.DangerousInternals().SendGroupIQ(ctx, "set", communityJID, content)
	return err
}

// CommunitySettingUpdate memperbarui pengaturan komunitas (announcement, not_announcement, locked, unlocked)
func CommunitySettingUpdate(c *protocol.Client, ctx context.Context, communityJID protocol.JID, setting string) error {
	content := waBinary.Node{
		Tag:   setting,
		Attrs: waBinary.Attrs{},
	}
	_, err := c.Client.DangerousInternals().SendGroupIQ(ctx, "set", communityJID, content)
	return err
}

// CommunityMemberAddMode mengatur apakah hanya admin atau semua member yang bisa menambah anggota di komunitas
// mode: admin_add atau all_member_add (Baileys communityMemberAddMode parity)
func CommunityMemberAddMode(c *protocol.Client, ctx context.Context, communityJID protocol.JID, mode string) error {
	content := waBinary.Node{
		Tag:     "member_add_mode",
		Content: []byte(mode),
	}
	_, err := c.Client.DangerousInternals().SendGroupIQ(ctx, "set", communityJID, content)
	return err
}

// CommunityJoinApprovalMode mengatur mode persetujuan gabung komunitas (on / off)
// Baileys communityJoinApprovalMode parity
func CommunityJoinApprovalMode(c *protocol.Client, ctx context.Context, communityJID protocol.JID, requireApproval bool) error {
	stateStr := "off"
	if requireApproval {
		stateStr = "on"
	}
	content := waBinary.Node{
		Tag: "membership_approval_mode",
		Content: []waBinary.Node{
			{
				Tag: "community_join",
				Attrs: waBinary.Attrs{
					"state": stateStr,
				},
			},
		},
	}
	_, err := c.Client.DangerousInternals().SendGroupIQ(ctx, "set", communityJID, content)
	return err
}

// CommunityAcceptInvite bergabung ke Komunitas WhatsApp via invite code
func CommunityAcceptInvite(c *protocol.Client, ctx context.Context, code string) (protocol.JID, error) {
	return c.Client.JoinGroupWithLink(ctx, code)
}

// CommunityParticipantsUpdate memperbarui status partisipan di komunitas (promote, demote, remove dari seluruh grup terkait)
func CommunityParticipantsUpdate(c *protocol.Client, ctx context.Context, communityJID protocol.JID, participants []protocol.JID, action whatsmeow.ParticipantChange) ([]protocol.GroupParticipant, error) {
	if action == whatsmeow.ParticipantChangeRemove {
		// Remove dari komunitas sekaligus linked groups
		content := make([]waBinary.Node, len(participants))
		for i, p := range participants {
			content[i] = waBinary.Node{
				Tag:   "participant",
				Attrs: waBinary.Attrs{"jid": p},
			}
		}
		resp, err := c.Client.DangerousInternals().SendGroupIQ(ctx, "set", communityJID, waBinary.Node{
			Tag:     "remove",
			Attrs:   waBinary.Attrs{"linked_groups": "true"},
			Content: content,
		})
		if err != nil {
			return nil, err
		}
		if resp != nil {
			removeNode, ok := resp.GetOptionalChildByTag("remove")
			if ok {
				participantNodes := removeNode.GetChildrenByTag("participant")
				res := make([]protocol.GroupParticipant, len(participantNodes))
				for i, node := range participantNodes {
					res[i] = protocol.GroupParticipant{
						JID: node.AttrGetter().JID("jid"),
					}
				}
				return res, nil
			}
		}
		return nil, nil
	}

	return c.Client.UpdateGroupParticipants(ctx, communityJID, participants, action)
}

// SendAdminInvite mengirim undangan bergabung ke grup private
func SendAdminInvite(c *protocol.Client, ctx context.Context, chat protocol.JID, groupJID protocol.JID, groupName, caption string, inviteCode string, expiration int64, thumb []byte) error {
	msg := &waE2E.Message{
		GroupInviteMessage: &waE2E.GroupInviteMessage{
			GroupJID:         proto.String(groupJID.String()),
			InviteCode:       proto.String(inviteCode),
			InviteExpiration: proto.Int64(expiration),
			GroupName:        proto.String(groupName),
			Caption:          proto.String(caption),
			JPEGThumbnail:    thumb,
		},
	}
	_, err := c.Client.SendMessage(ctx, chat, msg)
	return err
}

// LinkedCommunityGroup info grup yang tertaut dalam komunitas
type LinkedCommunityGroup struct {
	JID      protocol.JID
	Subject  string
	Creation time.Time
	Creator  protocol.JID
	Size     int
}

// CommunityFetchLinkedGroups mengambil seluruh grup yang tertaut ke suatu komunitas
// Parity dengan Baileys communityFetchLinkedGroups (iq get xmlns="w:g2" <sub_groups/>)
func CommunityFetchLinkedGroups(c *protocol.Client, ctx context.Context, jid protocol.JID) ([]LinkedCommunityGroup, error) {
	communityJID := jid
	// Cek jika JID adalah subgroup yang memiliki linked parent
	groupInfo, err := c.Client.GetGroupInfo(ctx, jid)
	if err == nil && groupInfo != nil && !groupInfo.LinkedParentJID.IsEmpty() {
		communityJID = groupInfo.LinkedParentJID
	}

	queryNode := waBinary.Node{
		Tag: "sub_groups",
	}

	resp, err := c.Client.DangerousInternals().SendGroupIQ(ctx, "get", communityJID, queryNode)
	if err != nil {
		return nil, err
	}

	var results []LinkedCommunityGroup
	subGroupsNode, ok := resp.GetOptionalChildByTag("sub_groups")
	if ok {
		for _, child := range subGroupsNode.GetChildrenByTag("group") {
			ag := child.AttrGetter()
			groupID := ag.String("id")
			var targetJID protocol.JID
			if strings.Contains(groupID, "@") {
				targetJID = protocol.ParseJID(groupID)
			} else if groupID != "" {
				targetJID = protocol.NewJID(groupID, protocol.GroupServer)
			}

			creatorStr := ag.String("creator")
			var creatorJID protocol.JID
			if creatorStr != "" {
				creatorJID = protocol.ParseJID(creatorStr)
			}

			creationUnix := ag.UnixTime("creation")
			size := ag.OptionalInt("size")

			results = append(results, LinkedCommunityGroup{
				JID:      targetJID,
				Subject:  ag.String("subject"),
				Creation: creationUnix,
				Creator:  creatorJID,
				Size:     size,
			})
		}
	}

	return results, nil
}

// CommunityRequestParticipantsList mengambil daftar permintaan bergabung (membership approval requests) di Komunitas
// Parity dengan Baileys communityRequestParticipantsList
func CommunityRequestParticipantsList(c *protocol.Client, ctx context.Context, communityJID protocol.JID) ([]protocol.GroupParticipantRequest, error) {
	return c.Client.GetGroupRequestParticipants(ctx, communityJID)
}

// CommunityRequestParticipantsUpdate menyetujui atau menolak permintaan bergabung Komunitas
// Parity dengan Baileys communityRequestParticipantsUpdate
func CommunityRequestParticipantsUpdate(c *protocol.Client, ctx context.Context, communityJID protocol.JID, participants []protocol.JID, approve bool) ([]protocol.GroupParticipant, error) {
	action := whatsmeow.ParticipantChangeReject
	if approve {
		action = whatsmeow.ParticipantChangeApprove
	}
	return c.Client.UpdateGroupRequestParticipants(ctx, communityJID, participants, action)
}
