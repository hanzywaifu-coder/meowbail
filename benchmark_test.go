package meowbail_test

import (
	"testing"
	"time"

	meowbail "github.com/hanzywaifu-coder/dongtube-meowbail"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

func BenchmarkRateLimiterWaitOrThrottle(b *testing.B) {
	limiter := meowbail.NewRateLimiter(100*time.Millisecond, 100)
	jid := types.NewJID("628123456789", types.DefaultUserServer)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = limiter.WaitOrThrottle(jid)
	}
}

func BenchmarkExtractPureText(b *testing.B) {
	msg := &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: proto.String(".remini https://cdn.dongtube.id/sample.jpg"),
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = meowbail.ExtractPureText(msg)
	}
}

func BenchmarkExtractMentions(b *testing.B) {
	text := "Halo @628111111111 dan @628222222222 serta @628333333333"

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = meowbail.ExtractMentions(text)
	}
}

func BenchmarkBufferPool(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		buf := meowbail.GetSmallBuffer()
		buf.WriteString("Benchmark Dongtube MeowBail Reusable Buffer")
		meowbail.PutSmallBuffer(buf)
	}
}

func BenchmarkMessageDeduplicator(b *testing.B) {
	dedup := meowbail.NewMessageDeduplicator(1000, 5*time.Minute)
	id := types.MessageID("3EB0123456789ABCDEF0")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = dedup.CheckAndMark(id)
	}
}
