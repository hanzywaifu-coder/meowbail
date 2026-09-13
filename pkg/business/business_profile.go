package business

import (
	"context"
	protocol "github.com/hanzywaifu-coder/dongtube-meowbail/internal/protocol"

	waBinary "go.mau.fi/whatsmeow/binary"
)

// BusinessProfileProps mendefinisikan field profil bisnis WhatsApp
type BusinessProfileProps struct {
	Address     string
	Email       string
	Description string
	Websites    []string
}

// UpdateBusinessProfile memperbarui profil bisnis akun WhatsApp (alamat, email, deskripsi, website)
// Mengikuti implementasi Baileys updateBussinesProfile (iq set xmlns="w:biz")
func UpdateBusinessProfile(c *protocol.Client, ctx context.Context, props BusinessProfileProps) error {
	var nodes []waBinary.Node

	if props.Address != "" {
		nodes = append(nodes, waBinary.Node{
			Tag:     "address",
			Content: []byte(props.Address),
		})
	}
	if props.Email != "" {
		nodes = append(nodes, waBinary.Node{
			Tag:     "email",
			Content: []byte(props.Email),
		})
	}
	if props.Description != "" {
		nodes = append(nodes, waBinary.Node{
			Tag:     "description",
			Content: []byte(props.Description),
		})
	}
	for _, web := range props.Websites {
		if web != "" {
			nodes = append(nodes, waBinary.Node{
				Tag:     "website",
				Content: []byte(web),
			})
		}
	}

	bizNode := waBinary.Node{
		Tag: "business_profile",
		Attrs: waBinary.Attrs{
			"v":             "3",
			"mutation_type": "delta",
		},
		Content: nodes,
	}

	queryNode := waBinary.Node{
		Tag: "iq",
		Attrs: waBinary.Attrs{
			"to":    protocol.ServerJID.String(),
			"type":  "set",
			"xmlns": "w:biz",
		},
		Content: []waBinary.Node{bizNode},
	}

	return c.Client.DangerousInternals().SendNode(ctx, queryNode)
}

// BusinessProfileInfo data detail profil bisnis WhatsApp
type BusinessProfileInfo struct {
	JID         protocol.JID
	Address     string
	Description string
	Website     []string
	Email       string
	Category    string
}

// GetBusinessProfile mengambil detail profil bisnis akun WhatsApp (Baileys getBusinessProfile parity)
func GetBusinessProfile(c *protocol.Client, ctx context.Context, jid protocol.JID) (*BusinessProfileInfo, error) {
	if jid.IsEmpty() {
		if c.Client != nil && c.Client.Store != nil && c.Client.Store.ID != nil {
			jid = *c.Client.Store.ID
		}
	}

	queryNode := waBinary.Node{
		Tag: "iq",
		Attrs: waBinary.Attrs{
			"to":    protocol.ServerJID.String(),
			"type":  "get",
			"xmlns": "w:biz",
		},
		Content: []waBinary.Node{
			{
				Tag: "business_profile",
				Attrs: waBinary.Attrs{
					"v": "244",
				},
				Content: []waBinary.Node{
					{
						Tag: "profile",
						Attrs: waBinary.Attrs{
							"jid": jid.ToNonAD().String(),
						},
					},
				},
			},
		},
	}

	resp, err := c.Client.DangerousInternals().SendNodeAndGetData(ctx, queryNode)
	if err != nil {
		return nil, err
	}

	node, err := waBinary.Unmarshal(resp)
	if err != nil {
		return nil, err
	}

	bpNode, ok := node.GetOptionalChildByTag("business_profile")
	if !ok {
		return nil, nil
	}

	profNode, ok := bpNode.GetOptionalChildByTag("profile")
	if !ok {
		return nil, nil
	}

	info := &BusinessProfileInfo{
		JID: jid,
	}

	if addr, ok := profNode.GetOptionalChildByTag("address"); ok {
		info.Address = string(addr.Content.([]byte))
	}
	if desc, ok := profNode.GetOptionalChildByTag("description"); ok {
		info.Description = string(desc.Content.([]byte))
	}
	if email, ok := profNode.GetOptionalChildByTag("email"); ok {
		info.Email = string(email.Content.([]byte))
	}
	if web, ok := profNode.GetOptionalChildByTag("website"); ok {
		info.Website = append(info.Website, string(web.Content.([]byte)))
	}
	if cats, ok := profNode.GetOptionalChildByTag("categories"); ok {
		if cat, ok := cats.GetOptionalChildByTag("category"); ok {
			info.Category = string(cat.Content.([]byte))
		}
	}

	return info, nil
}
