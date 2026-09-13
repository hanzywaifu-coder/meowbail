# Dongtube-meowbail

> **Disclaimer / Peringatan Risiko**
> Library ini **tidak resmi** dan tidak berafiliasi dengan WhatsApp maupun Meta.
> Menggunakan library pihak ketiga untuk terhubung ke WhatsApp **melanggar Ketentuan
> Layanan (ToS) WhatsApp** dan dapat mengakibatkan nomor Anda **dibanned secara
> permanen**. Gunakan dengan risiko Anda sendiri — jangan pakai nomor utama Anda.
> Proyek ini dibuat untuk tujuan edukasi dan otomasi yang bertanggung jawab.

**Dongtube-meowbail** adalah library WhatsApp multi-device untuk Go yang menyatukan
kecepatan whatsmeow dengan fitur interaktif modern: tombol native flow, menu CDN,
newsletter/saluran, fake reply, pairing code kustom, dan reliability layer bawaan.

## Fitur

- **Engine cepat & hemat resource** — dibangun di atas whatsmeow (multi-device),
  dengan rate limiter per-chat, deduplikator pesan, LRU upload cache, buffer pool,
  dan retry anti-spiral. Default aman: rate limit aktif sejak menit pertama.
- **Pairing code kustom** — pairing dengan kode 8 karakter pilihan Anda sendiri
  (mis. `DONG-TUBE`), fitur first-class, bukan patch.
- **Pesan interaktif** — quick reply, CTA URL, CTA call, copy code, single select,
  carousel, product message — semuanya native flow terbaru.
- **Newsletter / Saluran** — posting teks & media, metadata, follow/unfollow,
  dengan konteks newsletter pada pesan keluar (fake reply & branding saluran).
- **LID → nomor asli** — resolver bawaan supaya mention/tag memakai nomor asli
  pengguna, bukan JID LID.
- **Fake reply & menu CDN** — builder context lengkap (troli, toko, Meta AI,
  saluran) persis klien resmi.
- **Struktur profesional** — facade tunggal `meowbail`, engine internal, dan
  package domain yang kecil-kecil dan jelas.

## Instalasi

```bash
go get github.com/hanzywaifu-coder/dongtube-meowbail@latest
```

Atau langsung dari source:

```bash
git clone https://github.com/hanzywaifu-coder/dongtube-meowbail.git
cd dongtube-meowbail && go build ./...
```

> Library ini self-hosted: Anda menjalankannya di resource Anda sendiri.
> Tidak ada layanan eksternal yang diperlukan.

## Contoh cepat

```go
package main

import (
    "context"
    "fmt"
    "os"
    "os/signal"

    meowbail "github.com/hanzywaifu-coder/dongtube-meowbail"
    "go.mau.fi/whatsmeow"
    "go.mau.fi/whatsmeow/store/sqlstore"
    waLog "go.mau.fi/whatsmeow/util/log"
    _ "modernc.org/sqlite"
)

func main() {
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
    defer stop()

    container, _ := sqlstore.New(ctx, "sqlite", "file:session.db?_pragma=journal_mode(WAL)", waLog.Stdout("DB", "WARN", true))
    device, _ := container.GetFirstDevice(ctx)

    waClient := whatsmeow.NewClient(device, waLog.Stdout("WA", "WARN", true))

    client := meowbail.NewClientFromWhatsmeow(
        waClient,
        &meowbail.Config{
            NewsletterJID:  "120363394448728781@newsletter",
            NewsletterName: "Dongtube",
            AutoReconnect:  true,
        },
    )

    client.AddEventHandler(func(evt interface{}) {
        msg := meowbail.ParseMessageEvent(evt)
        if msg == nil || msg.IsFromMe || msg.Text == "" {
            return
        }
        if meowbail.ParseCommand(msg.Text, ".") == "ping" {
            client.SendText(context.Background(), msg.Chat, "pong!")
        }
    })

    if client.Store.ID == nil {
        // Pairing dengan kode kustom 8 karakter:
        code, _ := client.PairPhoneWithCode(ctx, "628123456789", "DONGTUBE")
        fmt.Println("Pairing code:", code)
    }

    if err := client.Connect(ctx); err != nil {
        panic(err)
    }
    <-ctx.Done()
    client.Disconnect()
}
```

Contoh lengkap yang bisa langsung dijalankan ada di [`examples/`](examples/).

## Arsitektur

```
dongtube-meowbail/
├── meowbail.go            # Facade publik — satu import untuk semuanya
├── internal/protocol/     # Engine: client, session, pairing, reliability,
│                          # rate limiter, dedup, LID resolver, buffer pool
├── pkg/
│   ├── messaging/         # Kirim pesan, tombol, fake reply, menu, preview
│   ├── media/             # Upload/download, sticker, LRU upload cache
│   ├── groups/            # Grup: admin, invite, tag, picture, analytics
│   ├── channels/          # Newsletter/saluran: post, metadata, pin
│   ├── business/          # Katalog, cart, label, order, quick reply bisnis
│   ├── buttons/           # Native flow buttons, carousel, poll, response
│   └── privacy/           # Presence, privacy setting, kontak, community
├── third_party/whatsmeow/ # Vendored whatsmeow (MPL-2.0) + patch pairing code
├── nodejs/                # Layer Node.js standalone (bridging ke core Go menyusul)
└── examples/              # Contoh bot minimal
```

**Prinsip desain:** satu masalah, satu solusi terbaik. Event-driven, tanpa polling
yang tidak perlu, alokasi minim (buffer pool + LRU cache), dan reliability jujur:
reconnect stabil, rate limit aman-default, error handling matang.

## Reliability, bukan "antiban"

Fitur delay natural dan rate limit di library ini ada untuk **keandalan dan kesopanan
jaringan** — bukan untuk mengakali deteksi. Tidak ada library pihak ketiga yang bisa
menjamin akun bebas dari tindakan WhatsApp. Tetap gunakan secara wajar.

## Benchmark

Hasil `go test -bench` pada server produksi (AMD EPYC 9575F):

| Benchmark                 | ns/op  | B/op | allocs/op |
|---------------------------|--------|------|-----------|
| ExtractPureText           | 2.9    | 0    | 0         |
| BufferPool                | 18.8   | 0    | 0         |
| MessageDeduplicator       | 80.8   | 0    | 0         |
| RateLimiterWaitOrThrottle | 231.5  | 64   | 3         |
| ExtractMentions           | 502.8  | 384  | 4         |

## Lisensi

Kode asli proyek ini: **MIT** — lihat [LICENSE](LICENSE).
Vendored whatsmeow di `third_party/whatsmeow`: **MPL-2.0** — lisensinya dipertahankan
utuh sesuai ketentuan, lihat [NOTICE](NOTICE).

Node.js layer: `cd nodejs && node index.js` (lihat `nodejs/package.json`).
