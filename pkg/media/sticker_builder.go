package media

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
)

// StickerMetadata menyimpan paket metadata untuk stiker WhatsApp
type StickerMetadata struct {
	PackName  string   `json:"sticker-pack-name"`
	Publisher string   `json:"sticker-pack-publisher"`
	Emojis    []string `json:"emojis,omitempty"`
	StickerID string   `json:"sticker-pack-id,omitempty"`
	IsAvatar  bool     `json:"is-avatar-sticker,omitempty"`
}

// BuildExifChunk membuat chunk EXIF WebP untuk menyematkan PackName & Publisher pada stiker
func BuildExifChunk(metadata StickerMetadata) ([]byte, error) {
	if metadata.PackName == "" {
		metadata.PackName = "Dongtube Sticker"
	}
	if metadata.Publisher == "" {
		metadata.Publisher = "Dongtube Bot"
	}
	if len(metadata.Emojis) == 0 {
		metadata.Emojis = []string{"🐱", "✨"}
	}

	jsonBytes, err := json.Marshal(map[string]interface{}{
		"sticker-pack-id":        metadata.StickerID,
		"sticker-pack-name":      metadata.PackName,
		"sticker-pack-publisher": metadata.Publisher,
		"emojis":                 metadata.Emojis,
		"is-avatar-sticker":      metadata.IsAvatar,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal sticker metadata failed: %w", err)
	}

	header := []byte{
		0x49, 0x49, 0x2A, 0x00, 0x08, 0x00, 0x00, 0x00,
		0x01, 0x00, 0x41, 0x57, 0x07, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x16, 0x00, 0x00, 0x00,
	}

	binary.LittleEndian.PutUint32(header[14:18], uint32(len(jsonBytes)))

	var buf bytes.Buffer
	buf.Write([]byte("Exif\x00\x00"))
	buf.Write(header)
	buf.Write(jsonBytes)

	return buf.Bytes(), nil
}
