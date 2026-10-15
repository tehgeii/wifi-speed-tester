# WiFi Speed + Ping Tester

> Cek kecepatan. Ukur ping. Pahami koneksimu.

Aplikasi Windows yang ringan dan portable untuk mengukur **download, upload,
ping, jitter, packet loss**, dan **ping saat sibuk** (latency under load).
Aplikasi ini juga menampilkan detail jaringan dan Wi-Fi, lalu memberi penilaian
koneksi dengan bahasa yang mudah dipahami. Tidak perlu instal, tidak ada
layanan yang berjalan di latar belakang, dan tidak ikut menyala saat Windows
dibuka: cukup ekstrak zip-nya lalu jalankan `WiFiSpeedTester.exe`.

```
WiFiSpeedTester-Portable-<versi>-win-x64.zip
├── WiFiSpeedTester.exe
├── WiFiSpeedTester.config.example.json
├── README.txt
├── LICENSE.txt
└── THIRD_PARTY_NOTICES.txt
```

## Download

Ambil versi terbaru di
**[Releases](https://github.com/tehgeii/wifi-speed-tester/releases/latest)**:

- **`…-win-x64.zip`**: untuk kebanyakan PC/laptop Windows 10/11 (Intel / AMD).
- **`…-win-arm64.zip`**: untuk Windows on ARM (Snapdragon).

Ekstrak zip-nya, lalu klik dua kali `WiFiSpeedTester.exe`. Aplikasi ini belum
punya tanda tangan digital (code signing), jadi Windows SmartScreen mungkin
menampilkan "Windows protected your PC". Klik **More info → Run anyway** (di
Windows berbahasa Indonesia: **Info selengkapnya → Tetap jalankan**).

## Fitur

| Bagian | Isinya |
|--------|--------|
| Deteksi jaringan | Adapter yang aktif, jenis koneksi (Wi-Fi / Ethernet / hotspot-tethering / seluler / VPN), IPv4, IPv6, gateway, DNS |
| Detail Wi-Fi | SSID, BSSID, sinyal %, RSSI, band, channel, frekuensi, standar Wi-Fi, link speed (diberi keterangan bahwa ini *bukan* kecepatan internet) |
| Cek internet | Adapter ✓ → Gateway ✓ → DNS ✓ → Internet ✓; berhenti dengan pesan yang jelas kalau internet tidak bisa dijangkau |
| Ping | Bawaannya ke router, Cloudflare DNS dan Google DNS, ditambah sampai 5 target pilihanmu sendiri (misalnya server game, di **Pengaturan ⚙**): terkirim/diterima, min/rata-rata/maks, jitter, loss |
| Download / Upload | Tes HTTP dengan beberapa koneksi sekaligus, masa pemanasan tidak dihitung, gauge langsung, grafik throughput, coba ulang dan pindah ke server cadangan kalau gagal |
| Ping saat sibuk | Ping diukur selama download dan upload, dengan petunjuk bufferbloat yang hati-hati (tidak langsung memvonis) |
| Rating | SANGAT BAGUS / BAGUS / CUKUP / BURUK / TIDAK STABIL, berdasarkan batas nilai yang terdokumentasi dan bisa diatur ([docs/QUALITY.md](docs/QUALITY.md)). **Kenapa rating-nya segini?** menampilkan nilai dan batas tiap metrik |
| Saran | Langkah nyata sesuai hasil yang terukur (nyalakan SQM/QoS, pakai 5 GHz, coba kabel LAN, restart router, VPN aktif, hubungi ISP…) |
| Paket ISP kamu | Isi kecepatan paket internetmu sekali saja (**Pengaturan ⚙**); hasil tes lalu menampilkan "62% dari paket", dan muncul saran kalau kecepatannya jauh di bawah paket |
| Mode Gaming | Menilai ping, jitter, loss, dan kestabilan ping, bukan besar bandwidth |
| Ping Cepat | Sekitar 5 detik: cek koneksi dan ping saja, untuk cek kilat sebelum main game |
| Server tes | Bawaannya Cloudflare; **Cari server terdekat** mengurutkan server LibreSpeed publik berdasarkan latensi, server pilihanmu dipakai duluan dan Cloudflare jadi cadangan |
| Bahasa | Bahasa Inggris dan Bahasa Indonesia (saat pertama dibuka mengikuti bahasa Windows; bisa diganti di **Pengaturan ⚙**) |
| Satuan | Mbps (bawaan) atau MB/s (80 Mbps = 10 MB/s) |
| Export | **Bagikan**: **Salin hasil** dan **Kartu PNG**, yang tidak pernah memuat nama Wi-Fi atau alamat IP. **Simpan sebagai file**: TXT, JSON dan CSV, yang menyamarkan data itu (`Ho*****`, `192.168.1.x`) kecuali kamu mencentang **Sertakan identitas jaringan**; keterangan di bawah tombol menunjukkan persis apa yang disembunyikan. TXT memuat nama Wi-Fi dan IP, JSON juga BSSID, CSV hanya nama Wi-Fi |
| Riwayat | Disimpan di komputer sendiri sebagai JSON di samping file exe; grafik tren kecepatan dan ping dari waktu ke waktu, rata-rata per jenis koneksi (Wi-Fi vs Ethernet…), % dari paket, filter 7/30 hari, buka ulang hasil, export ke CSV, hapus |
| Laporan ISP | **Riwayat → Laporan ISP** membuat ringkasan teks untuk dikirim ke ISP: rata-rata, % dari paket, hasil per jam, tes paling lambat, dan hasil monitor stabilitas |
| Monitor stabilitas | **Alat** → ping ke satu target dan ke router setiap detik selama 5–30 menit untuk menangkap putus sesaat dan lonjakan lag, dengan grafik langsung, kesimpulan (stabil / ada lonjakan / putus; masalahnya di router atau setelahnya) dan laporan TXT. Hanya berjalan selama jendela aplikasi terbuka, dan berhenti saat kamu menekan Stop atau menutup aplikasi |
| Tes DNS | **Alat** → membandingkan seberapa cepat DNS kamu sekarang, Cloudflare, Google dan Quad9 menjawab, lalu bilang perlu ganti atau tidak |
| Cek channel Wi-Fi | **Alat** → menampilkan jaringan Wi-Fi di sekitar (hanya membaca, dari Windows) dan seberapa ramai tiap channel, plus saran channel untuk diatur di router |
| Salin hasil | Ringkasan singkat untuk WhatsApp, Discord atau di mana saja, tanpa nama Wi-Fi atau alamat IP |
| Update | Cek versi baru saat aplikasi dibuka (bisa dimatikan di **Pengaturan ⚙**), dengan link ke halaman download |
| Error | Penyebab dan saran dengan bahasa sederhana, detail teknisnya bisa dibuka kalau perlu |
| Batal | Menghentikan semua aktivitas jaringan saat itu juga (tombol Esc juga bisa) |
| Satu jendela | Kalau exe-nya dibuka lagi, jendela yang sudah terbuka langsung muncul ke depan, bukan membuka jendela kedua |

## Kebutuhan sistem

- Windows 10 atau 11, x64 atau ARM64.
- [Microsoft Edge WebView2 Runtime](https://developer.microsoft.com/microsoft-edge/webview2/),
  yang sudah ada di Windows 11 dan kebanyakan Windows 10. Kalau belum ada,
  aplikasi akan memberi tahu dan menawarkan link download-nya. Mode `--cli`
  tetap jalan tanpa WebView2.
- Tidak perlu hak administrator. Ping memakai fitur ICMP bawaan Windows, bukan
  raw socket.

## Privasi

Semua tes berjalan di komputermu sendiri, kecuali lalu lintas yang memang
dibutuhkan untuk mengukur koneksi:

- data speed test ke server tes yang dipakai (bawaannya Cloudflare, `speed.cloudflare.com`);
- ping (ICMP echo) ke router, ke target bawaan (`1.1.1.1`, `8.8.8.8`), dan ke target yang kamu tambahkan sendiri;
- pencarian DNS untuk nama server tes;
- hanya saat kamu menjalankan **Tes DNS**: pencarian nama situs populer
  (google.com, youtube.com…) ke DNS kamu sekarang dan ke `1.1.1.1`, `8.8.8.8`
  dan `9.9.9.9`;
- hanya selama **Monitor stabilitas** berjalan: satu ping per detik ke target
  yang dipilih dan ke router;
- hanya saat kamu menekan **Cari server terdekat**: daftar server publik dari
  `librespeed.org` dan satu permintaan kecil ke tiap server di daftar itu;
- hanya kalau **Cek update saat aplikasi dibuka** menyala (**Pengaturan ⚙**,
  menyala secara bawaan): satu permintaan ke `api.github.com` untuk melihat
  versi terbaru.

**Cek channel Wi-Fi** tidak mengirim apa pun: fitur ini hanya meminta daftar
jaringan yang memang sudah terlihat oleh Windows. Aplikasi ini tidak membaca
file, riwayat browser atau password, tidak punya telemetri, dan tidak berjalan
di latar belakang. Dialog **Tentang** (tombol `?`) menampilkan ke mana saja
lalu lintas dikirim sesuai pengaturan yang sedang dipakai.

File yang dibuat aplikasi:

- Folder `WiFiSpeedTester-data\` di samping exe berisi `history.json`,
  `settings.json` (bahasa, satuan, mode, cek update, kecepatan paket),
  `server.json` (server yang dipilih di aplikasi, kalau ada), `targets.json`
  (target ping tambahanmu) dan `monitors.json` (20 ringkasan monitor
  stabilitas terakhir). Kalau folder itu tidak bisa ditulisi, aplikasi
  memakai `%APPDATA%\WiFiSpeedTester`.
- `%LOCALAPPDATA%\WiFiSpeedTester\WebView2` adalah cache mesin browser.

> Di Windows 11, nama Wi-Fi (SSID) baru terlihat kalau
> *Settings → Privacy & security → Location → Let desktop apps access your
> location* dinyalakan (nama menunya bisa sedikit berbeda di Windows berbahasa
> Indonesia). Fitur lain tetap jalan tanpa itu, dan aplikasi menampilkan
> keterangan kalau SSID disembunyikan.

## Akurasi

Hasil tes dipengaruhi server tes, rute jaringan, ISP, sinyal Wi-Fi, pemakaian
lain di jaringan yang sama, VPN, dan perangkatnya sendiri. Aplikasi ini
melaporkan apa yang **terukur** ("Download terukur 48.6 Mbps") dan tidak
pernah menyebutnya sebagai kecepatan paket ISP-mu.

Cara mengukurnya:

- **Download**: 4 koneksi HTTP/1.1 paralel (masing-masing satu koneksi TCP)
  mengalirkan data selama 10 detik. Ukuran permintaan dimulai dari 10 MB dan
  naik sampai 100 MB di koneksi cepat, supaya jumlah permintaannya sedikit
  dan server tidak membatasi (rate-limit). Kalau server menolak ukuran yang
  lebih besar, tes tetap memakai ukuran terbesar yang berhasil. 1,5 detik
  pertama (masa pemanasan) tidak dihitung, dan kecepatannya dihitung dari data
  yang benar-benar diterima. Kalau server berhenti menjawab di tengah tes,
  data yang sudah diterima tetap dilaporkan, dengan tanda "tidak lengkap".
- **Upload**: 4 koneksi paralel mengirim (POST) data yang tidak bisa
  dikompresi. Hanya permintaan yang sudah **dikonfirmasi** server yang
  dihitung, masing-masing sesuai bagian yang masuk ke jendela pengukuran,
  jadi data yang masih tertahan di buffer komputer tidak bisa membuat hasilnya
  terlihat lebih tinggi.
- **Ping**: ping dikirim setiap 200 ms tanpa menunggu balasan sebelumnya,
  sama seperti perintah `ping`. Jitter adalah rata-rata selisih antara dua
  hasil ping yang berurutan.
- **ICMP diblokir?** Beberapa jaringan (kantor, cloud) memblokir ping sama
  sekali. Kalau tidak ada target internet yang menjawab, latensi diukur dari
  waktu membuka koneksi TCP ke server tes, dan hasilnya memberi keterangan
  begitu. Monitor stabilitas juga melakukan hal yang sama.
- **Upload sangat lambat**: kalau tidak ada satu pun permintaan upload yang
  selesai tepat waktu (di bawah sekitar 0,1 Mbps), kecepatan upload
  diperkirakan dari data yang terkirim dan diberi tanda "tidak lengkap",
  bukan dianggap gagal.
- **Monitor stabilitas**: satu ping per detik. *Putus* artinya 3 ping atau
  lebih berturut-turut hilang; *lonjakan* artinya balasan yang jauh di atas
  ping biasanya (lebih dari dua kali median dan minimal 50 ms di atasnya).
- **Tes DNS**: 2 putaran berisi 10 nama situs populer untuk tiap server, lewat
  UDP biasa, jadi cache DNS Windows tidak ikut berpengaruh. Yang dibandingkan
  adalah waktu jawab biasanya (median). Server yang belum menjawab setelah 3
  kali dicoba dilewati.
- **Cek channel Wi-Fi**: jaringan tetangga dihitung berdasarkan kuat
  sinyalnya; di 2,4 GHz, channel yang saling tumpang-tindih dihitung
  sebagian. Di 2,4 GHz hanya channel 1, 6 dan 11 yang disarankan. Jaringan
  dengan nama yang sama dengan jaringanmu tidak dihitung.

## Konfigurasi

Buat file `WiFiSpeedTester.config.json` di samping exe. Cara paling gampang:
ubah nama file `WiFiSpeedTester.config.example.json` yang ikut di zip. Bagian
yang tidak kamu tulis tetap memakai nilai bawaan. Kalau file-nya salah format,
aplikasi memberi tahu di layar dan kembali memakai pengaturan bawaan.

```jsonc
{
  "servers": [                       // dicoba berurutan; yang berikutnya jadi cadangan
    { "name": "Cloudflare",
      "downloadUrl": "https://speed.cloudflare.com/__down?bytes={bytes}",
      "uploadUrl":   "https://speed.cloudflare.com/__up" },
    { "name": "My LibreSpeed", "type": "librespeed",   // opsional
      "url": "https://speedtest.example.com/backend/" }
  ],
  "pingTargets": [                   // "gateway" = router kamu
    { "label": "Local Gateway",  "host": "gateway" },
    { "label": "Cloudflare DNS", "host": "1.1.1.1" },
    { "label": "Google DNS",     "host": "8.8.8.8" }
  ],
  "pingCount": 20, "pingIntervalMs": 200, "pingTimeoutMs": 1000,
  "downloadSeconds": 10, "uploadSeconds": 10, "warmupSeconds": 1.5,
  "streams": 4,
  "measureLoadedPing": true,
  "connectivityHost": "speed.cloudflare.com",
  "serverListUrl": "https://librespeed.org/backend-servers/servers.php",
  "dnsServers": [                    // DNS publik yang dibandingkan di Tes DNS (alamat IP)
    { "label": "Cloudflare", "host": "1.1.1.1" },
    { "label": "Google",     "host": "8.8.8.8" },
    { "label": "Quad9",      "host": "9.9.9.9" }
  ],
  "planLowPct": 50,                  // di bawah % paket ini, muncul saran yang bisa dilakukan
  "historyLimit": 200,
  "profiles": { "general": { ... }, "gaming": { ... }, "quick": { ... } }   // lihat docs/QUALITY.md
}
```

Server berbentuk template URL harus menerima `GET <downloadUrl>` (dengan
`{bytes}` diganti jumlah byte) dan mengirim balik data sebanyak itu. Server
itu juga harus menerima `POST` dengan ukuran berapa pun ke `uploadUrl`. Server
`librespeed` cukup diisi `url` backend-nya; aplikasi memakai `garbage.php` dan
`empty.php` di sana (bisa diganti lewat `dlPath`, `ulPath`, `pingPath`). Pakai
hanya server yang memang boleh kamu pakai untuk tes.

## Command line

```
WiFiSpeedTester.exe                 buka jendela aplikasi
WiFiSpeedTester.exe --cli           jalankan tes di terminal dan tampilkan laporan
WiFiSpeedTester.exe --cli --gaming  pakai profil gaming
WiFiSpeedTester.exe --cli --quick   ping saja (sekitar 5 detik)
WiFiSpeedTester.exe --cli --lang id laporan dalam Bahasa Indonesia
WiFiSpeedTester.exe --cli --json    hasil lengkap dalam JSON (simpan dengan > hasil.json)
WiFiSpeedTester.exe --list-servers  urutkan server LibreSpeed publik berdasarkan latensi
WiFiSpeedTester.exe --dns           bandingkan server DNS
WiFiSpeedTester.exe --wifi-scan     tampilkan jaringan Wi-Fi di sekitar dan saran channel
WiFiSpeedTester.exe --print-config  tampilkan konfigurasi bawaan
WiFiSpeedTester.exe --write-config  buat file WiFiSpeedTester.config.json di samping exe
```

## Build (untuk developer)

Butuh Go 1.24 atau lebih baru. Tidak perlu compiler C: build untuk Windows
murni memakai Go (`CGO_ENABLED=0`) dan bisa dibuat dari Linux atau macOS.

```sh
scripts/build.sh 1.0.0      # -> dist/WiFiSpeedTester-Portable-1.0.0-win-x64.zip (+ win-arm64)
go test ./...               # tes unit, engine, dan app
scripts/ui-test.sh          # tes UI end-to-end (butuh Node dan Playwright, lihat di bawah)
```

CI (`.github/workflows/build.yml`) menjalankan `go vet` dan semua tes di Linux
dan Windows, menjalankan tes UI di Chromium, dan membuat kedua zip sebagai
artifact. CI juga menjalankan smoke test sungguhan untuk `--cli`, `--dns` dan
`--wifi-scan` di runner Windows.

Untuk menerbitkan rilis: buka **Actions → build → Run workflow**, pilih
`main`, lalu isi versinya, misalnya `1.0.1`. Workflow akan membuat kedua zip
dan `SHA256SUMS.txt`, membuat tag `v1.0.1`, lalu menerbitkan rilis dengan
catatan dari `packaging/RELEASE_NOTES.md` (perbarui dulu bagian "What's
new"-nya). Push tag `v*` juga menghasilkan hal yang sama.

Ikon, manifest DPI dan info versi exe berasal dari
`cmd/wifispeedtester/rsrc_windows_*.syso`. Buat ulang file itu setelah
mengubah isi folder `winres/` dengan:

```sh
go run github.com/tc-hib/go-winres@v0.3.3 make --in winres/winres.json --out cmd/wifispeedtester/rsrc --arch amd64,arm64
```

### Mengerjakan UI tanpa Windows

Jendela aplikasi sebenarnya adalah halaman web (`internal/ui/web`) yang
ditampilkan lewat WebView2. Halaman dan API yang sama bisa dijalankan secara
lokal:

```sh
go run ./cmd/fakespeedserver &          # server speed test lokal, kecepatannya dibatasi
# arahkan WiFiSpeedTester.config.json ke http://127.0.0.1:8099/down?bytes={bytes} dan /up
go run ./cmd/wifispeedtester --serve 8088   # lalu buka http://127.0.0.1:8088
```

Isi `WST_FAKE_WIFI_SCAN` dengan file JSON berisi daftar jaringan (contohnya
`ui-tests/wifi-scan.json`) untuk mencoba Cek channel Wi-Fi tanpa Wi-Fi.

`scripts/ui-test.sh` melakukan semua itu secara otomatis dan menguji
halamannya dengan Playwright (`ui-tests/run.cjs`): pengaturan, Ping Cepat dan
tes lengkap, export, riwayat dan Laporan ISP, semua alat, kedua bahasa, tema
terang dan gelap, serta jendela yang sempit. Instal Playwright sekali dengan
`npm install --no-save playwright && npx playwright install chromium`.

## Arsitektur

Logika pengukuran tidak pernah ada di UI:

```
Tes jaringan (engine, measure, network)
  → Model hasil tes (model)
  → Analisis (analysis)
  → View-model / API (app)
  → UI (ui: jendela WebView2 atau dev server → web/app.js)
```

```
cmd/wifispeedtester   titik masuk: jendela, --cli, --serve
internal/network      deteksi adapter (GetAdaptersAddresses), info dan scan Wi-Fi (wlanapi), cek DNS/internet
internal/measure      ping (IcmpSendEcho), tes download/upload, hitungan jitter / packet loss / satuan
internal/engine       menjalankan tes lengkap, mengirim progres, pembatalan, pesan error untuk pengguna
internal/monitor      monitor stabilitas: ping tiap 1 detik, putus dan lonjakan
internal/dnsbench     perbandingan server DNS lewat UDP
internal/analysis     ConnectionQualityEngine: rating, catatan dan saran; kesimpulan DNS, channel Wi-Fi dan monitor
internal/i18n         teks Inggris / Indonesia untuk semua yang ditulis sisi Go
internal/servers      daftar server LibreSpeed dan pengurutan latensi
internal/update       cek rilis terbaru di GitHub
internal/export       TXT / JSON / CSV, Laporan ISP dan laporan monitor, dengan penyamaran data identitas
internal/history      riwayat lokal dalam JSON
internal/config       konfigurasi dan nilai bawaan
internal/app          view-model: API yang dipanggil UI; pekerjaan lama berjalan di luar thread UI
internal/ui           host WebView2, dialog Save, dev server, file web yang ditanam di exe
```

## Batasan

Ini aplikasi sederhana untuk mengetes koneksi. Aplikasi ini bukan platform
monitoring, penganalisis paket, alat admin router, alat bobol Wi-Fi atau
pemindai celah keamanan, dan memang tidak akan diarahkan ke sana. Monitor
stabilitas adalah pengecekan berwaktu yang kamu mulai dan pantau sendiri;
tidak ada mode latar belakang atau mode tray. Cek channel Wi-Fi hanya membaca
apa yang sudah terlihat oleh Windows.

## Lisensi

Hak cipta (c) 2026-sekarang tehgeii. Seluruh hak dilindungi.

Kamu boleh melihat kodenya, menjalankan rilis resmi untuk pemakaian pribadi,
dan membagikan link Releases atau file rilis yang tidak diubah secara gratis.
Menyalin atau memakai ulang kode, menyebarkan versi yang sudah diubah, serta
mengunggah ulang atau menjual aplikasinya harus dengan izin tertulis.
Ketentuan lengkapnya, dalam bahasa Inggris dan Bahasa Indonesia, ada di
[LICENSE](LICENSE). Versi 1.2.1 dan sebelumnya diterbitkan dengan Lisensi MIT.

Komponen dari pihak ketiga tetap memakai lisensinya masing-masing, yang
tercantum di [THIRD_PARTY_NOTICES.txt](THIRD_PARTY_NOTICES.txt).

## Pembuat

Dibuat oleh **[@tehgeii](https://github.com/tehgeii)**.
