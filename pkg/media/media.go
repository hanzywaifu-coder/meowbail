package media

import (
	"context"
	protocol "github.com/hanzywaifu-coder/dongtube-meowbail/internal/protocol"
	"io"
	"net/http"
	"os"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

// DownloadMedia downloads media from a DownloadableMessage
func DownloadMedia(c *protocol.Client, ctx context.Context, msg whatsmeow.DownloadableMessage) ([]byte, error) {
	if msg == nil {
		return nil, io.ErrNoProgress
	}
	return c.Client.Download(ctx, msg)
}

// UploadMedia uploads media and returns upload info with automatic smart LRU deduplication cache
func UploadMedia(c *protocol.Client, ctx context.Context, data []byte, mediaType whatsmeow.MediaType) (*whatsmeow.UploadResponse, error) {
	cache := GetGlobalUploadCache()
	if cachedResp, hit := cache.Get(mediaType, data); hit {
		return &cachedResp, nil
	}

	resp, err := c.Client.Upload(ctx, data, mediaType)
	if err != nil {
		return nil, err
	}

	cache.Put(mediaType, data, resp)
	return &resp, nil
}

var pooledHTTPClient = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
	},
}

// FetchURL fetches data from a URL with timeout reusing pooled transport
func FetchURL(url string, timeout ...time.Duration) ([]byte, error) {
	t := 30 * time.Second
	if len(timeout) > 0 {
		t = timeout[0]
	}

	ctx, cancel := context.WithTimeout(context.Background(), t)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pooledHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// ReadFile reads a file and returns its contents
func ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// SaveFile saves data to a file
func SaveFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}

// BuildImageMessage builds an image message with newsletter context
func BuildImageMessage(resp *whatsmeow.UploadResponse, caption string, cfg *protocol.Config) *waE2E.Message {
	return &waE2E.Message{
		ImageMessage: &waE2E.ImageMessage{
			URL:           &resp.URL,
			Mimetype:      proto.String("image/jpeg"),
			Caption:       proto.String(caption),
			FileEncSHA256: resp.FileEncSHA256,
			FileSHA256:    resp.FileSHA256,
			FileLength:    &resp.FileLength,
			DirectPath:    &resp.DirectPath,
			MediaKey:      resp.MediaKey,
			ContextInfo:   newsletterContext(cfg),
		},
	}
}

// BuildVideoMessage builds a video message with newsletter context
func BuildVideoMessage(resp *whatsmeow.UploadResponse, caption string, cfg *protocol.Config) *waE2E.Message {
	return &waE2E.Message{
		VideoMessage: &waE2E.VideoMessage{
			URL:           &resp.URL,
			Mimetype:      proto.String("video/mp4"),
			Caption:       proto.String(caption),
			FileEncSHA256: resp.FileEncSHA256,
			FileSHA256:    resp.FileSHA256,
			FileLength:    &resp.FileLength,
			DirectPath:    &resp.DirectPath,
			MediaKey:      resp.MediaKey,
			ContextInfo:   newsletterContext(cfg),
		},
	}
}

// BuildDocumentMessage builds a document message with newsletter context
func BuildDocumentMessage(resp *whatsmeow.UploadResponse, filename, mimetype, caption string, cfg *protocol.Config) *waE2E.Message {
	return &waE2E.Message{
		DocumentMessage: &waE2E.DocumentMessage{
			URL:           &resp.URL,
			Mimetype:      proto.String(mimetype),
			FileName:      proto.String(filename),
			Caption:       proto.String(caption),
			FileEncSHA256: resp.FileEncSHA256,
			FileSHA256:    resp.FileSHA256,
			FileLength:    &resp.FileLength,
			DirectPath:    &resp.DirectPath,
			MediaKey:      resp.MediaKey,
			ContextInfo:   newsletterContext(cfg),
		},
	}
}

func newsletterContext(cfg *protocol.Config) *waE2E.ContextInfo {
	if cfg != nil && cfg.DefaultFakeReply != nil {
		return proto.Clone(cfg.DefaultFakeReply).(*waE2E.ContextInfo)
	}

	if cfg == nil || cfg.NewsletterJID == "" {
		return &waE2E.ContextInfo{}
	}

	return &waE2E.ContextInfo{
		IsForwarded:     proto.Bool(false),
		ForwardingScore: proto.Uint32(0),
		BusinessMessageForwardInfo: &waE2E.ContextInfo_BusinessMessageForwardInfo{
			BusinessOwnerJID: proto.String(cfg.BusinessOwnerJID),
		},
		ForwardedNewsletterMessageInfo: &waE2E.ContextInfo_ForwardedNewsletterMessageInfo{
			NewsletterJID:  proto.String(cfg.NewsletterJID),
			NewsletterName: proto.String(cfg.NewsletterName),
		},
	}
}
