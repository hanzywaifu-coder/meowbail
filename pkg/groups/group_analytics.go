package groups

import (
	"context"
	"fmt"
	protocol "github.com/hanzywaifu-coder/dongtube-meowbail/internal/protocol"
	"sort"
	"sync"
	"time"
)

// GroupMemberActivity catatan keaktifan anggota grup
type GroupMemberActivity struct {
	JID          protocol.JID `json:"jid"`
	MessageCount int          `json:"message_count"`
	LastActive   time.Time    `json:"last_active"`
}

// GroupAnalyticsTracker pelacak statistik aktivitas grup WhatsApp
type GroupAnalyticsTracker struct {
	mu         sync.RWMutex
	activities map[string]map[string]*GroupMemberActivity // groupJID -> senderJID -> activity
}

var DefaultGroupTracker = &GroupAnalyticsTracker{
	activities: make(map[string]map[string]*GroupMemberActivity),
}

// TrackMessage mencatat pesan yang masuk dari anggota grup
func (t *GroupAnalyticsTracker) TrackMessage(groupJID, senderJID protocol.JID) {
	t.mu.Lock()
	defer t.mu.Unlock()

	gKey := groupJID.ToNonAD().String()
	sKey := senderJID.ToNonAD().String()

	if _, ok := t.activities[gKey]; !ok {
		t.activities[gKey] = make(map[string]*GroupMemberActivity)
	}

	act, ok := t.activities[gKey][sKey]
	if !ok {
		act = &GroupMemberActivity{
			JID: senderJID,
		}
		t.activities[gKey][sKey] = act
	}

	act.MessageCount++
	act.LastActive = time.Now()
}

// GetTopActiveMembers mengambil daftar anggota paling aktif di grup
func (t *GroupAnalyticsTracker) GetTopActiveMembers(groupJID protocol.JID, limit int) []*GroupMemberActivity {
	t.mu.RLock()
	defer t.mu.RUnlock()

	gKey := groupJID.ToNonAD().String()
	mMap, ok := t.activities[gKey]
	if !ok || len(mMap) == 0 {
		return nil
	}

	var list []*GroupMemberActivity
	for _, act := range mMap {
		list = append(list, act)
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].MessageCount > list[j].MessageCount
	})

	if limit > 0 && len(list) > limit {
		list = list[:limit]
	}

	return list
}

// GetInactiveMembers mendeteksi anggota grup yang tidak pernah berkirim pesan
func GetInactiveMembers(c *protocol.Client, ctx context.Context, groupJID protocol.JID) ([]protocol.GroupParticipant, error) {
	info, err := GetGroupInfo(c, ctx, groupJID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil info grup: %w", err)
	}

	DefaultGroupTracker.mu.RLock()
	defer DefaultGroupTracker.mu.RUnlock()

	gKey := groupJID.ToNonAD().String()
	activeMap := DefaultGroupTracker.activities[gKey]

	var inactives []protocol.GroupParticipant
	for _, p := range info.Participants {
		pKey := p.JID.ToNonAD().String()
		if activeMap == nil || activeMap[pKey] == nil || activeMap[pKey].MessageCount == 0 {
			inactives = append(inactives, p)
		}
	}

	return inactives, nil
}
