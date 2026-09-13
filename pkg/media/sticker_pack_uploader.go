package media

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	protocol "github.com/hanzywaifu-coder/dongtube-meowbail/internal/protocol"
	"io"
	"net/http"
	"net/url"

	"go.mau.fi/util/random"
	"go.mau.fi/whatsmeow/socket"
	"go.mau.fi/whatsmeow/util/cbcutil"
	"go.mau.fi/whatsmeow/util/hkdfutil"
)

type uploadedPackItem struct {
	DirectPath    string
	FileSHA256    []byte
	FileEncSHA256 []byte
	MediaKey      []byte
	FileLength    uint64
}

// uploadRawMediaToMMS mengunggah payload langsung ke endpoint MMS WhatsApp dengan HKDF dan Path kustom
func uploadRawMediaToMMS(c *protocol.Client, ctx context.Context, plaintext []byte, appInfo string, mmsPath string, mediaKey []byte) (*uploadedPackItem, error) {
	if len(mediaKey) == 0 {
		mediaKey = random.Bytes(32)
	}

	mediaKeyExpanded := hkdfutil.SHA256(mediaKey, nil, []byte(appInfo), 112)
	iv := mediaKeyExpanded[:16]
	cipherKey := mediaKeyExpanded[16:48]
	macKey := mediaKeyExpanded[48:80]

	ciphertext, err := cbcutil.Encrypt(cipherKey, iv, plaintext)
	if err != nil {
		return nil, fmt.Errorf("cbc encrypt failed: %w", err)
	}

	h := hmac.New(sha256.New, macKey)
	h.Write(iv)
	h.Write(ciphertext)
	mac := h.Sum(nil)[:10]

	encBuffer := append(ciphertext, mac...)
	encSha256 := sha256.Sum256(encBuffer)
	plainSha256 := sha256.Sum256(plaintext)

	mediaConn, err := c.Client.DangerousInternals().RefreshMediaConn(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("refresh media conn failed: %w", err)
	}

	token := base64.RawURLEncoding.EncodeToString(encSha256[:])
	uploadQuery := url.Values{
		"auth":  []string{mediaConn.Auth},
		"token": []string{token},
	}

	var directPath string
	var lastErr error

	for _, host := range mediaConn.Hosts {
		uploadURL := url.URL{
			Scheme:   "https",
			Host:     host.Hostname,
			Path:     fmt.Sprintf("%s/%s", mmsPath, token),
			RawQuery: uploadQuery.Encode(),
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL.String(), bytes.NewReader(encBuffer))
		if err != nil {
			lastErr = err
			continue
		}

		req.Header.Set("Origin", socket.Origin)
		req.Header.Set("Referer", socket.Origin+"/")
		req.Header.Set("Content-Type", "application/octet-stream")
		req.Header.Set("Content-Length", fmt.Sprintf("%d", len(encBuffer)))

		httpResp, err := http.DefaultClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		bodyBytes, _ := io.ReadAll(httpResp.Body)
		httpResp.Body.Close()

		if httpResp.StatusCode >= 200 && httpResp.StatusCode < 300 {
			var respData struct {
				URL        string `json:"url"`
				DirectPath string `json:"direct_path"`
			}
			if err := json.Unmarshal(bodyBytes, &respData); err == nil {
				directPath = respData.DirectPath
				if directPath == "" {
					directPath = respData.URL
				}
				break
			}
		}
		lastErr = fmt.Errorf("upload to host %s failed with status %d: %s", host.Hostname, httpResp.StatusCode, string(bodyBytes))
	}

	if directPath == "" {
		return nil, fmt.Errorf("all upload hosts failed: %v", lastErr)
	}

	return &uploadedPackItem{
		DirectPath:    directPath,
		FileSHA256:    plainSha256[:],
		FileEncSHA256: encSha256[:],
		MediaKey:      mediaKey,
		FileLength:    uint64(len(plaintext)),
	}, nil
}
