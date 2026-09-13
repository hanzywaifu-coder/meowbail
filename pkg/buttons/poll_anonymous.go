package buttons

import (
	"context"
	"fmt"
	protocol "github.com/hanzywaifu-coder/dongtube-meowbail/internal/protocol"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

// SendAnonymousPoll mengirim polling voting dengan fitur Anonymous / Voter Names Hidden (PollCreationMessageV3)
func SendAnonymousPoll(c *protocol.Client, ctx context.Context, chat protocol.JID, question string, options []string, selectableCount int) error {
	if len(options) < 2 {
		return fmt.Errorf("polling membutuhkan minimal 2 opsi")
	}

	pollOpts := make([]*waE2E.PollCreationMessage_Option, len(options))
	for i, opt := range options {
		pollOpts[i] = &waE2E.PollCreationMessage_Option{
			OptionName: proto.String(opt),
		}
	}

	if selectableCount <= 0 {
		selectableCount = 1
	}

	pollType := waE2E.PollType_POLL

	msg := &waE2E.Message{
		PollCreationMessageV3: &waE2E.PollCreationMessage{
			Name:                   proto.String(question),
			Options:                pollOpts,
			SelectableOptionsCount: proto.Uint32(uint32(selectableCount)),
			PollType:               &pollType,
		},
		MessageContextInfo: &waE2E.MessageContextInfo{
			MessageSecret: []byte(c.Client.GenerateMessageID()[:32]),
		},
	}

	_, err := c.Client.SendMessage(ctx, chat, msg)
	return err
}
