package channels

import (
	"context"
	"fmt"
	protocol "github.com/hanzywaifu-coder/dongtube-meowbail/internal/protocol"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

// GroupStatusMedia holds the payload for group status (SWGC)
type GroupStatusMedia struct {
	Image []byte
	Video []byte
	Audio []byte
	Text  string
}

// SendGroupStatus uploads a story/status to a specific group JID (SWGC feature from Baileys)
func SendGroupStatus(c *protocol.Client, ctx context.Context, groupJID protocol.JID, media GroupStatusMedia) error {
	var innerMsg *waE2E.Message

	if len(media.Image) > 0 {
		resp, err := c.Client.Upload(ctx, media.Image, whatsmeow.MediaImage)
		if err != nil {
			return fmt.Errorf("upload image for group status failed: %w", err)
		}
		innerMsg = &waE2E.Message{
			ImageMessage: &waE2E.ImageMessage{
				URL:           &resp.URL,
				Mimetype:      proto.String("image/jpeg"),
				Caption:       proto.String(media.Text),
				FileEncSHA256: resp.FileEncSHA256,
				FileSHA256:    resp.FileSHA256,
				FileLength:    &resp.FileLength,
				DirectPath:    &resp.DirectPath,
				MediaKey:      resp.MediaKey,
			},
		}
	} else if len(media.Video) > 0 {
		resp, err := c.Client.Upload(ctx, media.Video, whatsmeow.MediaVideo)
		if err != nil {
			return fmt.Errorf("upload video for group status failed: %w", err)
		}
		innerMsg = &waE2E.Message{
			VideoMessage: &waE2E.VideoMessage{
				URL:           &resp.URL,
				Mimetype:      proto.String("video/mp4"),
				Caption:       proto.String(media.Text),
				FileEncSHA256: resp.FileEncSHA256,
				FileSHA256:    resp.FileSHA256,
				FileLength:    &resp.FileLength,
				DirectPath:    &resp.DirectPath,
				MediaKey:      resp.MediaKey,
			},
		}
	} else if len(media.Audio) > 0 {
		resp, err := c.Client.Upload(ctx, media.Audio, whatsmeow.MediaAudio)
		if err != nil {
			return fmt.Errorf("upload audio for group status failed: %w", err)
		}
		innerMsg = &waE2E.Message{
			AudioMessage: &waE2E.AudioMessage{
				URL:           &resp.URL,
				Mimetype:      proto.String("audio/mp4"),
				FileEncSHA256: resp.FileEncSHA256,
				FileSHA256:    resp.FileSHA256,
				FileLength:    &resp.FileLength,
				DirectPath:    &resp.DirectPath,
				MediaKey:      resp.MediaKey,
			},
		}
	} else if media.Text != "" {
		innerMsg = &waE2E.Message{
			ExtendedTextMessage: &waE2E.ExtendedTextMessage{
				Text: proto.String(media.Text),
			},
		}
	} else {
		return fmt.Errorf("empty status media or caption")
	}

	// Bungkus ke dalam GroupStatusMessageV2 (sesuai spesifikasi protobuf WhatsApp & dongtube-meowbail)
	statusMsg := &waE2E.Message{
		GroupStatusMessageV2: &waE2E.FutureProofMessage{
			Message: innerMsg,
		},
	}

	_, err := c.Client.SendMessage(ctx, groupJID, statusMsg)
	return err
}
