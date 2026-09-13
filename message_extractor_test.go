package meowbail_test

import (
	"testing"

	meowbail "github.com/hanzywaifu-coder/dongtube-meowbail"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

func TestExtractPureText(t *testing.T) {
	// Plain conversation
	msg1 := &waE2E.Message{
		Conversation: proto.String("Halo Dongtube"),
	}
	if text := meowbail.ExtractPureText(msg1); text != "Halo Dongtube" {
		t.Fatalf("expected 'Halo Dongtube', got '%s'", text)
	}

	// ExtendedTextMessage
	msg2 := &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: proto.String(".menu all"),
		},
	}
	if text := meowbail.ExtractPureText(msg2); text != ".menu all" {
		t.Fatalf("expected '.menu all', got '%s'", text)
	}

	// Ephemeral unwrapping
	msg3 := &waE2E.Message{
		EphemeralMessage: &waE2E.FutureProofMessage{
			Message: &waE2E.Message{
				Conversation: proto.String("Ephemeral secret"),
			},
		},
	}
	if text := meowbail.ExtractPureText(msg3); text != "Ephemeral secret" {
		t.Fatalf("expected 'Ephemeral secret', got '%s'", text)
	}
}

func TestExtractMentions(t *testing.T) {
	text := "Halo @628123456789 dan @628987654321, tolong cek."
	mentions := meowbail.ExtractMentions(text)

	if len(mentions) != 2 {
		t.Fatalf("expected 2 mentions, got %d", len(mentions))
	}

	expected1 := "628123456789@s.whatsapp.net"
	expected2 := "628987654321@s.whatsapp.net"

	if mentions[0].String() != expected1 || mentions[1].String() != expected2 {
		t.Fatalf("unexpected parsed mentions: %v", mentions)
	}
}
