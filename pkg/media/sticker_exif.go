package media

import (
	"fmt"
	"os"
	"os/exec"
)

// FixWebPRIFFHeader memperbaiki ukuran RIFF header file WebP jika dihasilkan dari ffmpeg pipe (yang menulis RIFF size 0)
func FixWebPRIFFHeader(data []byte) []byte {
	if len(data) >= 12 && string(data[0:4]) == "RIFF" && string(data[8:12]) == "WEBP" {
		size := uint32(len(data) - 8)
		data[4] = byte(size)
		data[5] = byte(size >> 8)
		data[6] = byte(size >> 16)
		data[7] = byte(size >> 24)
	}
	return data
}

// AddStickerMetadata menyuntikkan metadata EXIF resmi (PackName, Publisher, Emojis) ke dalam binary file WebP menggunakan webpmux
func AddStickerMetadata(webpBytes []byte, packName, publisher string) ([]byte, error) {
	if len(webpBytes) == 0 {
		return nil, fmt.Errorf("webp bytes kosong")
	}

	webpBytes = FixWebPRIFFHeader(webpBytes)

	meta := StickerMetadata{
		PackName:  packName,
		Publisher: publisher,
	}
	exifChunk, err := BuildExifChunk(meta)
	if err != nil {
		return webpBytes, err
	}

	tmpWebp, err := os.CreateTemp("", "webp-*.webp")
	if err != nil {
		return webpBytes, nil
	}
	tmpExif, err := os.CreateTemp("", "exif-*.exif")
	if err != nil {
		_ = os.Remove(tmpWebp.Name())
		return webpBytes, nil
	}

	defer os.Remove(tmpWebp.Name())
	defer os.Remove(tmpExif.Name())

	_, _ = tmpWebp.Write(webpBytes)
	_, _ = tmpExif.Write(exifChunk)
	_ = tmpWebp.Close()
	_ = tmpExif.Close()

	outPath := tmpWebp.Name() + ".out.webp"
	defer os.Remove(outPath)

	cmd := exec.Command("webpmux", "-set", "exif", tmpExif.Name(), tmpWebp.Name(), "-o", outPath)
	if err := cmd.Run(); err != nil {
		// Fallback jika webpmux gagal
		return webpBytes, nil
	}

	muxedBytes, err := os.ReadFile(outPath)
	if err != nil || len(muxedBytes) == 0 {
		return webpBytes, nil
	}

	return muxedBytes, nil
}
