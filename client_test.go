package meowbail_test

import (
	"context"
	"testing"

	meowbail "github.com/hanzywaifu-coder/dongtube-meowbail"
)

func TestFluentBuilder(t *testing.T) {
	builder := meowbail.NewButtonBuilder().
		SetBody("Test body").
		SetFooter("Test footer").
		AddCTAURL("Kunjungi Web", "https://dongtube.cyou").
		AddQuickReply("Bantuan", ".help").
		AddCopyCode("Salin Token", "ABC-123")

	msg, err := builder.Build()
	if err != nil {
		t.Fatalf("FluentButtonBuilder.Build() failed: %v", err)
	}
	if msg == nil {
		t.Fatal("Expected non-nil message")
	}
}

func TestAIBadgeFormat(t *testing.T) {
	opts := meowbail.AIBadgeOptions{
		PersonaID: "test-persona",
		ModelName: "Llama-3",
	}
	if opts.PersonaID != "test-persona" {
		t.Errorf("Unexpected PersonaID: %s", opts.PersonaID)
	}
}

func TestContext(t *testing.T) {
	ctx := context.Background()
	if ctx == nil {
		t.Fatal("context is nil")
	}
}
