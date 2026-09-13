package messaging

import (
	"bytes"
	"context"
	protocol "github.com/hanzywaifu-coder/dongtube-meowbail/internal/protocol"
	"io"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"time"

	media "github.com/hanzywaifu-coder/dongtube-meowbail/pkg/media"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

var urlRegex = regexp.MustCompile(`(?i)\bhttps?://[^\s<>"{}|\^~\[\]` + "`" + `]+`)

// LinkPreviewInfo menampung metadata hasil parsing OpenGraph / Web
type LinkPreviewInfo struct {
	MatchedURL   string
	CanonicalURL string
	Title        string
	Description  string
	ImageURL     string
	JPEGThumb    []byte
}

// ExtractFirstURL mencari URL pertama dari pesan teks
func ExtractFirstURL(text string) string {
	return urlRegex.FindString(text)
}

// FetchLinkPreview mengekstrak open-graph title, description, dan thumbnail dari web URL
func FetchLinkPreview(targetURL string, timeout ...time.Duration) (*LinkPreviewInfo, error) {
	t := 5 * time.Second
	if len(timeout) > 0 {
		t = timeout[0]
	}

	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "WhatsApp/2.24.12 (compatible; OpenGraphScraper)")

	client := &http.Client{Timeout: t}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 1024*512)) // Max 512KB
	if err != nil {
		return nil, err
	}
	html := string(bodyBytes)

	info := &LinkPreviewInfo{
		MatchedURL:   targetURL,
		CanonicalURL: targetURL,
	}

	// Regex pencarian OpenGraph Meta
	reTitle := regexp.MustCompile(`(?i)<meta\s+property=["']og:title["']\s+content=["']([^"']+)["']`)
	if m := reTitle.FindStringSubmatch(html); len(m) > 1 {
		info.Title = m[1]
	} else {
		reHTMLTitle := regexp.MustCompile(`(?i)<title>([^<]+)</title>`)
		if mTitle := reHTMLTitle.FindStringSubmatch(html); len(mTitle) > 1 {
			info.Title = strings.TrimSpace(mTitle[1])
		}
	}

	reDesc := regexp.MustCompile(`(?i)<meta\s+property=["']og:description["']\s+content=["']([^"']+)["']`)
	if m := reDesc.FindStringSubmatch(html); len(m) > 1 {
		info.Description = m[1]
	} else {
		reMetaDesc := regexp.MustCompile(`(?i)<meta\s+name=["']description["']\s+content=["']([^"']+)["']`)
		if mDesc := reMetaDesc.FindStringSubmatch(html); len(mDesc) > 1 {
			info.Description = strings.TrimSpace(mDesc[1])
		}
	}

	reImg := regexp.MustCompile(`(?i)<meta\s+property=["']og:image["']\s+content=["']([^"']+)["']`)
	if m := reImg.FindStringSubmatch(html); len(m) > 1 {
		info.ImageURL = m[1]
	}

	// Download & resize image thumbnail jika tersedia
	if info.ImageURL != "" {
		imgReq, err := http.NewRequest("GET", info.ImageURL, nil)
		if err == nil {
			imgResp, err := client.Do(imgReq)
			if err == nil && imgResp.StatusCode == 200 {
				defer imgResp.Body.Close()
				rawImg, err := io.ReadAll(io.LimitReader(imgResp.Body, 2*1024*1024))
				if err == nil && len(rawImg) > 0 {
					cmd := exec.Command("ffmpeg", "-i", "pipe:0", "-vf", "scale=192:192:force_original_aspect_ratio=decrease,pad=192:192:(192-iw)/2:(192-ih)/2:color=0x00000000", "-vcodec", "mjpeg", "-f", "image2", "pipe:1")
					cmd.Stdin = bytes.NewReader(rawImg)
					thumb, errThumb := cmd.Output()
					if errThumb == nil && len(thumb) > 0 {
						info.JPEGThumb = thumb
					}

				}
			}
		}
	}

	return info, nil
}

// SendTextMessageWithPreview mengirim pesan teks dengan link preview kaya (judul, deskripsi, & thumbnail kartu)
func SendTextMessageWithPreview(c *protocol.Client, ctx context.Context, chat protocol.JID, text string, preview *LinkPreviewInfo) error {
	if preview == nil {
		matchedURL := ExtractFirstURL(text)
		if matchedURL != "" {
			fetched, err := FetchLinkPreview(matchedURL)
			if err == nil {
				preview = fetched
			}
		}
	}

	extMsg := &waE2E.ExtendedTextMessage{
		Text: proto.String(text),
	}

	if preview != nil && (preview.Title != "" || len(preview.JPEGThumb) > 0) {
		extMsg.MatchedText = proto.String(preview.MatchedURL)
		if preview.Title != "" {
			extMsg.Title = proto.String(preview.Title)
		}
		if preview.Description != "" {
			extMsg.Description = proto.String(preview.Description)
		}
		if len(preview.JPEGThumb) > 0 {
			extMsg.JPEGThumbnail = preview.JPEGThumb

			// Mengunggah thumbnail link beresolusi tinggi (HighQualityThumbnail)
			uploaded, err := media.UploadMedia(c, ctx, preview.JPEGThumb, whatsmeow.MediaLinkThumbnail)
			if err == nil && uploaded != nil {
				extMsg.ThumbnailDirectPath = proto.String(uploaded.DirectPath)
				extMsg.ThumbnailSHA256 = uploaded.FileSHA256
				extMsg.ThumbnailEncSHA256 = uploaded.FileEncSHA256
				extMsg.MediaKey = uploaded.MediaKey
				extMsg.MediaKeyTimestamp = proto.Int64(time.Now().Unix())
				extMsg.ThumbnailWidth = proto.Uint32(192)
				extMsg.ThumbnailHeight = proto.Uint32(192)
			}
		}
	}

	msg := &waE2E.Message{
		ExtendedTextMessage: extMsg,
	}
	if extMsg.ContextInfo == nil {
		extMsg.ContextInfo = BuildNewsletterContext(c.Config())
	}

	_, err := c.Client.SendMessage(ctx, chat, msg)
	return err
}
