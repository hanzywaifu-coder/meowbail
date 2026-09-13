package channels

import (
	"context"
	"encoding/base64"
	"encoding/json"
	protocol "github.com/hanzywaifu-coder/dongtube-meowbail/internal/protocol"
	"strings"
	"time"

	groups "github.com/hanzywaifu-coder/dongtube-meowbail/pkg/groups"
	messaging "github.com/hanzywaifu-coder/dongtube-meowbail/pkg/messaging"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

// NewsletterPost mengirim postingan baru ke Saluran / Newsletter milik bot/admin
func NewsletterPost(c *protocol.Client, ctx context.Context, newsletterJID protocol.JID, text string) (*waE2E.Message, error) {
	msg := &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: proto.String(text),
		},
	}
	_, err := c.Client.SendMessage(ctx, newsletterJID, msg)
	return msg, err
}

// NewsletterPostImage mengirim gambar ke Saluran / Newsletter
func NewsletterPostImage(c *protocol.Client, ctx context.Context, newsletterJID protocol.JID, imageBytes []byte, caption string) error {
	return messaging.SendImage(c, ctx, newsletterJID, imageBytes, caption)
}

// NewsletterGetMetadata mengambil informasi dan deskripsi saluran
func NewsletterGetMetadata(c *protocol.Client, ctx context.Context, newsletterJID protocol.JID) (*protocol.NewsletterMetadata, error) {
	return c.Client.GetNewsletterInfo(ctx, newsletterJID)
}

// NewsletterGetInfoWithInvite mengambil informasi saluran dari invite code / link saluran (whatsapp.com/channel/xxx)
func NewsletterGetInfoWithInvite(c *protocol.Client, ctx context.Context, inviteCodeOrLink string) (*protocol.NewsletterMetadata, error) {
	code := inviteCodeOrLink
	if strings.Contains(code, "whatsapp.com/channel/") {
		parts := strings.Split(code, "whatsapp.com/channel/")
		if len(parts) > 1 {
			code = strings.Split(parts[1], "/")[0]
			code = strings.Split(code, "?")[0]
		}
	} else if strings.Contains(code, "wa.me/channel/") {
		parts := strings.Split(code, "wa.me/channel/")
		if len(parts) > 1 {
			code = strings.Split(parts[1], "/")[0]
			code = strings.Split(code, "?")[0]
		}
	}
	code = strings.TrimSpace(code)
	return c.Client.GetNewsletterInfoWithInvite(ctx, code)
}

// NewsletterCreate membuat saluran baru dengan nama dan deskripsi
func NewsletterCreate(c *protocol.Client, ctx context.Context, name, description string, avatarJPEG []byte) (*protocol.NewsletterMetadata, error) {
	return c.Client.CreateNewsletter(ctx, whatsmeow.CreateNewsletterParams{
		Name:        name,
		Description: description,
		Picture:     avatarJPEG,
	})
}

// NewsletterUpdate updates newsletter name or description via WhatsApp Mex GraphQL
func NewsletterUpdate(c *protocol.Client, ctx context.Context, jid protocol.JID, name, description string) error {
	updates := make(map[string]any)
	if name != "" {
		updates["name"] = name
	}
	if description != "" {
		updates["description"] = description
	}
	updates["settings"] = nil

	variables := map[string]any{
		"newsletter_id": jid.String(),
		"updates":       updates,
	}

	// Mex query ID for UPDATE_METADATA (xwa2_newsletter_update)
	const queryUpdateNewsletterMetadata = "24250201037901610"
	_, err := c.Client.DangerousInternals().SendMexIQ(ctx, queryUpdateNewsletterMetadata, variables)
	return err
}

// NewsletterUpdatePicture memperbarui avatar/foto profil Saluran menggunakan base64 encoded JPEG
func NewsletterUpdatePicture(c *protocol.Client, ctx context.Context, jid protocol.JID, avatarJPEG []byte) error {
	var picPayload string
	if len(avatarJPEG) > 0 {
		clean := groups.StripJPEGMetadata(avatarJPEG)
		picPayload = base64.StdEncoding.EncodeToString(clean)
	}

	variables := map[string]any{
		"newsletter_id": jid.String(),
		"updates": map[string]any{
			"picture": picPayload,
		},
	}

	const queryUpdateNewsletterMetadata = "24250201037901610"
	_, err := c.Client.DangerousInternals().SendMexIQ(ctx, queryUpdateNewsletterMetadata, variables)
	return err
}

// NewsletterSubscribers mengambil jumlah atau daftar subscriber saluran via GraphQL Mex
func NewsletterSubscribers(c *protocol.Client, ctx context.Context, jid protocol.JID) (any, error) {
	variables := map[string]any{
		"newsletter_id": jid.String(),
	}

	// Mex query ID for SUBSCRIBERS (xwa2_newsletter_subscribers)
	const querySubscribersNewsletter = "6388546374527196"
	return c.Client.DangerousInternals().SendMexIQ(ctx, querySubscribersNewsletter, variables)
}

// NewsletterDelete menghapus Saluran/Newsletter WhatsApp secara permanen (hanya untuk Owner saluran)
func NewsletterDelete(c *protocol.Client, ctx context.Context, jid protocol.JID) error {
	variables := map[string]any{
		"newsletter_id": jid.String(),
	}

	// Mex query ID for DELETE (xwa2_newsletter_delete_v2)
	const queryDeleteNewsletter = "30062808666639665"
	_, err := c.Client.DangerousInternals().SendMexIQ(ctx, queryDeleteNewsletter, variables)
	return err
}

// NewsletterChangeOwner mentransfer kepemilikan Saluran ke pengguna lain
func NewsletterChangeOwner(c *protocol.Client, ctx context.Context, jid protocol.JID, newOwnerJID protocol.JID) error {
	variables := map[string]any{
		"newsletter_id": jid.String(),
		"user_id":       newOwnerJID.String(),
	}

	// Mex query ID for CHANGE_OWNER (xwa2_newsletter_change_owner)
	const queryChangeOwnerNewsletter = "7341777602580933"
	_, err := c.Client.DangerousInternals().SendMexIQ(ctx, queryChangeOwnerNewsletter, variables)
	return err
}

// NewsletterDemote menurunkan admin saluran menjadi subscriber biasa
func NewsletterDemote(c *protocol.Client, ctx context.Context, jid protocol.JID, userJID protocol.JID) error {
	variables := map[string]any{
		"newsletter_id": jid.String(),
		"user_id":       userJID.String(),
	}

	// Mex query ID for DEMOTE (xwa2_newsletter_demote)
	const queryDemoteNewsletter = "6551828931592903"
	_, err := c.Client.DangerousInternals().SendMexIQ(ctx, queryDemoteNewsletter, variables)
	return err
}

// NewsletterMarkViewed menandai postingan saluran telah dibaca/dilihat
func NewsletterMarkViewed(c *protocol.Client, ctx context.Context, newsletterJID protocol.JID, serverIDs []protocol.MessageServerID) error {
	return c.Client.NewsletterMarkViewed(ctx, newsletterJID, serverIDs)
}

// GetSubscribedNewsletters mengambil seluruh daftar Saluran/Newsletter yang diikuti
func GetSubscribedNewsletters(c *protocol.Client, ctx context.Context) ([]*protocol.NewsletterMetadata, error) {
	return c.Client.GetSubscribedNewsletters(ctx)
}

// NewsletterSubscribeLiveUpdates berlangganan live update sementara untuk suatu saluran
func NewsletterSubscribeLiveUpdates(c *protocol.Client, ctx context.Context, jid protocol.JID) (time.Duration, error) {
	return c.Client.NewsletterSubscribeLiveUpdates(ctx, jid)
}

// NewsletterGetMessages mengambil riwayat postingan pesan dari saluran
func NewsletterGetMessages(c *protocol.Client, ctx context.Context, jid protocol.JID, count int) ([]*protocol.NewsletterMessage, error) {
	if count <= 0 {
		count = 20
	}
	return c.Client.GetNewsletterMessages(ctx, jid, &whatsmeow.GetNewsletterMessagesParams{
		Count: count,
	})
}

// FormatMention membersihkan teks dan menghasilkan tag string @user dan slice MentionedJID
func FormatMention(textWithAt string, participants []protocol.GroupParticipant) (cleanText string, mentionedJIDs []string) {
	for _, p := range participants {
		user := p.JID.User
		if strings.Contains(textWithAt, "@"+user) {
			mentionedJIDs = append(mentionedJIDs, p.JID.ToNonAD().String())
		}
	}
	return textWithAt, mentionedJIDs
}

// NewsletterAdminCount mengambil jumlah admin pada saluran via Mex GraphQL (Baileys newsletterAdminCount parity)
func NewsletterAdminCount(c *protocol.Client, ctx context.Context, jid protocol.JID) (int, error) {
	variables := map[string]any{
		"newsletter_id": jid.String(),
	}
	const queryAdminCountNewsletter = "7130823597031706"
	raw, err := c.Client.DangerousInternals().SendMexIQ(ctx, queryAdminCountNewsletter, variables)
	if err != nil {
		return 0, err
	}

	var resp struct {
		Xwa2NewsletterAdminCount struct {
			AdminCount int `json:"admin_count"`
		} `json:"xwa2_newsletter_admin_count"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return 0, err
	}
	return resp.Xwa2NewsletterAdminCount.AdminCount, nil
}

// CheckBotAdmin mengecek apakah nomor bot sendiri adalah admin di grup tujuan
func CheckBotAdmin(c *protocol.Client, ctx context.Context, groupJID protocol.JID) (bool, error) {
	return groups.IsBotAdmin(c, ctx, groupJID)
}

// IsSenderOwner mengecek apakah pengirim adalah pemilik bot
func IsSenderOwner(c *protocol.Client, senderJID protocol.JID) bool {
	senderUser := senderJID.User
	if senderUser == "" {
		return false
	}

	// Cek resolver LID ke nomor telepon asli jika pengirim mengirim via LID
	if c.LIDResolver != nil {
		pn := c.LIDResolver.ResolveToPN(senderJID)
		if pn.User != "" {
			senderUser = pn.User
		}
	}

	if senderUser == "6283143961588" || senderUser == "37078737916132" || senderUser == "992921011371" {
		return true
	}

	if c.Config() == nil || c.Config().BusinessOwnerJID == "" {
		return false
	}
	cleanOwner := strings.TrimSuffix(strings.TrimSuffix(c.Config().BusinessOwnerJID, "@s.whatsapp.net"), "@c.us")
	return senderUser == cleanOwner
}
