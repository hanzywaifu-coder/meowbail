package protocol

import (
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types/events"
)

// UnwrappedMessage menyimpan pesan asli yang telah di-unmask dari pembungkus (ViewOnce / Ephemeral / Document)
type UnwrappedMessage struct {
	OriginalType string
	IsViewOnce   bool
	IsEphemeral  bool
	Message      *waE2E.Message
}

// ExtractUnderlyingMessage membongkar lapisan pesan WhatsApp modern
// seperti ViewOnceMessage, ViewOnceMessageV2, EphemeralMessage, atau TemplateMessage
func ExtractUnderlyingMessage(msg *waE2E.Message) *UnwrappedMessage {
	if msg == nil {
		return nil
	}
	result := &UnwrappedMessage{
		OriginalType: "standard",
		Message:      msg,
	}
	// 1. Unwrap Ephemeral Message (pesan sementara)
	if msg.EphemeralMessage != nil && msg.EphemeralMessage.Message != nil {
		result.IsEphemeral = true
		result.OriginalType = "ephemeral"
		msg = msg.EphemeralMessage.Message
	}
	// 2. Unwrap ViewOnce Message V1 (sekali lihat gambar/video)
	if msg.ViewOnceMessage != nil && msg.ViewOnceMessage.Message != nil {
		result.IsViewOnce = true
		result.OriginalType = "view_once"
		msg = msg.ViewOnceMessage.Message
	}
	// 3. Unwrap ViewOnce Message V2
	if msg.ViewOnceMessageV2 != nil && msg.ViewOnceMessageV2.Message != nil {
		result.IsViewOnce = true
		result.OriginalType = "view_once_v2"
		msg = msg.ViewOnceMessageV2.Message
	}
	// 4. Unwrap ViewOnce Message V2 Extension
	if msg.ViewOnceMessageV2Extension != nil && msg.ViewOnceMessageV2Extension.Message != nil {
		result.IsViewOnce = true
		result.OriginalType = "view_once_v2_extension"
		msg = msg.ViewOnceMessageV2Extension.Message
	}
	// 5. Unwrap Document With Caption Message
	if msg.DocumentWithCaptionMessage != nil && msg.DocumentWithCaptionMessage.Message != nil {
		msg = msg.DocumentWithCaptionMessage.Message
	}
	result.Message = msg
	return result
}

// DetectDeletedMessage mendeteksi sinyal protokol pencabutan/penghapusan pesan oleh lawan bicara (Anti-Delete)
func DetectDeletedMessage(evt interface{}) (targetID string, isRevoked bool) {
	e, ok := evt.(*events.Message)
	if !ok || e.Message == nil {
		return "", false
	}
	if e.Message.ProtocolMessage != nil {
		protoType := e.Message.ProtocolMessage.GetType()
		if protoType == waE2E.ProtocolMessage_REVOKE {
			key := e.Message.ProtocolMessage.GetKey()
			if key != nil && key.ID != nil {
				return *key.ID, true
			}
		}
	}
	return "", false
}
