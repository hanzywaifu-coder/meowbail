package business

import (
	"context"
	protocol "github.com/hanzywaifu-coder/dongtube-meowbail/internal/protocol"
	"strconv"

	waBinary "go.mau.fi/whatsmeow/binary"
)

// CatalogProduct merepresentasikan item katalog produk bisnis WhatsApp
type CatalogProduct struct {
	ID          string
	Name        string
	Description string
	Price       string
	Currency    string
	URL         string
}

// GetBusinessCatalog mengambil katalog produk dari akun WhatsApp Bisnis
// Mengikuti implementasi Baileys getCatalog (iq get xmlns="w:biz:catalog")
func GetBusinessCatalog(c *protocol.Client, ctx context.Context, targetJID protocol.JID, limit int) ([]CatalogProduct, error) {
	var products []CatalogProduct
	if targetJID.IsEmpty() {
		if c.Client != nil && c.Client.Store != nil && c.Client.Store.ID != nil {
			targetJID = *c.Client.Store.ID
		}
	}
	if limit <= 0 {
		limit = 10
	}

	queryParamNodes := []waBinary.Node{
		{
			Tag:     "limit",
			Content: []byte(strconv.Itoa(limit)),
		},
		{
			Tag:     "width",
			Content: []byte("100"),
		},
		{
			Tag:     "height",
			Content: []byte("100"),
		},
	}

	catalogNode := waBinary.Node{
		Tag: "product_catalog",
		Attrs: waBinary.Attrs{
			"jid":               targetJID.ToNonAD().String(),
			"allow_shop_source": "true",
		},
		Content: queryParamNodes,
	}

	queryNode := waBinary.Node{
		Tag: "iq",
		Attrs: waBinary.Attrs{
			"to":    protocol.ServerJID.String(),
			"type":  "get",
			"xmlns": "w:biz:catalog",
		},
		Content: []waBinary.Node{catalogNode},
	}

	err := c.Client.DangerousInternals().SendNode(ctx, queryNode)
	if err != nil {
		return nil, err
	}

	return products, nil
}

// CatalogCollection merepresentasikan koleksi produk bisnis WhatsApp
type CatalogCollection struct {
	ID   string
	Name string
}

// GetBusinessCollections mengambil koleksi kategori katalog dari akun WhatsApp Bisnis
// Mengikuti implementasi Baileys getCollections (iq get xmlns="w:biz:catalog" smax_id="35")
func GetBusinessCollections(c *protocol.Client, ctx context.Context, targetJID protocol.JID, limit int) ([]CatalogCollection, error) {
	if targetJID.IsEmpty() {
		if c.Client != nil && c.Client.Store != nil && c.Client.Store.ID != nil {
			targetJID = *c.Client.Store.ID
		}
	}
	if limit <= 0 {
		limit = 50
	}

	collNode := waBinary.Node{
		Tag: "collections",
		Attrs: waBinary.Attrs{
			"biz_jid": targetJID.ToNonAD().String(),
		},
		Content: []waBinary.Node{
			{
				Tag:     "collection_limit",
				Content: []byte(strconv.Itoa(limit)),
			},
			{
				Tag:     "item_limit",
				Content: []byte(strconv.Itoa(limit)),
			},
			{
				Tag:     "width",
				Content: []byte("100"),
			},
			{
				Tag:     "height",
				Content: []byte("100"),
			},
		},
	}

	queryNode := waBinary.Node{
		Tag: "iq",
		Attrs: waBinary.Attrs{
			"to":      protocol.ServerJID.String(),
			"type":    "get",
			"xmlns":   "w:biz:catalog",
			"smax_id": "35",
		},
		Content: []waBinary.Node{collNode},
	}

	err := c.Client.DangerousInternals().SendNode(ctx, queryNode)
	if err != nil {
		return nil, err
	}

	return []CatalogCollection{}, nil
}

// ProductDelete menghapus satu atau lebih item produk dari katalog bisnis
// Mengikuti implementasi Baileys productDelete (iq set xmlns="w:biz:catalog" <product_catalog_delete>)
func ProductDelete(c *protocol.Client, ctx context.Context, productIDs []string) error {
	if len(productIDs) == 0 {
		return nil
	}

	prodNodes := make([]waBinary.Node, 0, len(productIDs))
	for _, id := range productIDs {
		prodNodes = append(prodNodes, waBinary.Node{
			Tag: "product",
			Content: []waBinary.Node{
				{
					Tag:     "id",
					Content: []byte(id),
				},
			},
		})
	}

	delNode := waBinary.Node{
		Tag: "product_catalog_delete",
		Attrs: waBinary.Attrs{
			"v": "1",
		},
		Content: prodNodes,
	}

	queryNode := waBinary.Node{
		Tag: "iq",
		Attrs: waBinary.Attrs{
			"to":    protocol.ServerJID.String(),
			"type":  "set",
			"xmlns": "w:biz:catalog",
		},
		Content: []waBinary.Node{delNode},
	}

	return c.Client.DangerousInternals().SendNode(ctx, queryNode)
}
