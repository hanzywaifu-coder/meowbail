package messaging

import (
	"context"
	"fmt"
	protocol "github.com/hanzywaifu-coder/dongtube-meowbail/internal/protocol"
	"time"

	media "github.com/hanzywaifu-coder/dongtube-meowbail/pkg/media"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waCommon"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

// SendWhatsAppMediaAlbum mengirim banyak foto/video menjadi 1 kesatuan Album Bubble resmi WhatsApp
func SendWhatsAppMediaAlbum(c *protocol.Client, ctx context.Context, chat protocol.JID, images [][]byte, captions []string) ([]protocol.MessageID, error) {
	if len(images) == 0 {
		return nil, fmt.Errorf("album kosong")
	}

	// 1. Jika hanya 1 foto, kirim foto biasa
	if len(images) == 1 {
		cpt := ""
		if len(captions) > 0 {
			cpt = captions[0]
		}
		err := SendImage(c, ctx, chat, images[0], cpt)
		return nil, err
	}

	// 2. Buat Parent Album Message (Album Header)
	albumParentMsg := &waE2E.Message{
		AlbumMessage: &waE2E.AlbumMessage{
			ExpectedImageCount: proto.Uint32(uint32(len(images))),
			ExpectedVideoCount: proto.Uint32(0),
			ContextInfo:        BuildNewsletterContext(c.Config()),
		},
	}

	parentResp, err := c.Client.SendMessage(ctx, chat, albumParentMsg)
	if err != nil {
		// Fallback kirim berkala jika AlbumMessage tidak didukung server
		for i, img := range images {
			cpt := ""
			if i < len(captions) {
				cpt = captions[i]
			}
			_ = SendImage(c, ctx, chat, img, cpt)
			time.Sleep(300 * time.Millisecond)
		}
		return nil, nil
	}

	parentKey := &waCommon.MessageKey{
		RemoteJID: proto.String(chat.String()),
		FromMe:    proto.Bool(true),
		ID:        proto.String(string(parentResp.ID)),
	}

	var sentIDs []protocol.MessageID
	assocType := waE2E.MessageAssociation_MEDIA_ALBUM

	// 3. Kirim masing-masing foto yang dikaitkan ke Parent Album via MessageAssociation
	for i, imgBytes := range images {
		uploaded, err := media.UploadMedia(c, ctx, imgBytes, whatsmeow.MediaImage)
		if err != nil {
			continue
		}

		caption := ""
		if i < len(captions) {
			caption = captions[i]
		}

		childMsg := &waE2E.Message{
			MessageContextInfo: &waE2E.MessageContextInfo{
				MessageAssociation: &waE2E.MessageAssociation{
					AssociationType:  &assocType,
					ParentMessageKey: parentKey,
					MessageIndex:     proto.Int32(int32(i)),
				},
			},
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

		resp, err := c.Client.SendMessage(ctx, chat, childMsg)
		if err == nil {
			sentIDs = append(sentIDs, resp.ID)
		}
		time.Sleep(250 * time.Millisecond)
	}

	return sentIDs, nil
}
