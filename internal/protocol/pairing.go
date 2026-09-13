package protocol

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"go.mau.fi/util/random"
	"go.mau.fi/whatsmeow/util/keys"
	"golang.org/x/crypto/pbkdf2"
	"regexp"
	"strings"
)

var notNumRegex = regexp.MustCompile("[^0-9]")
var CrockfordBase32 = base32.NewEncoding("123456789ABCDEFGHJKLMNPQRSTVWXYZ")

// PairingManager menangani pembuatan pairing code kustom maupun standar
type PairingManager struct {
	Salt []byte
	IV   []byte
}

func NewPairingManager() *PairingManager {
	return &PairingManager{
		Salt: random.Bytes(32),
		IV:   random.Bytes(16),
	}
}

// GenerateCode membuat atau memformat custom pairing code 8 karakter (format: ABCD-EFGH)
func (pm *PairingManager) GenerateCode(custom string) (rawCode string, formatted string, err error) {
	if custom != "" {
		clean := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(custom), "-", ""))
		if len(clean) != 8 {
			return "", "", fmt.Errorf("custom pairing code must be exactly 8 characters")
		}
		return clean, clean[:4] + "-" + clean[4:], nil
	}
	codeBytes := random.Bytes(5)
	raw := CrockfordBase32.EncodeToString(codeBytes)
	return raw, raw[:4] + "-" + raw[4:], nil
}

// DeriveCompanionEphemeralKey menghitung ephemeral key untuk proses companion registration
func (pm *PairingManager) DeriveCompanionEphemeralKey(code string) (*keys.KeyPair, []byte) {
	ephemeralKeyPair := keys.NewKeyPair()
	linkCodeKey := pbkdf2.Key([]byte(code), pm.Salt, 2<<16, 32, sha256.New)
	linkCipherBlock, _ := aes.NewCipher(linkCodeKey)
	encryptedPubkey := make([]byte, len(ephemeralKeyPair.Pub))
	copy(encryptedPubkey, ephemeralKeyPair.Pub[:])
	cipher.NewCTR(linkCipherBlock, pm.IV).XORKeyStream(encryptedPubkey, encryptedPubkey)
	ephemeralKey := make([]byte, 80)
	copy(ephemeralKey[0:32], pm.Salt)
	copy(ephemeralKey[32:48], pm.IV)
	copy(ephemeralKey[48:80], encryptedPubkey)
	return ephemeralKeyPair, ephemeralKey
}

// CleanPhoneNumber membersihkan nomor telepon dan memvalidasi format internasional
func CleanPhoneNumber(phone string) (string, error) {
	clean := notNumRegex.ReplaceAllString(phone, "")
	if len(clean) <= 6 {
		return "", fmt.Errorf("phone number too short")
	}
	if strings.HasPrefix(clean, "0") {
		return "", fmt.Errorf("phone number must start with international country code (e.g. 62xxx)")
	}
	return clean, nil
}
