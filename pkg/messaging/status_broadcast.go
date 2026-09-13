package messaging

import (
	"context"
	"fmt"
	protocol "github.com/hanzywaifu-coder/dongtube-meowbail/internal/protocol"

	media "github.com/hanzywaifu-coder/dongtube-meowbail/pkg/media"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

// SendStatusBroadcastText membuat status/story WhatsApp resmi (status@broadcast) berupa teks
func SendStatusBroadcastText(c *protocol.Client, ctx context.Context, text string, backgroundColorARGB uint32, font int32) error {
	if text == "" {
		return fmt.Errorf("teks status kosong")
	}

	if backgroundColorARGB == 0 {
		backgroundColorARGB = 0xFF243B55 // Dongtube gradient midnight blue
	}

	msg := &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text:           proto.String(text),
			BackgroundArgb: proto.Uint32(backgroundColorARGB),
			Font:           (*waE2E.ExtendedTextMessage_FontType)(&font),
		},
	}

	_, err := c.Client.SendMessage(ctx, protocol.StatusBroadcastJID, msg)
	return err
}

// SendStatusBroadcastMedia membuat status/story WhatsApp resmi (status@broadcast) berupa gambar atau video
func SendStatusBroadcastMedia(c *protocol.Client, ctx context.Context, mediaBytes []byte, isVideo bool, caption string) error {
	if len(mediaBytes) == 0 {
		return fmt.Errorf("media status kosong")
	}

	if isVideo {
		uploaded, err := media.UploadMedia(c, ctx, mediaBytes, whatsmeow.MediaVideo)
		if err != nil {
			return fmt.Errorf("upload video status: %w", err)
		}
		msg := &waE2E.Message{
			VideoMessage: &waE2E.VideoMessage{
				URL:           proto.String(uploaded.URL),
				DirectPath:    proto.String(uploaded.DirectPath),
				MediaKey:      uploaded.MediaKey,
				Mimetype:      proto.String("video/mp4"),
				FileEncSHA256: uploaded.FileEncSHA256,
				FileSHA256:    uploaded.FileSHA256,
				FileLength:    proto.Uint64(uploaded.FileLength),
				Caption:       proto.String(caption),
			},
		}
		_, err = c.Client.SendMessage(ctx, protocol.StatusBroadcastJID, msg)
		return err
	}

	uploaded, err := media.UploadMedia(c, ctx, mediaBytes, whatsmeow.MediaImage)
	if err != nil {
		return fmt.Errorf("upload image status: %w", err)
	}

	msg := &waE2E.Message{
		ImageMessage: &waE2E.ImageMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			Mimetype:      proto.String("image/jpeg"),
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uploaded.FileLength),
			Caption:       proto.String(caption),
		},
	}

	_, err = c.Client.SendMessage(ctx, protocol.StatusBroadcastJID, msg)
	if err != nil {
		return err
	}
	return nil
}

// SendStatusBroadcastAudio membuat status/story WhatsApp resmi suara/audio (VN Status)
func SendStatusBroadcastAudio(c *protocol.Client, ctx context.Context, audioBytes []byte, durationSeconds uint32, waveform []byte, backgroundColorARGB uint32) error {
	if len(audioBytes) == 0 {
		return fmt.Errorf("audio status kosong")
	}

	if backgroundColorARGB == 0 {
		backgroundColorARGB = 0xFF243B55
	}

	if durationSeconds == 0 {
		durationSeconds = 1
	}

	if len(waveform) == 0 {
		// Default simulated waveform jika kosong (64 bytes)
		waveform = make([]byte, 64)
		for i := range waveform {
			waveform[i] = byte((i * 17) % 100)
		}
	}

	uploaded, err := media.UploadMedia(c, ctx, audioBytes, whatsmeow.MediaAudio)
	if err != nil {
		return fmt.Errorf("upload audio status: %w", err)
	}

	msg := &waE2E.Message{
		AudioMessage: &waE2E.AudioMessage{
			URL:            proto.String(uploaded.URL),
			DirectPath:     proto.String(uploaded.DirectPath),
			MediaKey:       uploaded.MediaKey,
			Mimetype:       proto.String("audio/ogg; codecs=opus"),
			FileEncSHA256:  uploaded.FileEncSHA256,
			FileSHA256:     uploaded.FileSHA256,
			FileLength:     proto.Uint64(uploaded.FileLength),
			Seconds:        proto.Uint32(durationSeconds),
			PTT:            proto.Bool(true),
			Waveform:       waveform,
			BackgroundArgb: proto.Uint32(backgroundColorARGB),
		},
	}

	_, err = c.Client.SendMessage(ctx, protocol.StatusBroadcastJID, msg)
	return err
}
