package protocol

import (
	"context"

	"go.mau.fi/whatsmeow"
)

// PairPhoneWithCode requests a pairing code for the given phone number with a
// custom 8-character alphanumeric code (e.g. "DONGTUBE", formatted "DONG-TUBE").
//
// This is the dongtube-meowbail first-class custom pairing feature: the code is
// chosen by this client and any 8-char alphanumeric string works. Pass an empty
// customCode to let the server generate a random code.
//
// Implementation detail: the vendored whatsmeow encrypts the ephemeral public
// key exactly once with the custom code (see third_party/whatsmeow/pair-code.go).
func (c *Client) PairPhoneWithCode(ctx context.Context, phone, customCode string) (string, error) {
	if c == nil || c.Client == nil {
		return "", whatsmeow.ErrClientIsNil
	}
	return c.Client.PairPhoneWithCode(
		ctx, phone, customCode,
		true,
		whatsmeow.PairClientChrome,
		"Dongtube-meowbail",
	)
}
