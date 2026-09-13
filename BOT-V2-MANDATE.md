# Dongtube Bot v2 — Mandat Rebuild

> Status: **siap dibangun** — core `dongtube-meowbail` v2 sudah stabil
> (build + vet + test + benchmark bersih, commit `9160d06`).
> Bot ini adalah **contoh implementasi resmi** di atas library `dongtube-meowbail`.
> Semua pengetahuan dari bot lama (`/root/dongtube-bot`) menjadi referensi migrasi,
> bukan disalin mentah.

## 0. Prinsip

1. **Library-first.** Setiap fitur bot harus lewat API asli `dongtube-meowbail`
   (facade `meowbail` / `pkg/*`). **Dilarang** membuat shim, stub, atau proto
   duplikat di sisi bot (kesalahan fatal bot lama: `internal/wae2e` shim cacat).
2. **Kecil dan jelas.** Struktur direktori datar sesedikit mungkin. 221 file
   plugin datar di bot lama = tidak profesional. Grup per kategori, file per
   fitur, satu tanggung jawab per file.
3. **Event-driven.** Tidak ada polling. Satu handler event → dispatcher →
   middleware → plugin registry. Backpressure via queue bila perlu.
4. **Resource hemat.** RAM < 15 MB idle, startup < 1 detik, 0 goroutine bocor.
   Reuse `meowbail.RateLimiter`, `MessageDeduplicator`, buffer pool dari library —
   bot tidak menggandakan mekanisme itu.
5. **Branding 100% Dongtube.** Nama, teks menu, pesan error, device props.
   Notice lisensi tetap dijaga (kewajiban hukum, bukan watermark).

## 1. Struktur target

```
dongtube-bot/
├── main.go                 # wiring: store → client → dispatcher → registry
├── config.go               # env/file config: owner, prefix, channel, session dir
├── internal/
│   ├── dispatch/           # event → MessageEvent → middleware chain → plugins
│   │   ├── dispatcher.go   #   core loop (goroutine pool kecil, worker N=CPU)
│   │   ├── middleware.go   #   dedup, owner check, admin check, rate limit
│   │   └── context.go      #   *Ctx: akses lib (client), msg, args, reply helper
│   └── plugin/
│       ├── registry.go     # registrasi: name, category, handler, permission
│       └── loader.go       # daftar plugin aktif (import langsung, tanpa reflect)
├── plugins/
│   ├── general/            # ping, menu, info, runtime
│   ├── group/              # tagall, welcome, antilink, kick/add, promote
│   ├── owner/              # upsw, join, leave, broadcast, setpp
│   ├── downloader/         # play, ytmp3/ytmp4, tiktok, ig (via API eksternal)
│   ├── sticker/            # s, spack, take, exif (via meowbail media pkg)
│   └── tools/              # quote, readmore, react, cek resi, dll
├── go.mod                  # require github.com/hanzywaifu-coder/dongtube-meowbail
└── README.md
```

Perkiraan total: **< 40 file**, masing-masing < 300 baris.

## 2. Kontrak plugin (satu tipe, tanpa magic)

```go
package plugins

type Plugin struct {
    Name       string                            // "ping"
    Category   string                            // "general"
    Permission string                            // "all" | "admin" | "owner"
    Handler    func(ctx *dispatch.Ctx) error     // error ditangani dispatcher
}
```

- `Ctx` membungkus `*meowbail.Client`, `*meowbail.MessageEvent`, args, dan helper
  `Reply(text, fake...)` yang **otomatis** menerapkan fake reply + newsletter
  context dari `Config.DefaultFakeReply` di library (fitur v2 sudah ada).
- Permission dicek middleware, bukan di tiap plugin → tidak ada lagi bug
  "bot harus admin" yang tidak konsisten (pelajaran dari audit bot lama).

## 3. Aturan perilaku yang wajib benar (dari audit bot lama)

1. **Tag/mention = nomor asli.** Melalui `LIDResolver` library +
   `ResolveParticipantPN` — bukan JID `@lid`.
2. **Dedup benar arah.** `MessageDeduplicator` library: `CheckAndMark` = true
   berarti **baru** → proses. (Bug fatal bot lama: logika terbalik.)
3. **Handler tidak blocking.** Dispatcher menjalankan plugin di worker pool;
   whatsmeow event loop tidak pernah dibekukan (pelajaran bug `.menu`).
4. **Media download = `client.DownloadMedia`.** Tanpa stub "not implemented".
5. **Admin check lintas LID↔PN.** Pakai `IsUserAdmin` library.
6. **Pairing code `DONGTUBE`.** Via `client.PairPhoneWithCode(ctx, phone, "DONGTUBE")`
   — first-class di library v2.
7. **Device name "Dongtube.id".** Saat registrasi (device props), bukan "whatsmeow".

## 4. Urutan eksekusi

1. Scaffold `go.mod` + `main.go` + `config.go` (session SQLite, auto-reconnect).
2. `internal/dispatch` + `internal/plugin` (registry, middleware, Ctx).
3. Migrasi plugin prioritas: ping, menu, tagall, sticker (.s), upsw, play.
4. Sisanya per kategori, satu commit per kategori.
5. Build + vet + smoke test real device (pairing → .ping → .menu → .tagall).
6. Benchmark RAM/CPU idle + saat burst 100 pesan → catat di README.

## 5. Yang TIDAK dikerjakan (jaga fokus)

- Tidak ada framework plugin eksternal / reflect-based loader (berat + tidak perlu).
- Tidak ada database tambahan — session pakai SQLite store whatsmeow via library.
- Tidak ada fallback ganda: satu jalur error handling yang benar.
- Tidak memindahkan logika reliability ke bot — itu tugas library.
