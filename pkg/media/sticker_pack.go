package media

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	protocol "github.com/hanzywaifu-coder/dongtube-meowbail/internal/protocol"
	"os"
	"os/exec"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

// SendStickerPackMultipleMedia mengemas kumpulan media stiker ke dalam zip format WhatsApp sticker pack resmi dan mengirim StickerPackMessage
func SendStickerPackMultipleMedia(c *protocol.Client, ctx context.Context, chat protocol.JID, stickerItems [][]byte, packName, publisher string) error {
	if len(stickerItems) == 0 {
		return fmt.Errorf("daftar stiker kosong")
	}

	if packName == "" {
		packName = "Dongtube Pack"
	}
	if publisher == "" {
		publisher = "Dongtube"
	}

	packHash := sha256.Sum256([]byte(fmt.Sprintf("%s_%d", packName, time.Now().UnixNano())))
	packID := fmt.Sprintf("Pack_%x", packHash[:8])
	trayIconFileName := "tray_icon.webp"

	// Build ZIP container murni (metode Store, Flags = 0 tanpa bit data descriptor)
	// sesuai spesifikasi parser C++ WhatsApp Android & iOS
	zipWriter := NewWhatsAppZipWriter()

	var stickers []*waE2E.StickerPackMessage_Sticker

	for _, rawItem := range stickerItems {
		// Perbaiki ukuran header RIFF jika berasal dari pipe streaming
		rawItem = FixWebPRIFFHeader(rawItem)

		// Tambahkan metadata EXIF resmi jika belum ada
		itemData, errMeta := AddStickerMetadata(rawItem, packName, publisher)
		if errMeta != nil || len(itemData) == 0 {
			itemData = rawItem
		}
		itemData = FixWebPRIFFHeader(itemData)

		// Deteksi apakah stiker WebP animasi (memiliki chunk ANIM / ANMF)
		isAnim := bytes.Contains(itemData, []byte("ANIM")) || bytes.Contains(itemData, []byte("ANMF"))

		h := sha256.Sum256(itemData)
		// Format Base64URL tanpa padding (RawURLEncoding) 43 karakter sesuai WhatsApp / Baileys
		b64Clean := base64.RawURLEncoding.EncodeToString(h[:])
		fileName := b64Clean + ".webp"

		zipWriter.AddFile(fileName, itemData)

		stickers = append(stickers, &waE2E.StickerPackMessage_Sticker{
			Emojis:             []string{""},
			FileName:           proto.String(fileName),
			IsAnimated:         proto.Bool(isAnim),
			AccessibilityLabel: proto.String(packName),
			IsLottie:           proto.Bool(false),
			Mimetype:           proto.String("image/webp"),
		})
	}

	// Ekstrak frame 1 jika stiker pertama adalah WebP animasi (agar ffmpeg tidak error 'skipping unsupported chunk: ANIM')
	coverSample := FixWebPRIFFHeader(stickerItems[0])
	if bytes.Contains(coverSample, []byte("ANIM")) || bytes.Contains(coverSample, []byte("ANMF")) {
		tmpIn, errIn := os.CreateTemp("", "anim-cover-*.webp")
		if errIn == nil {
			defer os.Remove(tmpIn.Name())
			_, _ = tmpIn.Write(coverSample)
			_ = tmpIn.Close()
			frameOut := tmpIn.Name() + ".frame1.webp"
			defer os.Remove(frameOut)
			cmdExtract := exec.Command("webpmux", "-get", "frame", "1", tmpIn.Name(), "-o", frameOut)
			if errExt := cmdExtract.Run(); errExt == nil {
				if fb, errRead := os.ReadFile(frameOut); errRead == nil && len(fb) > 0 {
					coverSample = FixWebPRIFFHeader(fb)
				}
			}
		}
	}

	// Buat tray_icon.webp 96x96 sesuai ukuran standar WhatsApp tray icon
	cmdTray := exec.Command("ffmpeg", "-i", "pipe:0", "-vf", "scale=96:96:force_original_aspect_ratio=decrease,pad=96:96:(96-iw)/2:(96-ih)/2:color=0x00000000", "-vcodec", "libwebp", "-f", "webp", "pipe:1")
	cmdTray.Stdin = bytes.NewReader(coverSample)
	trayBytes, err := cmdTray.Output()
	if err != nil || len(trayBytes) == 0 {
		trayBytes = coverSample
	}
	trayBytes = FixWebPRIFFHeader(trayBytes)

	// Masukkan tray cover ke zip dengan nama baku WhatsApp: tray_icon.webp
	zipWriter.AddFile(trayIconFileName, trayBytes)

	zipBytes := zipWriter.Bytes()

	// 1. Upload ZIP stiker pack ke WhatsApp MMS khusus Sticker Pack
	uploadedZip, err := uploadRawMediaToMMS(c, ctx, zipBytes, "WhatsApp Sticker Pack Keys", "/mms/sticker-pack", nil)
	if err != nil {
		// Fallback ke uploader umum jika host khusus gagal
		resp, errUp := UploadMedia(c, ctx, zipBytes, whatsmeow.MediaStickerPack)
		if errUp != nil {
			return fmt.Errorf("upload zip sticker pack: %w", errUp)
		}
		uploadedZip = &uploadedPackItem{
			DirectPath:    resp.DirectPath,
			FileSHA256:    resp.FileSHA256,
			FileEncSHA256: resp.FileEncSHA256,
			MediaKey:      resp.MediaKey,
			FileLength:    resp.FileLength,
		}
	}

	// Generate 252x252 JPEG thumbnail untuk tray icon
	cmdThumb := exec.Command("ffmpeg", "-i", "pipe:0", "-vf", "scale=252:252:force_original_aspect_ratio=decrease,pad=252:252:(252-iw)/2:(252-ih)/2:color=white", "-vcodec", "mjpeg", "-f", "image2", "pipe:1")
	cmdThumb.Stdin = bytes.NewReader(coverSample)
	thumbJpeg, err := cmdThumb.Output()
	if err != nil || len(thumbJpeg) == 0 {
		cmdThumbFallback := exec.Command("ffmpeg", "-i", "pipe:0", "-vf", "scale=252:252", "-vcodec", "mjpeg", "-f", "image2", "pipe:1")
		cmdThumbFallback.Stdin = bytes.NewReader(coverSample)
		thumbJpeg, _ = cmdThumbFallback.Output()
	}

	var thumbDirectPath string
	var thumbSha256 []byte
	var thumbEncSha256 []byte
	var imageDataHash string

	// 2. Upload thumbnail JPEG ke MMS khusus Thumbnail Sticker Pack dengan MediaKey yang SAMA dengan ZIP pack
	if len(thumbJpeg) > 0 {
		uploadedThumb, errThumb := uploadRawMediaToMMS(c, ctx, thumbJpeg, "WhatsApp Sticker Pack Thumbnail Keys", "/mms/thumbnail-sticker-pack", uploadedZip.MediaKey)
		if errThumb == nil && uploadedThumb != nil {
			thumbDirectPath = uploadedThumb.DirectPath
			thumbSha256 = uploadedThumb.FileSHA256
			thumbEncSha256 = uploadedThumb.FileEncSHA256
			imageDataHash = base64.StdEncoding.EncodeToString(thumbSha256)
		}
	}

	if thumbDirectPath == "" {
		thumbDirectPath = uploadedZip.DirectPath
		thumbSha256 = uploadedZip.FileSHA256
		thumbEncSha256 = uploadedZip.FileEncSHA256
		if imageDataHash == "" {
			imageDataHash = base64.StdEncoding.EncodeToString(thumbSha256)
		}
	}

	origin := waE2E.StickerPackMessage_USER_CREATED

	msg := &waE2E.Message{
		StickerPackMessage: &waE2E.StickerPackMessage{
			Stickers:            stickers,
			StickerPackID:       proto.String(packID),
			Name:                proto.String(packName),
			Publisher:           proto.String(publisher),
			FileLength:          proto.Uint64(uploadedZip.FileLength),
			FileSHA256:          uploadedZip.FileSHA256,
			FileEncSHA256:       uploadedZip.FileEncSHA256,
			MediaKey:            uploadedZip.MediaKey,
			DirectPath:          proto.String(uploadedZip.DirectPath),
			PackDescription:     proto.String("Sticker pack dari album"),
			MediaKeyTimestamp:   proto.Int64(time.Now().Unix()),
			TrayIconFileName:    proto.String(trayIconFileName),
			ThumbnailDirectPath: proto.String(thumbDirectPath),
			ThumbnailSHA256:     thumbSha256,
			ThumbnailEncSHA256:  thumbEncSha256,
			ThumbnailHeight:     proto.Uint32(252),
			ThumbnailWidth:      proto.Uint32(252),
			ImageDataHash:       proto.String(imageDataHash),
			StickerPackSize:     proto.Uint64(uint64(len(zipBytes))),
			StickerPackOrigin:   &origin,
		},
	}

	_, err = c.Client.SendMessage(ctx, chat, msg)
	return err
}

// SendStickerPackFromMedia wrapper backward-compatibility untuk 1 stiker
func SendStickerPackFromMedia(c *protocol.Client, ctx context.Context, chat protocol.JID, stickerWebpData []byte, packName, publisher string) error {
	return SendStickerPackMultipleMedia(c, ctx, chat, [][]byte{stickerWebpData}, packName, publisher)
}
