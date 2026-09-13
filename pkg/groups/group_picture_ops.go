package groups

import (
	"bytes"
	"context"
	"fmt"
	protocol "github.com/hanzywaifu-coder/dongtube-meowbail/internal/protocol"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"os/exec"
	"strings"
)

// SetGroupProfilePicture mengonversi, menormalkan, dan mengatur foto profil grup (avatar)
// Mengikuti standar persis Baileys/WhatsApp:
// - Dimensi 640x640 center crop
// - Baseline JPEG murni (strip seluruh marker APP0-APP15 dan COM)
// - Stanza IQ xmlns="w:profile:picture" target=groupJID
func SetGroupProfilePicture(c *protocol.Client, ctx context.Context, groupJID protocol.JID, rawImageBytes []byte) (string, error) {
	if len(rawImageBytes) == 0 {
		return c.Client.SetGroupPhoto(ctx, groupJID, nil)
	}

	// 1. Cek apakah bot adalah admin
	isBotAdmin, errCheck := IsBotAdmin(c, ctx, groupJID)
	if errCheck == nil && !isBotAdmin {
		return "", fmt.Errorf("bot bukan admin di grup ini, tidak dapat mengubah foto profil grup")
	}

	// 2. Normalisasi gambar menjadi baseline JPEG 640x640 murni identik dengan Sharp (Baileys standard)
	// Jalankan Sharp processor script
	var jpegBytes []byte
	sharpCmd := exec.Command("node", "-e", `
const sharp = require('/root/sc/node_modules/sharp');
const chunks = [];
process.stdin.on('data', c => chunks.push(c));
process.stdin.on('end', async () => {
  try {
    const input = Buffer.concat(chunks);
    const out = await sharp(input)
      .resize(640, 640, { fit: 'cover' })
      .jpeg({ quality: 50, progressive: false })
      .toBuffer();
    process.stdout.write(out);
  } catch (err) {
    process.stderr.write(err.message);
    process.exit(1);
  }
});
`)
	sharpCmd.Stdin = bytes.NewReader(rawImageBytes)
	sharpOut, sharpErr := sharpCmd.Output()
	if sharpErr == nil && len(sharpOut) > 0 {
		jpegBytes = sharpOut
	} else {
		// Fallback ke ffmpeg jika Sharp tidak tersedia
		cmd := exec.Command(
			"ffmpeg",
			"-y",
			"-i", "pipe:0",
			"-vf", "scale=640:640:force_original_aspect_ratio=increase,crop=640:640",
			"-pix_fmt", "yuv420p",
			"-f", "image2",
			"pipe:1",
		)
		cmd.Stdin = bytes.NewReader(rawImageBytes)
		jpegBytes, _ = cmd.Output()
	}

	// 3. Fallback jika pipeline eksternal gagal: gunakan image decoding Go standard library (quality 50)
	if len(jpegBytes) == 0 {
		img, _, decErr := image.Decode(bytes.NewReader(rawImageBytes))
		if decErr != nil {
			return "", fmt.Errorf("format gambar tidak dikenali: %w", decErr)
		}

		b := img.Bounds()
		w := b.Dx()
		h := b.Dy()
		minDim := w
		if h < minDim {
			minDim = h
		}

		startX := b.Min.X + (w-minDim)/2
		startY := b.Min.Y + (h-minDim)/2

		dst := image.NewRGBA(image.Rect(0, 0, 640, 640))
		for y := 0; y < 640; y++ {
			for x := 0; x < 640; x++ {
				srcX := startX + (x*minDim)/640
				srcY := startY + (y*minDim)/640
				dst.Set(x, y, img.At(srcX, srcY))
			}
		}

		var buf bytes.Buffer
		encErr := jpeg.Encode(&buf, dst, &jpeg.Options{Quality: 50})
		if encErr != nil {
			return "", fmt.Errorf("gagal encode jpeg: %w", encErr)
		}
		jpegBytes = buf.Bytes()
	}

	// 4. Bersihkan seluruh marker APP0-APP15 (JFIF, EXIF, dsb) dan komentar (COM)
	cleanBytes := StripJPEGMetadata(jpegBytes)

	// Kirim via standard whatsmeow SetGroupPhoto
	picID, err := c.Client.SetGroupPhoto(ctx, groupJID, cleanBytes)
	if err == nil && picID != "" {
		return picID, nil
	}

	if err != nil {
		if strings.Contains(err.Error(), "not-acceptable") || strings.Contains(err.Error(), "406") {
			return "", fmt.Errorf("server WhatsApp menolak (406 not-acceptable). Bot harus dijadikan Admin di grup ini untuk mengubah foto profil grup")
		}
	}

	return "", err
}

// SetOwnProfilePicture mengubah foto profil (avatar) akun WhatsApp bot itu sendiri (Personal / Account PP)
// Parity dengan Baileys updateProfilePicture(sock.user.id)
func SetOwnProfilePicture(c *protocol.Client, ctx context.Context, rawImageBytes []byte) (string, error) {
	if c.Client.Store == nil || c.Client.Store.ID == nil {
		return "", fmt.Errorf("koneksi bot belum terautentikasi")
	}

	ownJID := c.Client.Store.ID.ToNonAD()
	if len(rawImageBytes) == 0 {
		return c.Client.SetGroupPhoto(ctx, ownJID, nil)
	}

	var jpegBytes []byte
	sharpCmd := exec.Command("node", "-e", `
const sharp = require('/root/sc/node_modules/sharp');
const chunks = [];
process.stdin.on('data', c => chunks.push(c));
process.stdin.on('end', async () => {
  try {
    const input = Buffer.concat(chunks);
    const out = await sharp(input)
      .resize(640, 640, { fit: 'cover' })
      .jpeg({ quality: 50, progressive: false })
      .toBuffer();
    process.stdout.write(out);
  } catch (err) {
    process.stderr.write(err.message);
    process.exit(1);
  }
});
`)
	sharpCmd.Stdin = bytes.NewReader(rawImageBytes)
	sharpOut, sharpErr := sharpCmd.Output()
	if sharpErr == nil && len(sharpOut) > 0 {
		jpegBytes = sharpOut
	} else {
		cmd := exec.Command(
			"ffmpeg",
			"-i", "pipe:0",
			"-vf", "scale=640:640:force_original_aspect_ratio=increase,crop=640:640",
			"-f", "image2",
			"-vcodec", "mjpeg",
			"-q:v", "5",
			"pipe:1",
		)
		cmd.Stdin = bytes.NewReader(rawImageBytes)
		out, err := cmd.Output()
		if err == nil && len(out) > 0 {
			jpegBytes = out
		}
	}

	if len(jpegBytes) == 0 {
		src, _, err := image.Decode(bytes.NewReader(rawImageBytes))
		if err != nil {
			return "", fmt.Errorf("gagal decode gambar: %w", err)
		}
		b := src.Bounds()
		sz := min(b.Dx(), b.Dy())
		x0 := b.Min.X + (b.Dx()-sz)/2
		y0 := b.Min.Y + (b.Dy()-sz)/2
		dst := image.NewRGBA(image.Rect(0, 0, sz, sz))
		for y := 0; y < sz; y++ {
			for x := 0; x < sz; x++ {
				dst.Set(x, y, src.At(x0+x, y0+y))
			}
		}
		var buf bytes.Buffer
		encErr := jpeg.Encode(&buf, dst, &jpeg.Options{Quality: 50})
		if encErr != nil {
			return "", fmt.Errorf("gagal encode jpeg: %w", encErr)
		}
		jpegBytes = buf.Bytes()
	}

	cleanBytes := StripJPEGMetadata(jpegBytes)
	return c.Client.SetGroupPhoto(ctx, ownJID, cleanBytes)
}

// RemoveOwnProfilePicture menghapus foto profil akun WhatsApp bot itu sendiri
// Parity dengan Baileys removeProfilePicture(sock.user.id)
func RemoveOwnProfilePicture(c *protocol.Client, ctx context.Context) error {
	if c.Client.Store == nil || c.Client.Store.ID == nil {
		return fmt.Errorf("koneksi bot belum terautentikasi")
	}

	ownJID := c.Client.Store.ID.ToNonAD()
	_, err := c.Client.SetGroupPhoto(ctx, ownJID, nil)
	return err
}

// StripJPEGMetadata menghapus marker APP0..APP15 dan COM dari stream JPEG,
// menghasilkan pure baseline stream yang identik dengan output library Sharp (Baileys standard).
func StripJPEGMetadata(data []byte) []byte {
	if len(data) < 4 || data[0] != 0xff || data[1] != 0xd8 {
		return data
	}

	res := make([]byte, 0, len(data))
	res = append(res, 0xff, 0xd8)
	i := 2
	for i < len(data) {
		if data[i] == 0xff && i+1 < len(data) {
			marker := data[i+1]
			// Marker tanpa payload length
			if marker == 0xd8 || marker == 0x00 || (marker >= 0xd0 && marker <= 0xd7) {
				res = append(res, data[i], marker)
				i += 2
				continue
			}
			// End of Image (EOI) atau Start of Scan (SOS): scan data dimulai, salin sisa payload
			if marker == 0xd9 || marker == 0xda {
				res = append(res, data[i:]...)
				break
			}

			if i+3 >= len(data) {
				res = append(res, data[i:]...)
				break
			}
			length := (int(data[i+2]) << 8) | int(data[i+3])
			if i+2+length > len(data) {
				res = append(res, data[i:]...)
				break
			}

			// Buang APP0-APP15 (0xe0 - 0xef) dan Comment (0xfe)
			if (marker >= 0xe0 && marker <= 0xef) || marker == 0xfe {
				i += 2 + length
			} else {
				res = append(res, data[i:i+2+length]...)
				i += 2 + length
			}
		} else {
			res = append(res, data[i])
			i++
		}
	}
	return res
}
