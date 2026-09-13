# MANDAT REBUILD — Dongtube-meowbail v2 & Dongtube Bot

> Dokumen kerja resmi. Semua keputusan di bawah ini TERKUNCI oleh owner.
> Dibangun di folder baru; repo lama tidak disentuh sampai review lulus.

## 0. Prinsip

1. **Bukan tulis ulang buta.** Ilmu dari kode lama dipindahkan: custom pairing
   code, fake reply builder, menu CDN, newsletter context, LID resolver,
   retry/reliability, buffer pool. Yang dirombak adalah *struktur*, bukan
   *pengetahuan*.
2. **Jujur dari atas.** README dibuka dengan disclaimer risiko WhatsApp
   (library tidak resmi = melanggar ToS). Notice lisensi whatsmeow (MPL-2.0)
   dan dependency dasar lain tetap wajib di NOTICE — itu kewajiban hukum,
   bukan watermark.
3. **Branding publik 100% Dongtube-meowbail.** API, dokumentasi, komentar
   publik — tidak campur brand lain.
4. **Sedikit kode, bukan banyak kode.** Refactor = menutup kekurangan +
   menstabilkan, bukan menambah kompleksitas. Satu masalah → satu solusi.
5. **Reliability layer jujur.** Fitur "antiban" lama (delay natural, rate
   limit) direframe jadi `reliability` dengan dokumentasi jujur — proteksi
   akun dari flag Meta, bukan fitur "cheat".

## 1. Keputusan arsitektur (terkunci)

- **Model:** library/SDK self-hosted, publik, install dari GitHub (bukan npm
  registry). Orang jalankan sendiri pakai resource sendiri.
- **Dua layer:** `dongtube-meowbail` (library Go, core di atas whatsmeow) dan
  `Dongtube bot` (contoh implementasi di atasnya). Node.js dipertahankan dan
  direstrukturisasi standalone dulu, dijembatani ke core Go belakangan.
- **API publik:** root facade `meowbail.NewClient()` satu pintu; domain
  packages di belakangnya. User lama tidak rusak.
- **Struktur folder library:**

```
dongtube-meowbail/
├── go.mod, go.sum
├── README.md            # disclaimer risiko di atas + install GitHub
├── LICENSE, NOTICE      # whatsmeow MPL-2.0 & dependency dasar
├── client.go            # facade: type Client + NewClient() + delegasi top-30 API
├── internal/
│   └── protocol/        # wrapper whatsmeow: client, session, pairing, events,
│                        #   reliability, lid, media, util
├── pkg/
│   ├── session/         # device/session store API publik
│   ├── messaging/       # teks, reply, fake reply, mentions, menu CDN
│   ├── media/           # upload/download image/video/audio/sticker/document
│   ├── groups/          # info, participants (PN resolution), admin, invite
│   ├── channels/        # newsletter: join, posts, react, mute, management
│   ├── business/        # catalog, cart, labels, payment, interactive
│   ├── buttons/         # buttons v2, carousel, list, poll, native flow
│   └── privacy/         # presence, typing, profile, privacy settings
├── examples/
│   ├── basic/           # minimal: connect + pairing
│   └── bot/             # contoh bot command
└── nodejs/              # standalone, modul kecil: session/messaging/media
```

## 2. Kerja (order)

1. Scaffold repo baru + go.mod + NOTICE + LICENSE + README skeleton.
2. `internal/protocol/`: port dari root + core/ lama —
   - `client.go` — Client struct (embed whatsmeow.Client), connect/reconnect,
     event loop buffered (drop + counter, non-blocking)
   - `session.go` — SQLite WAL store, GetFirstDevice, session persist
   - `pairing.go` — PairPhone + **PairPhoneWithCode (custom code first-class,
     port dari patch whatsmeow kemarin + `generateCustomCompanionEphemeralKey`)**
   - `events.go` — dispatch buffered channel, backpressure
   - `reliability.go` — throttle, jitter, retry tracker, rate limiter
     (merge antiban_client.go + anti_spam.go + core/antiban.go + retry_tracker.go)
   - `lid.go` — LID↔PN resolver (bridge whatsmeow store)
   - `media.go` — upload/download glue
   - `util.go` — randHex, proto helpers, batch_buffer/buffer_pool
3. `pkg/*` domains — migrasi file lama per domain, gabung `_v2/_v3/_v4`
   (isi berbeda → gabung ke file domain dengan nama jelas, bukan dihapus):
   - messaging: messages.go, fake_reply.go, menu_cdn_*.go, message_extractor_v2.go
   - media: media.go, sticker.go, album_unified.go
   - groups: groups*.go, group_invite_v4.go
   - channels: channels_*.go, newsletter_flow_v4.go
   - business: business_*.go, business_interactive_v3.go
   - buttons: buttons_v2.go, carousel.go, button_response.go, poll
   - privacy: presence, profile, privacy, contacts, blocklist, call_link
   - session: store helpers publik
4. Root facade `client.go`: `NewClient(Config)`, delegasi top API.
5. `nodejs/` pecah index.js → modul kecil, bersihkan komentar.
6. README lengkap.
7. **Dongtube bot v2** di `/root/dongtube-bot-v2`:
   - `cmd/bot/main.go` thin (config + wiring)
   - `internal/config/` — settings.json, default aman
   - `internal/dispatch/` — worker pool, queue, dedup, backpressure
   - `internal/middleware/` — owner check, admin check (LID-aware), group-only,
     rate limit, logging — **semua pengecekan di satu lapisan**
   - `internal/reply/` — fake reply contexts + newsletter branding
   - `plugins/` — menu, ping, tagall (nomor asli), sticker, dst
8. Benchmark RAM/CPU per sesi → tag `v1.0.0` → push GitHub.

## 3. Pengetahuan yang dipindahkan (non-negotiable)

- Custom pairing code `DONGTUBE` (PairPhoneWithCode + vendor whatsmeow patch)
- Fake reply builder lengkap (Troli/Toko/MetaAI/Saluran/Payment/dst)
- Menu CDN native (thumbnail upload + sections + CTA)
- Newsletter context (ForwardedNewsletterMessageInfo + BusinessMessageForwardInfo)
- LID↔PN resolution (Store.LIDs + participant PhoneNumber/LID fields)
- Download media asli (core.DownloadAny) — bukan stub
- DSN SQLite: WAL + busy_timeout + foreign_keys (concurrency tinggi)
- Display name format `Browser (OS)` (server-required, 400 tanpa itu)
- Push presence available setelah connect/reconnect
- Event bridge: quoted media detection (view-once unwrap), text extraction