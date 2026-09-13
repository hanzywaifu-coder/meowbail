// wtypes.go re-exports the shared whatsmeow types so domain packages can
// reference them via the single protocol package (e.g. protocol.JID).
package protocol

import (
	"go.mau.fi/whatsmeow/types"
)

type (
	JID                     = types.JID
	MessageID               = types.MessageID
	MessageServerID         = types.MessageServerID
	ContactInfo             = types.ContactInfo
	Blocklist               = types.Blocklist
	GroupInfo               = types.GroupInfo
	GroupLinkTarget         = types.GroupLinkTarget
	GroupParent             = types.GroupParent
	GroupParticipant        = types.GroupParticipant
	GroupParticipantRequest = types.GroupParticipantRequest
	NewsletterMessage       = types.NewsletterMessage
	NewsletterMetadata      = types.NewsletterMetadata
	PrivacySetting          = types.PrivacySetting
	PrivacySettingType      = types.PrivacySettingType
	PrivacySettings         = types.PrivacySettings
	StatusPrivacy           = types.StatusPrivacy
	SetStatusInput          = types.SetStatusInput
	UserInfo                = types.UserInfo
	ProfilePictureInfo      = types.ProfilePictureInfo
	ChatPresence            = types.ChatPresence
	ChatPresenceMedia       = types.ChatPresenceMedia
	Presence                = types.Presence
	ReceiptType             = types.ReceiptType
)

const (
	DefaultUserServer              = types.DefaultUserServer
	HiddenUserServer               = types.HiddenUserServer
	GroupServer                    = types.GroupServer
	PresenceAvailable              = types.PresenceAvailable
	PresenceUnavailable            = types.PresenceUnavailable
	ChatPresenceComposing          = types.ChatPresenceComposing
	ChatPresencePaused             = types.ChatPresencePaused
	ChatPresenceMediaText          = types.ChatPresenceMediaText
	ChatPresenceMediaAudio         = types.ChatPresenceMediaAudio
	PrivacySettingAll              = types.PrivacySettingAll
	PrivacySettingContacts         = types.PrivacySettingContacts
	PrivacySettingKnown            = types.PrivacySettingKnown
	PrivacySettingMatchLastSeen    = types.PrivacySettingMatchLastSeen
	PrivacySettingNone             = types.PrivacySettingNone
	PrivacySettingTypeCallAdd      = types.PrivacySettingTypeCallAdd
	PrivacySettingTypeGroupAdd     = types.PrivacySettingTypeGroupAdd
	PrivacySettingTypeLastSeen     = types.PrivacySettingTypeLastSeen
	PrivacySettingTypeOnline       = types.PrivacySettingTypeOnline
	PrivacySettingTypeProfile      = types.PrivacySettingTypeProfile
	PrivacySettingTypeReadReceipts = types.PrivacySettingTypeReadReceipts
	PrivacySettingTypeStatus       = types.PrivacySettingTypeStatus
	ReceiptTypePlayed              = types.ReceiptTypePlayed
	GroupMemberAddModeAdmin        = types.GroupMemberAddModeAdmin
	GroupMemberAddModeAllMember    = types.GroupMemberAddModeAllMember
)

// JID values and constructor re-exported as vars (JIDs are structs, not constants).
var (
	StatusBroadcastJID = types.StatusBroadcastJID
	EmptyJID           = types.EmptyJID
	ServerJID          = types.ServerJID
	NewJID             = types.NewJID
)
