package groups

import (
	"strconv"
	"time"

	waBinary "go.mau.fi/whatsmeow/binary"
)

// buildBizAdditionalNodes membangun XML stanza node <biz> yang diwajibkan oleh protokol WhatsApp
// untuk merender tombol interaktif (native_flow, single_select, cta_url, dll)
func BuildBizAdditionalNodes() []waBinary.Node {
	ts := strconv.FormatInt(time.Now().Unix()-77980457, 10)
	return []waBinary.Node{
		{
			Tag: "biz",
			Attrs: waBinary.Attrs{
				"actual_actors":   "2",
				"host_storage":    "2",
				"privacy_mode_ts": ts,
			},
			Content: []waBinary.Node{
				{
					Tag: "engagement",
					Attrs: waBinary.Attrs{
						"customer_service_state": "open",
						"conversation_state":     "open",
					},
				},
				{
					Tag: "interactive",
					Attrs: waBinary.Attrs{
						"type": "native_flow",
						"v":    "1",
					},
					Content: []waBinary.Node{
						{
							Tag: "native_flow",
							Attrs: waBinary.Attrs{
								"name": "mixed",
								"v":    "9",
							},
						},
					},
				},
			},
		},
	}
}
