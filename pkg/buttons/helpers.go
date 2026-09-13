package buttons

import (
	waE2E "go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"

	protocol "github.com/hanzywaifu-coder/dongtube-meowbail/internal/protocol"
)

// newsletterContext mirrors messaging.buildNewsletterContext (kept local to
// avoid an import cycle; buttons send with the newsletter branding).
func newsletterContext(cfg *protocol.Config) *waE2E.ContextInfo {
	if cfg != nil && cfg.DefaultFakeReply != nil {
		return proto.Clone(cfg.DefaultFakeReply).(*waE2E.ContextInfo)
	}
	if cfg == nil || cfg.NewsletterJID == "" {
		return &waE2E.ContextInfo{}
	}
	return &waE2E.ContextInfo{
		IsForwarded:     proto.Bool(false),
		ForwardingScore: proto.Uint32(0),
		BusinessMessageForwardInfo: &waE2E.ContextInfo_BusinessMessageForwardInfo{
			BusinessOwnerJID: proto.String(cfg.BusinessOwnerJID),
		},
		ForwardedNewsletterMessageInfo: &waE2E.ContextInfo_ForwardedNewsletterMessageInfo{
			NewsletterJID:  proto.String(cfg.NewsletterJID),
			NewsletterName: proto.String(cfg.NewsletterName),
		},
	}
}
