// Package meowbail is the public facade of the Dongtube-meowbail WhatsApp library.
//
// It re-exports the engine (internal/protocol) and the domain packages under a
// single import so users get one flat, stable API:
//
//	import meowbail "github.com/hanzywaifu-coder/dongtube-meowbail"
//
// The engine lives in internal/protocol (client, session, pairing, reliability);
// domain features live in pkg/messaging, pkg/media, pkg/groups, pkg/channels,
// pkg/business, pkg/buttons and pkg/privacy. Client exposes the most-used
// features as Baileys-style methods; every domain function is also available
// package-level for advanced use.
package meowbail

import (
	"context"
	"time"

	protocol "github.com/hanzywaifu-coder/dongtube-meowbail/internal/protocol"
	"github.com/hanzywaifu-coder/dongtube-meowbail/pkg/business"
	"github.com/hanzywaifu-coder/dongtube-meowbail/pkg/buttons"
	"github.com/hanzywaifu-coder/dongtube-meowbail/pkg/channels"
	"github.com/hanzywaifu-coder/dongtube-meowbail/pkg/groups"
	"github.com/hanzywaifu-coder/dongtube-meowbail/pkg/media"
	"github.com/hanzywaifu-coder/dongtube-meowbail/pkg/messaging"
	"github.com/hanzywaifu-coder/dongtube-meowbail/pkg/privacy"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
)

// Client is the high-level dongtube-meowbail client. It wraps the protocol
// engine and promotes its methods (Connect, Disconnect, AddEventHandler, ...)
// while adding domain helpers as direct methods.
type Client struct {
	*protocol.Client
}

// NewClient creates a client from a device store / existing whatsmeow client.
func NewClient(device interface{}, logger interface{}, config ...*Config) *Client {
	return &Client{Client: protocol.NewClient(device, logger, config...)}
}

// NewClientFromWhatsmeow wraps an existing whatsmeow client.
func NewClientFromWhatsmeow(cli *whatsmeow.Client, config ...*Config) *Client {
	return &Client{Client: protocol.NewClientFromWhatsmeow(cli, config...)}
}

// ---------------------------------------------------------------------------
// Messaging methods
// ---------------------------------------------------------------------------

func (s *Client) SendText(ctx context.Context, chat JID, text string, opts ...*MessageOptions) error {
	return messaging.SendText(s.Client, ctx, chat, text, opts...)
}

func (s *Client) SendTextMessage(ctx context.Context, chat JID, text string, opts ...*MessageOptions) (whatsmeow.SendResponse, error) {
	return messaging.SendTextMessage(s.Client, ctx, chat, text, opts...)
}

func (s *Client) SendImage(ctx context.Context, chat JID, data []byte, caption string, opts ...*MessageOptions) error {
	return messaging.SendImage(s.Client, ctx, chat, data, caption, opts...)
}

func (s *Client) SendVideo(ctx context.Context, chat JID, data []byte, caption string, opts ...*MessageOptions) error {
	return messaging.SendVideo(s.Client, ctx, chat, data, caption, opts...)
}

func (s *Client) SendAudio(ctx context.Context, chat JID, data []byte, opts ...*MessageOptions) error {
	return messaging.SendAudio(s.Client, ctx, chat, data, opts...)
}

func (s *Client) SendDocument(ctx context.Context, chat JID, data []byte, filename, mimetype, caption string, opts ...*MessageOptions) error {
	return messaging.SendDocument(s.Client, ctx, chat, data, filename, mimetype, caption, opts...)
}

func (s *Client) SendButtons(ctx context.Context, chat JID, text string, btns []Button, opts ...*MessageOptions) error {
	return messaging.SendButtons(s.Client, ctx, chat, text, btns, opts...)
}

func (s *Client) SendReaction(ctx context.Context, chat JID, msgID MessageID, emoji string, sender ...JID) error {
	return messaging.SendReaction(s.Client, ctx, chat, msgID, emoji, sender...)
}

func (s *Client) SendMention(ctx context.Context, chat JID, text string, mentions []JID, quotedContext ...*waE2E.ContextInfo) error {
	return messaging.SendMention(s.Client, ctx, chat, text, mentions, quotedContext...)
}

func (s *Client) RevokeMessage(ctx context.Context, chat JID, targetMsgID MessageID, fromMe bool, sender JID) error {
	return messaging.RevokeMessage(s.Client, ctx, chat, targetMsgID, fromMe, sender)
}

// ---------------------------------------------------------------------------
// Media methods
// ---------------------------------------------------------------------------

func (s *Client) DownloadMedia(ctx context.Context, msg whatsmeow.DownloadableMessage) ([]byte, error) {
	return media.DownloadMedia(s.Client, ctx, msg)
}

func (s *Client) UploadMedia(ctx context.Context, data []byte, mediaType whatsmeow.MediaType) (*whatsmeow.UploadResponse, error) {
	return media.UploadMedia(s.Client, ctx, data, mediaType)
}

// ---------------------------------------------------------------------------
// Groups methods
// ---------------------------------------------------------------------------

func (s *Client) GetGroupInfo(ctx context.Context, groupJID JID) (*GroupInfo, error) {
	return groups.GetGroupInfo(s.Client, ctx, groupJID)
}

func (s *Client) GetGroupParticipants(ctx context.Context, groupJID JID) ([]GroupParticipant, error) {
	return groups.GetGroupParticipants(s.Client, ctx, groupJID)
}

func (s *Client) TagAll(ctx context.Context, groupJID JID, message string) error {
	return groups.TagAll(s.Client, ctx, groupJID, message)
}

func (s *Client) IsUserAdmin(ctx context.Context, groupJID, userJID JID) (bool, error) {
	return groups.IsUserAdmin(s.Client, ctx, groupJID, userJID)
}

func (s *Client) SendSticker(ctx context.Context, chat JID, data []byte, opts ...*MessageOptions) error {
	return messaging.SendSticker(s.Client, ctx, chat, data, opts...)
}

func (s *Client) ExtractUnderlyingMessage(msg *waE2E.Message) *protocol.UnwrappedMessage {
	return protocol.ExtractUnderlyingMessage(msg)
}

func (s *Client) AddStickerMetadata(webp []byte, pack, author string) ([]byte, error) {
	return media.AddStickerMetadata(webp, pack, author)
}

func (s *Client) IsBotAdmin(ctx context.Context, groupJID JID) (bool, error) {
	return groups.IsBotAdmin(s.Client, ctx, groupJID)
}

func (s *Client) GetGroupInviteLink(ctx context.Context, groupJID JID, reset bool) (string, error) {
	return groups.GetGroupInviteLink(s.Client, ctx, groupJID, reset)
}

func (s *Client) LeaveGroup(ctx context.Context, groupJID JID) error {
	return groups.LeaveGroup(s.Client, ctx, groupJID)
}

func (s *Client) SetGroupName(ctx context.Context, groupJID JID, name string) error {
	return groups.SetGroupName(s.Client, ctx, groupJID, name)
}

func (s *Client) SetGroupDescription(ctx context.Context, groupJID JID, desc string) error {
	return groups.SetGroupDescription(s.Client, ctx, groupJID, desc)
}

func (s *Client) AddParticipant(ctx context.Context, groupJID, userJID JID) error {
	return groups.AddParticipant(s.Client, ctx, groupJID, userJID)
}

func (s *Client) RemoveParticipant(ctx context.Context, groupJID, userJID JID) error {
	return groups.RemoveParticipant(s.Client, ctx, groupJID, userJID)
}

func (s *Client) PromoteParticipant(ctx context.Context, groupJID, userJID JID) error {
	return groups.PromoteParticipant(s.Client, ctx, groupJID, userJID)
}

func (s *Client) DemoteParticipant(ctx context.Context, groupJID, userJID JID) error {
	return groups.DemoteParticipant(s.Client, ctx, groupJID, userJID)
}

// ---------------------------------------------------------------------------
// Channels / privacy methods
// ---------------------------------------------------------------------------

func (s *Client) NewsletterPost(ctx context.Context, newsletterJID JID, text string) error {
	_, err := channels.NewsletterPost(s.Client, ctx, newsletterJID, text)
	return err
}

func (s *Client) NewsletterGetMetadata(ctx context.Context, newsletterJID JID) (*NewsletterMetadata, error) {
	return channels.NewsletterGetMetadata(s.Client, ctx, newsletterJID)
}

func (s *Client) SetBotPresence(ctx context.Context, available bool) error {
	return privacy.SetBotPresence(s.Client, ctx, available)
}

func (s *Client) SimulateTyping(ctx context.Context, chat JID, duration time.Duration) error {
	return privacy.SimulateTyping(s.Client, ctx, chat, duration)
}

func (s *Client) MarkReadSimple(ctx context.Context, chat JID, messageIDs []MessageID) error {
	return privacy.MarkReadSimple(s.Client, ctx, chat, messageIDs)
}

// PairPhoneWithCode pairs with a phone using a custom 8-character pairing
// code (e.g. "DONGTUBE" → shown as "DONG-TUBE"). Pass "" for a random code.
func (s *Client) PairPhoneWithCode(ctx context.Context, phone, customCode string) (string, error) {
	return s.Client.PairPhoneWithCode(ctx, phone, customCode)
}

// ---------------------------------------------------------------------------
// Type aliases (engine + domains)
// ---------------------------------------------------------------------------

type (
	Config                = protocol.Config
	MessageEvent          = protocol.MessageEvent
	MessageOptions        = protocol.MessageOptions
	Button                = protocol.Button
	ButtonType            = protocol.ButtonType
	Section               = protocol.Section
	SectionRow            = protocol.SectionRow
	NewsletterContext     = protocol.NewsletterContext
	MediaUpload           = protocol.MediaUpload
	LIDResolver           = protocol.LIDResolver
	RateLimiter           = protocol.RateLimiter
	MessageDeduplicator   = protocol.MessageDeduplicator
	AntiBanEngine         = protocol.AntiBanEngine
	AntiBanConfig         = protocol.AntiBanConfig
	AntiSpamGuard         = protocol.AntiSpamGuard
	AutoReconnectManager  = protocol.AutoReconnectManager
	SocketHealthMonitor   = protocol.SocketHealthMonitor
	PairingManager        = protocol.PairingManager
	RetrySpiralingTracker = protocol.RetrySpiralingTracker
	JID                   = protocol.JID
	MessageID             = protocol.MessageID
	GroupInfo             = protocol.GroupInfo
	GroupParticipant      = protocol.GroupParticipant
	NewsletterMetadata    = protocol.NewsletterMetadata
	PrivacySettingType    = protocol.PrivacySettingType
	PrivacySettings       = protocol.PrivacySettings
	AIBadgeOptions        = messaging.AIBadgeOptions
	LinkPreviewInfo       = messaging.LinkPreviewInfo
	SingleSelectSection   = buttons.SingleSelectSection
	NativeFlowButton      = buttons.NativeFlowButton
	MediaUploadLRUCache   = media.MediaUploadLRUCache
	UploadedMedia         = privacy.UploadedMedia
	GroupStatusPayload    = privacy.GroupStatusPayload
)

// Button type constants.
const (
	ButtonQuickReply  = protocol.ButtonQuickReply
	ButtonCTAURL      = protocol.ButtonCTAURL
	ButtonPhoneNumber = protocol.ButtonPhoneNumber
	ButtonCopyText    = protocol.ButtonCopyText
	ButtonSection     = protocol.ButtonSection
)

// Shared server constants.
const (
	DefaultUserServer = protocol.DefaultUserServer
	HiddenUserServer  = protocol.HiddenUserServer
	GroupServer       = protocol.GroupServer
)

// ---------------------------------------------------------------------------
// Functional-style API (advanced use)
// ---------------------------------------------------------------------------

var (
	NewRateLimiter         = protocol.NewRateLimiter
	NewMessageDeduplicator = protocol.NewMessageDeduplicator
	NewLIDResolver         = protocol.NewLIDResolver
	DefaultConfig          = protocol.DefaultConfig
	ParseJID               = protocol.ParseJID
	NewJID                 = protocol.NewJID
	IsGroup                = protocol.IsGroup
	IsDM                   = protocol.IsDM
	ParseMessageEvent      = protocol.ParseMessageEvent
	ParseCommand           = protocol.ParseCommand
	FormatDuration         = protocol.FormatDuration
	GetSmallBuffer         = protocol.GetSmallBuffer
	PutSmallBuffer         = protocol.PutSmallBuffer
	GetMediumBuffer        = protocol.GetMediumBuffer
	PutMediumBuffer        = protocol.PutMediumBuffer
	GetLargeBuffer         = protocol.GetLargeBuffer
	PutLargeBuffer         = protocol.PutLargeBuffer

	SendText               = messaging.SendText
	SendTextMessage        = messaging.SendTextMessage
	SendImage              = messaging.SendImage
	SendVideo              = messaging.SendVideo
	SendAudio              = messaging.SendAudio
	SendDocument           = messaging.SendDocument
	SendButtons            = messaging.SendButtons
	SendReaction           = messaging.SendReaction
	SendMention            = messaging.SendMention
	ExtractPureText        = messaging.ExtractPureText
	ExtractMentions        = messaging.ExtractMentions
	SendTextWithFakeReply  = messaging.SendTextWithFakeReply
	SendTextWithNewsletter = messaging.SendTextWithNewsletter
	NewButtonBuilder       = messaging.NewButtonBuilder
	RevokeMessage          = messaging.RevokeMessage
	HandleButtonResponse   = buttons.HandleButtonResponse

	DownloadMedia            = media.DownloadMedia
	SendSticker              = messaging.SendSticker
	ExtractUnderlyingMessage = protocol.ExtractUnderlyingMessage
	AddStickerMetadata       = media.AddStickerMetadata
	UploadMedia              = media.UploadMedia
	FetchURL                 = media.FetchURL
	NewMediaUploadLRUCache   = media.NewMediaUploadLRUCache
	GetGlobalUploadCache     = media.GetGlobalUploadCache

	GetGroupInfo               = groups.GetGroupInfo
	GetGroupParticipants       = groups.GetGroupParticipants
	IsUserAdmin                = groups.IsUserAdmin
	JoinGroupWithInviteLink    = groups.JoinGroupWithInviteLink
	GetGroupInfoFromInviteLink = groups.GetGroupInfoFromInviteLink
	GetGroupInviteLink         = groups.GetGroupInviteLink
	GetJoinedGroups            = groups.GetJoinedGroups
	LeaveGroup                 = groups.LeaveGroup
	SetGroupName               = groups.SetGroupName
	SetGroupDescription        = groups.SetGroupDescription
	AddParticipant             = groups.AddParticipant
	RemoveParticipant          = groups.RemoveParticipant
	PromoteParticipant         = groups.PromoteParticipant
	DemoteParticipant          = groups.DemoteParticipant
	ApproveGroupRequests       = groups.ApproveGroupRequests
	TagAll                     = groups.TagAll
	ResolveParticipantPN       = groups.ResolveParticipantPN
	BuildBizAdditionalNodes    = groups.BuildBizAdditionalNodes
	IsBotAdmin                 = groups.IsBotAdmin

	NewsletterPost        = channels.NewsletterPost
	NewsletterPostImage   = channels.NewsletterPostImage
	NewsletterGetMetadata = channels.NewsletterGetMetadata
	NewsletterCreate      = channels.NewsletterCreate

	GetPrivacySettings      = privacy.GetPrivacySettings
	SetPrivacySetting       = privacy.SetPrivacySetting
	SetBotPresence          = privacy.SetBotPresence
	SimulateTyping          = privacy.SimulateTyping
	SimulateRecording       = privacy.SimulateRecording
	MarkReadSimple          = privacy.MarkReadSimple
	GetUserProfile          = privacy.GetUserProfile
	GetContactInfo          = privacy.GetContactInfo
	GetAllContacts          = privacy.GetAllContacts
	FetchUsername           = privacy.FetchUsername
	SetAboutStatus          = privacy.SetAboutStatus
	GetBlocklist            = privacy.GetBlocklist
	BuildGroupStatusMessage = privacy.BuildGroupStatusMessage

	GetBusinessProfile     = business.GetBusinessProfile
	GetBusinessCatalog     = business.GetBusinessCatalog
	GetBusinessCollections = business.GetBusinessCollections
	SendOrderDetails       = business.SendOrderDetails
)
