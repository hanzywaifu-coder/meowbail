package privacy

import (
	"fmt"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

// GroupStatusType mendefinisikan jenis status grup yang didukung
type GroupStatusType string

const (
	StatusImage GroupStatusType = "image"
	StatusVideo GroupStatusType = "video"
	StatusAudio GroupStatusType = "audio"
	StatusText  GroupStatusType = "text"
)

// GroupStatusPayload menyimpan data media atau teks untuk story grup
type GroupStatusPayload struct {
	Type     GroupStatusType
	Caption  string
	Uploaded *UploadedMedia
}

// UploadedMedia menyimpan informasi hasil upload media ke server WhatsApp
type UploadedMedia struct {
	URL           string
	DirectPath    string
	MediaKey      []byte
	FileEncSHA256 []byte
	FileSHA256    []byte
	FileLength    uint64
	Mimetype      string
}

// BuildGroupStatusMessage membungkus media menjadi GroupStatusMessageV2 yang sah
func BuildGroupStatusMessage(payload GroupStatusPayload) (*waE2E.Message, error) {
	var inner *waE2E.Message

	switch payload.Type {
	case StatusImage:
		if payload.Uploaded == nil {
			return nil, fmt.Errorf("uploaded media metadata is required for status image")
		}
		inner = &waE2E.Message{
			ImageMessage: &waE2E.ImageMessage{
				URL:           &payload.Uploaded.URL,
				Mimetype:      proto.String("image/jpeg"),
				Caption:       proto.String(payload.Caption),
				FileEncSHA256: payload.Uploaded.FileEncSHA256,
				FileSHA256:    payload.Uploaded.FileSHA256,
				FileLength:    &payload.Uploaded.FileLength,
				DirectPath:    &payload.Uploaded.DirectPath,
				MediaKey:      payload.Uploaded.MediaKey,
			},
		}
	case StatusVideo:
		if payload.Uploaded == nil {
			return nil, fmt.Errorf("uploaded media metadata is required for status video")
		}
		inner = &waE2E.Message{
			VideoMessage: &waE2E.VideoMessage{
				URL:           &payload.Uploaded.URL,
				Mimetype:      proto.String("video/mp4"),
				Caption:       proto.String(payload.Caption),
				FileEncSHA256: payload.Uploaded.FileEncSHA256,
				FileSHA256:    payload.Uploaded.FileSHA256,
				FileLength:    &payload.Uploaded.FileLength,
				DirectPath:    &payload.Uploaded.DirectPath,
				MediaKey:      payload.Uploaded.MediaKey,
			},
		}
	case StatusAudio:
		if payload.Uploaded == nil {
			return nil, fmt.Errorf("uploaded media metadata is required for status audio")
		}
		inner = &waE2E.Message{
			AudioMessage: &waE2E.AudioMessage{
				URL:           &payload.Uploaded.URL,
				Mimetype:      proto.String("audio/mp4"),
				FileEncSHA256: payload.Uploaded.FileEncSHA256,
				FileSHA256:    payload.Uploaded.FileSHA256,
				FileLength:    &payload.Uploaded.FileLength,
				DirectPath:    &payload.Uploaded.DirectPath,
				MediaKey:      payload.Uploaded.MediaKey,
			},
		}
	case StatusText:
		if payload.Caption == "" {
			return nil, fmt.Errorf("caption text cannot be empty for text status")
		}
		inner = &waE2E.Message{
			ExtendedTextMessage: &waE2E.ExtendedTextMessage{
				Text: proto.String(payload.Caption),
			},
		}
	default:
		return nil, fmt.Errorf("unsupported group status type: %s", payload.Type)
	}

	return &waE2E.Message{
		GroupStatusMessageV2: &waE2E.FutureProofMessage{
			Message: inner,
		},
	}, nil
}
