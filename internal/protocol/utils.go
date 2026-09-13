package protocol

import (
	"encoding/json"
	"fmt"
	"go.mau.fi/whatsmeow/types"
	"strings"
	"time"
)

// FormatDuration formats duration to human readable string
func FormatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	mins := int(d.Minutes()) % 60
	secs := int(d.Seconds()) % 60
	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, mins)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	return fmt.Sprintf("%dm %ds", mins, secs)
}

// GetGreeting returns time-based greeting
func GetGreeting() string {
	hour := time.Now().Hour()
	if hour >= 4 && hour < 11 {
		return "Selamat Pagi"
	} else if hour >= 11 && hour < 15 {
		return "Selamat Siang"
	} else if hour >= 15 && hour < 18 {
		return "Selamat Sore"
	}
	return "Selamat Malam"
}

// ParseJID parses a string to JID
func ParseJID(s string) types.JID {
	jid, _ := types.ParseJID(s)
	return jid
}

// IsGroup checks if a JID is a group
func IsGroup(jid types.JID) bool {
	return jid.Server == types.GroupServer
}

// IsDM checks if a JID is a direct message
func IsDM(jid types.JID) bool {
	return jid.Server == types.DefaultUserServer || jid.Server == "s.whatsapp.net"
}

// IsNewsletter checks if a JID is a channel/newsletter
func IsNewsletter(jid types.JID) bool {
	return jid.Server == types.NewsletterServer || strings.HasSuffix(jid.String(), "@newsletter")
}

// IsStatusBroadcast checks if a JID is status broadcast
func IsStatusBroadcast(jid types.JID) bool {
	return jid.Server == types.BroadcastServer && jid.User == "status"
}

// IsMetaAI checks if a JID is Meta AI
func IsMetaAI(jid types.JID) bool {
	return jid.Server == "bot" || jid.User == "13135550002"
}

// IsLID checks if a JID is a LID (Linked Identity)
func IsLID(jid types.JID) bool {
	return jid.Server == "lid"
}

// NormalizeUserJID converts any user JID to standard s.whatsapp.net without device suffix
func NormalizeUserJID(jid types.JID) types.JID {
	return jid.ToNonAD()
}

// GetPhoneFromJID extracts phone number from JID
func GetPhoneFromJID(jid types.JID) string {
	return jid.User
}

// FormatPhone formats phone number with country code
func FormatPhone(phone string, countryCode ...string) string {
	code := "62"
	if len(countryCode) > 0 {
		code = countryCode[0]
	}
	// Remove leading zeros
	for len(phone) > 0 && phone[0] == '0' {
		phone = phone[1:]
	}
	// Add country code if not present
	if len(phone) > 0 && phone[0] != code[0] {
		phone = code + phone
	}
	return phone
}

// StringToPtr converts string to *string
func StringToPtr(s string) *string {
	return &s
}

// PtrToString converts *string to string
func PtrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// BoolToPtr converts bool to *bool
func BoolToPtr(b bool) *bool {
	return &b
}

// PtrToBool converts *bool to bool
func PtrToBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// Uint32ToPtr converts uint32 to *uint32
func Uint32ToPtr(v uint32) *uint32 {
	return &v
}

// PtrToUint32 converts *uint32 to uint32
func PtrToUint32(v *uint32) uint32 {
	if v == nil {
		return 0
	}
	return *v
}

// Float64ToPtr converts float64 to *float64
func Float64ToPtr(v float64) *float64 {
	return &v
}

// PtrToFloat64 converts *float64 to float64
func PtrToFloat64(v *float64) float64 {
	if v == nil {
		return 0
	}
	return *v
}

// MapToJSON converts a map to JSON string
func MapToJSON(m map[string]interface{}) string {
	b, err := json.Marshal(m)
	if err != nil {
		return "{}"
	}
	return string(b)
}

// Contains checks if a string slice contains a string
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ParseCommand extracts command from text with prefix
func ParseCommand(text, prefix string) string {
	if !strings.HasPrefix(text, prefix) {
		return ""
	}
	rest := strings.TrimSpace(strings.TrimPrefix(text, prefix))
	if rest == "" {
		return ""
	}
	return strings.ToLower(strings.Fields(rest)[0])
}

// Remove removes a string from a slice
func Remove(slice []string, item string) []string {
	for i, s := range slice {
		if s == item {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

// Unique removes duplicates from a string slice
func Unique(slice []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, s := range slice {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}

// SliceContainsInt checks if an int slice contains an int
func SliceContainsInt(slice []int, item int) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Max returns the larger of two ints
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Min returns the smaller of two ints
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
