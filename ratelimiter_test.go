package meowbail_test

import (
	"testing"
	"time"

	meowbail "github.com/hanzywaifu-coder/dongtube-meowbail"
	"go.mau.fi/whatsmeow/types"
)

func TestRateLimiterBurst(t *testing.T) {
	limiter := meowbail.NewRateLimiter(100*time.Millisecond, 3)
	jid := types.NewJID("123456789", types.DefaultUserServer)

	// First call should have 0 wait
	wait := limiter.WaitOrThrottle(jid)
	if wait != 0 {
		t.Fatalf("Expected 0 wait on first call, got %v", wait)
	}

	// Immediate 2nd and 3rd call should pass due to burstLimit = 3
	wait = limiter.WaitOrThrottle(jid)
	if wait != 0 {
		t.Fatalf("Expected 0 wait on 2nd call, got %v", wait)
	}

	wait = limiter.WaitOrThrottle(jid)
	if wait != 0 {
		t.Fatalf("Expected 0 wait on 3rd call, got %v", wait)
	}

	// 4th call should be throttled (> 0 wait)
	wait = limiter.WaitOrThrottle(jid)
	if wait <= 0 {
		t.Fatalf("Expected throttling on 4th call, got %v", wait)
	}
}
