package protocol

import (
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"time"
)

// Button types
type ButtonType int

const (
	ButtonQuickReply ButtonType = iota
	ButtonCTAURL
	ButtonPhoneNumber
	ButtonCopyText
	ButtonSection
)

// Button represents a WhatsApp button
type Button struct {
	Type        ButtonType
	ID          string
	Text        string
	URL         string
	Phone       string
	DisplayText string
	Sections    []Section
}

// Section represents a list section
type Section struct {
	Title string
	Rows  []SectionRow
}

// SectionRow represents a row in a section
type SectionRow struct {
	Title       string
	Description string
	ID          string
}

// NewsletterContext represents newsletter/forwarded message context
type NewsletterContext struct {
	NewsletterJID    string
	NewsletterName   string
	BusinessOwnerJID string
	IsForwarded      bool
	ForwardingScore  uint32
}

// MessageOptions contains options for sending messages
type MessageOptions struct {
	Buttons    []Button
	Newsletter *NewsletterContext
	ReplyTo    *types.MessageID
	Mentions   []string
	Quoted     *waE2E.Message
	Disappear  time.Duration
}

// MediaUpload represents uploaded media
type MediaUpload struct {
	URL           string
	Mimetype      string
	FileEncSHA256 []byte
	FileSHA256    []byte
	FileLength    uint64
	DirectPath    string
	MediaKey      []byte
}

// Client wraps whatsmeow client with Baileys-style features
type Client struct {
	*whatsmeow.Client
	config        *Config
	LIDResolver   *LIDResolver
	RetryTracker  *RetrySpiralingTracker
	AntiBanEngine *AntiBanEngine
}

// Config contains configuration for the client
type Config struct {
	NewsletterJID    string
	NewsletterName   string
	AutoFollowJIDs   []string
	BusinessOwnerJID string
	AutoReconnect    bool
	MaxRetryCount    int
	DefaultFakeReply *waE2E.ContextInfo
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		AutoReconnect: true,
		MaxRetryCount: 3,
	}
}

// Event wrapper
type MessageEvent struct {
	*events.Message
	Sender   types.JID
	Chat     types.JID
	Text     string
	IsGroup  bool
	IsFromMe bool
}
