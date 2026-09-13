package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	meowbail "github.com/hanzywaifu-coder/dongtube-meowbail"
	qrcode "github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
	_ "modernc.org/sqlite"
)

const (
	CHANNEL_JID  = "120363394448728781@newsletter"
	CHANNEL_NAME = "Dongtube"
	OWNER        = "6283143961588"
	PREFIX       = "."
)

var botBootAt = time.Now()

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Initialize database
	dbLog := waLog.Stdout("DB", "WARN", true)
	os.MkdirAll("./wa-session", 0755)
	container, err := sqlstore.New(ctx, "sqlite", "file:./wa-session/session.db?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)", dbLog)
	if err != nil {
		log.Fatal("store:", err)
	}

	device, err := container.GetFirstDevice(ctx)
	if err != nil {
		log.Fatal("device:", err)
	}

	// Create whatsmeow client
	clientLog := waLog.Stdout("WA", "WARN", true)
	waClient := whatsmeow.NewClient(device, clientLog)
	waClient.EnableAutoReconnect = true

	// Create dongtube-meowbail client
	client := meowbail.NewClientFromWhatsmeow(waClient, &meowbail.Config{
		NewsletterJID:    CHANNEL_JID,
		NewsletterName:   CHANNEL_NAME,
		BusinessOwnerJID: OWNER + "@s.whatsapp.net",
		AutoReconnect:    true,
	})

	// Add event handler
	client.AddEventHandler(func(evt interface{}) {
		// Parse message event
		msg := meowbail.ParseMessageEvent(evt)
		if msg == nil || msg.IsFromMe {
			return
		}

		// Handle button responses
		if selectedID, isButton := meowbail.HandleButtonResponse(evt); isButton {
			chatType := "DM"
			if msg.IsGroup {
				chatType = "GRP"
			}
			log.Printf("[%s] @%s: [button] %s", chatType, msg.Sender.User, selectedID)

			switch selectedID {
			case "ping":
				uptime := time.Since(botBootAt)
				client.SendText(context.Background(), msg.Chat,
					fmt.Sprintf("╭┈┈⬡「 ᴘɪɴɢ 」\n┃ ᴜᴘᴛɪᴍᴇ : %s\n╰┈┈┈┈┈┈┈┈⬡",
						meowbail.FormatDuration(uptime)))
			case "menu":
				sendMenu(client, msg)
			}
			return
		}

		// Handle text commands
		if msg.Text == "" || len(msg.Text) < 2 || msg.Text[0] != PREFIX[0] {
			return
		}

		cmd := meowbail.ParseCommand(msg.Text, PREFIX)
		if cmd == "" {
			return
		}

		chatType := "DM"
		if msg.IsGroup {
			chatType = "GRP"
		}
		log.Printf("[%s] @%s: %s", chatType, msg.Sender.User, msg.Text)

		switch cmd {
		case "menu", "help", "start":
			sendMenu(client, msg)
		case "ping":
			uptime := time.Since(botBootAt)
			client.SendText(context.Background(), msg.Chat,
				fmt.Sprintf("╭┈┈⬡「 ᴘɪɴɢ 」\n┃ ᴜᴘᴛɪᴍᴇ : %s\n╰┈┈┈┈┈┈┈┈⬡",
					meowbail.FormatDuration(uptime)))
		case "tagall":
			if msg.IsGroup {
				client.TagAll(context.Background(), msg.Chat, "╭┈┈⬡「 ᴛᴀɢ ᴀʟʟ 」")
			}
		case "groupinfo":
			if msg.IsGroup {
				sendGroupInfo(client, msg)
			}
		}
	})

	// QR code handling
	if waClient.Store.ID == nil {
		qrChan, _ := waClient.GetQRChannel(ctx)
		go func() {
			for evt := range qrChan {
				switch evt.Event {
				case "code":
					png, _ := qrcode.Encode(evt.Code, qrcode.Medium, 512)
					os.WriteFile("/tmp/dongtube-qr.png", png, 0644)
					fmt.Println("QR CODE SAVED: /tmp/dongtube-qr.png")
				case "expired":
					log.Println("QR expired, waiting...")
				}
			}
		}()
	}

	// Connect
	if err := client.Connect(ctx); err != nil {
		log.Fatal("connect:", err)
	}

	log.Println("Bot started!")
	<-ctx.Done()
	client.Disconnect()
}

func sendMenu(client *meowbail.Client, msg *meowbail.MessageEvent) {
	uptime := time.Since(botBootAt)
	menuText := fmt.Sprintf(
		"╭┈┈⬡「 ɪɴꜰᴏ ʙᴏᴛ 」\n"+
			"┃ ɴᴀᴍᴇ     : Dongtube\n"+
			"┃ ᴠᴇʀꜱɪᴏɴ  : v1.0.0\n"+
			"┃ ᴜᴘᴛɪᴍᴇ   : %s\n"+
			"┃ ᴍᴏᴅᴇ     : ꜱᴇʟꜰ\n"+
			"┃ ᴄᴏᴍᴍᴀɴᴅꜱ : 13\n"+
			"╰┈┈┈┈┈┈┈┈⬡\n\n"+
			"╭┈┈⬡「 ᴋᴀᴛᴇɢᴏʀɪ 」\n"+
			"┃ GROUP : 10 ᴄᴏᴍᴍᴀɴᴅ\n"+
			"┃ MAIN : 1 ᴄᴏᴍᴍᴀɴᴅ\n"+
			"┃ MISC : 1 ᴄᴏᴍᴍᴀɴᴅ\n"+
			"┃ OWNER : 1 ᴄᴏᴍᴍᴀɴᴅ\n"+
			"╰┈┈┈┈┈┈┈┈⬡",
		meowbail.FormatDuration(uptime),
	)

	client.SendButtons(context.Background(), msg.Chat, menuText, []meowbail.Button{
		{Type: meowbail.ButtonQuickReply, ID: "ping", Text: "🏓 Pɪɴɢ"},
		{Type: meowbail.ButtonQuickReply, ID: "menu", Text: "📋 Aʟʟ Mᴇɴᴜ"},
		{Type: meowbail.ButtonCTAURL, Text: "👤 Oᴡɴᴇʀ", DisplayText: "Chat Owner", URL: fmt.Sprintf("https://wa.me/%s", OWNER)},
	})
}

func sendGroupInfo(client *meowbail.Client, msg *meowbail.MessageEvent) {
	info, err := client.GetGroupInfo(context.Background(), msg.Chat)
	if err != nil {
		client.SendText(context.Background(), msg.Chat, "❌ Gagal get info grup")
		return
	}

	adminCount := 0
	for _, p := range info.Participants {
		if p.IsAdmin {
			adminCount++
		}
	}

	text := fmt.Sprintf(
		"╭┈┈⬡「 ɪɴꜰᴏ ɢʀᴜᴘ 」\n"+
			"┃ ɴᴀᴍᴇ     : %s\n"+
			"┃ ᴍᴇᴍʙᴇʀ   : %d\n"+
			"┃ ᴀᴅᴍɪɴ    : %d\n"+
			"╰┈┈┈┈┈┈┈┈⬡",
		info.Name,
		len(info.Participants),
		adminCount,
	)

	client.SendText(context.Background(), msg.Chat, text)
}
