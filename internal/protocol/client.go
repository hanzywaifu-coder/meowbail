package protocol

import (
	"context"
	"encoding/json"
	"log"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

// NewClient creates a new dongtube-meowbail client
func NewClient(device interface{}, logger interface{}, config ...*Config) *Client {
	cfg := DefaultConfig()
	if len(config) > 0 && config[0] != nil {
		cfg = config[0]
	}
	// Use whatsmeow client as base
	var cli *whatsmeow.Client
	if device != nil {
		// Type assert to whatsmeow device store
		if ds, ok := device.(*whatsmeow.Client); ok {
			cli = ds
		}
	}
	if cli == nil {
		log.Fatal("dongtube-meowbail: invalid device store")
	}
	cli.EnableAutoReconnect = cfg.AutoReconnect
	c := &Client{
		Client:        cli,
		config:        cfg,
		LIDResolver:   NewLIDResolver(),
		RetryTracker:  NewRetrySpiralingTracker(5),
		AntiBanEngine: NewAntiBanEngine(DefaultAntiBanConfig()),
	}
	return c
}

// NewClientFromWhatsmeow wraps an existing whatsmeow client
func NewClientFromWhatsmeow(cli *whatsmeow.Client, config ...*Config) *Client {
	cfg := DefaultConfig()
	if len(config) > 0 && config[0] != nil {
		cfg = config[0]
	}
	cli.EnableAutoReconnect = cfg.AutoReconnect
	c := &Client{
		Client:        cli,
		config:        cfg,
		LIDResolver:   NewLIDResolver(),
		RetryTracker:  NewRetrySpiralingTracker(5),
		AntiBanEngine: NewAntiBanEngine(DefaultAntiBanConfig()),
	}
	return c
}

// SetNewsletter sets the newsletter context for all messages
func (c *Client) SetNewsletter(jid, name string) {
	c.Config().NewsletterJID = jid
	c.Config().NewsletterName = name
}

// SetBusinessOwner sets the business owner JID
func (c *Client) SetBusinessOwner(jid string) {
	c.Config().BusinessOwnerJID = jid
}

// SetDefaultFakeReply sets the global fake reply context info for all outgoing messages (text, media, etc)
func (c *Client) SetDefaultFakeReply(ctxInfo *waE2E.ContextInfo) {
	if c.Config() != nil {
		c.Config().DefaultFakeReply = ctxInfo
	}
}

// AddEventHandler wraps whatsmeow's AddEventHandler
func (c *Client) AddEventHandler(handler func(evt interface{})) uint32 {
	return c.Client.AddEventHandler(handler)
}

// Connect wraps whatsmeow's Connect and performs auto-follow if configured
func (c *Client) Connect(ctx context.Context) error {
	// Sync versi WhatsApp Web terbaru secara otomatis untuk mencegah 405 out of date
	go func() {
		defer func() { _ = recover() }()
		ver, err := whatsmeow.GetLatestVersion(context.Background(), nil)
		if err == nil && ver != nil {
			store.SetWAVersion(*ver)
		}
	}()
	err := c.Client.Connect()
	if err != nil {
		return err
	}
	// Auto-follow channels event-driven saat event LoggedIn / Connected masuk (mengeliminasi polling loop)
	c.Client.AddEventHandler(func(evt interface{}) {
		switch evt.(type) {
		case *events.Connected, *events.PairSuccess:
			if !c.Client.IsLoggedIn() {
				return
			}
			// Mark the device as online on every (re)connect, otherwise the
			// linked-device list shows it as offline after the first session.
			go func() {
				defer func() { _ = recover() }()
				_ = c.Client.SendPresence(context.Background(), types.PresenceAvailable)
			}()
			go func() {
				defer func() { _ = recover() }()
				var jids []string
				if c.Config().NewsletterJID != "" {
					jids = append(jids, c.Config().NewsletterJID)
				}
				jids = append(jids, c.Config().AutoFollowJIDs...)
				seen := make(map[string]bool)
				for _, rawJID := range jids {
					if rawJID == "" || seen[rawJID] {
						continue
					}
					seen[rawJID] = true
					parsed, err := types.ParseJID(rawJID)
					if err != nil {
						continue
					}
					_ = c.Client.FollowNewsletter(context.Background(), parsed)
				}
			}()
		}
	})
	return nil
}

// Disconnect wraps whatsmeow's Disconnect
func (c *Client) Disconnect() {
	c.Client.Disconnect()
}

// IsConnected wraps whatsmeow's IsConnected
func (c *Client) IsConnected() bool {
	return c.Client.IsConnected()
}

// IsLoggedIn wraps whatsmeow's IsLoggedIn
func (c *Client) IsLoggedIn() bool {
	return c.Client.IsLoggedIn()
}

// ParseMessageEvent parses a raw event into a MessageEvent
func ParseMessageEvent(evt interface{}) *MessageEvent {
	e, ok := evt.(*events.Message)
	if !ok {
		return nil
	}
	text := ""
	msg := e.Message
	if msg == nil {
		return nil
	}
	// Unwrap ephemeral message
	if msg.EphemeralMessage != nil && msg.EphemeralMessage.Message != nil {
		msg = msg.EphemeralMessage.Message
	}
	// Unwrap view once message
	if msg.ViewOnceMessage != nil && msg.ViewOnceMessage.Message != nil {
		msg = msg.ViewOnceMessage.Message
	}
	if msg.ViewOnceMessageV2 != nil && msg.ViewOnceMessageV2.Message != nil {
		msg = msg.ViewOnceMessageV2.Message
	}
	if msg.ViewOnceMessageV2Extension != nil && msg.ViewOnceMessageV2Extension.Message != nil {
		msg = msg.ViewOnceMessageV2Extension.Message
	}
	// Unwrap document with caption
	if msg.DocumentWithCaptionMessage != nil && msg.DocumentWithCaptionMessage.Message != nil {
		msg = msg.DocumentWithCaptionMessage.Message
	}
	if msg.GetConversation() != "" {
		text = msg.GetConversation()
	} else if msg.ExtendedTextMessage != nil {
		text = msg.ExtendedTextMessage.GetText()
	} else if msg.ImageMessage != nil {
		text = msg.ImageMessage.GetCaption()
	} else if msg.VideoMessage != nil {
		text = msg.VideoMessage.GetCaption()
	} else if msg.DocumentMessage != nil {
		text = msg.DocumentMessage.GetCaption()
	} else if msg.ButtonsResponseMessage != nil {
		text = msg.ButtonsResponseMessage.GetSelectedButtonID()
	} else if msg.ListResponseMessage != nil && msg.ListResponseMessage.GetSingleSelectReply() != nil {
		text = msg.ListResponseMessage.GetSingleSelectReply().GetSelectedRowID()
	} else if msg.InteractiveResponseMessage != nil {
		if nativeFlow := msg.InteractiveResponseMessage.GetNativeFlowResponseMessage(); nativeFlow != nil {
			var parsed map[string]interface{}
			if err := json.Unmarshal([]byte(nativeFlow.GetParamsJSON()), &parsed); err == nil {
				if id, ok := parsed["id"].(string); ok && id != "" {
					text = id
				} else if id, ok := parsed["selected_row_id"].(string); ok && id != "" {
					text = id
				}
			}
		}
	} else if msg.AudioMessage != nil {
		text = "" // audio has no text
	} else if msg.StickerMessage != nil {
		text = "" // sticker has no text
	}
	sender := e.Info.Sender
	if sender.Server == types.HiddenUserServer || sender.Server == "lid" {
		if e.Info.SenderAlt.Server == types.DefaultUserServer {
			sender = e.Info.SenderAlt
		}
	}
	return &MessageEvent{
		Message:  e,
		Sender:   sender,
		Chat:     e.Info.Chat,
		Text:     text,
		IsGroup:  e.Info.Chat.Server == "g.us",
		IsFromMe: e.Info.IsFromMe,
	}
}

// GetConfig returns the client configuration
func (c *Client) GetConfig() *Config {
	return c.Config()
}
