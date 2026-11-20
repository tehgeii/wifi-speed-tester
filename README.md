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

Website: **https://tehgeii.github.io/wifi-speed-tester/**

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
| Hotspot HP | Wi-Fi dari hotspot HP dikenali sebagai **Hotspot / Tethering**: Android menandai hotspot-nya *berkuota* (metered), dan Windows meneruskan status itu (Windows 10 versi 2004 ke atas); iPhone dan Android lama dikenali dari alamat gateway-nya. Lewat HP, ping ke HP sampai 60 ms dianggap wajar, sarannya khusus HP (posisi HP, hotspot 5 GHz, USB tethering), Diagnosa langsung menyimpulkan "Lewat hotspot HP", hasilnya tidak dibandingkan dengan paket internet rumah, dan Hemat kuota disarankan. Hotspot yang tidak dikenali bisa ditandai sendiri: **Pengaturan Windows → Jaringan → Wi-Fi → (nama hotspot) → Koneksi berkuota** |
| Detail Wi-Fi | SSID, BSSID, sinyal %, RSSI, band, channel, frekuensi, standar Wi-Fi, link speed (diberi keterangan bahwa ini *bukan* kecepatan internet) |
| Cek internet | Adapter ✓ → Gateway ✓ → DNS ✓ → Internet ✓; berhenti dengan pesan yang jelas kalau internet tidak bisa dijangkau |
| Ping | Bawaannya ke router, Cloudflare DNS dan Google DNS, ditambah sampai 5 target pilihanmu sendiri (misalnya server game, di **Pengaturan ⚙**): terkirim/diterima, min/rata-rata/maks, jitter, loss |
| Download / Upload | Tes HTTP dengan beberapa koneksi sekaligus, masa pemanasan tidak dihitung, gauge langsung, grafik throughput, coba ulang dan pindah ke server cadangan kalau gagal |
| Ping saat sibuk | Ping diukur selama download dan upload, dengan petunjuk bufferbloat yang hati-hati (tidak langsung memvonis) |
| Rating | SANGAT BAGUS / BAGUS / CUKUP / BURUK / TIDAK STABIL, berdasarkan batas nilai yang terdokumentasi dan bisa diatur ([docs/QUALITY.md](docs/QUALITY.md)). **Kenapa rating-nya segini?** menampilkan nilai dan batas tiap metrik |
| Saran | Langkah nyata sesuai hasil yang terukur (nyalakan SQM/QoS, pakai 5 GHz, coba kabel LAN, restart router, VPN aktif, hubungi ISP…; lewat hotspot HP: posisi HP, hotspot 5 GHz, USB tethering) |
| Dibanding tes sebelumnya | Di bawah hasil: "Dibanding tes sebelumnya: Download +18% · Upload −5% · Ping +12 ms" terhadap tes terakhir dengan jenis koneksi dan jenis tes yang sama. Hijau lebih baik, merah lebih buruk, perubahan di bawah 5% ditulis "hampir sama". Juga muncul saat membuka hasil dari Riwayat |
| Cocok untuk apa? | Daftar aktivitas sehari-hari (browsing & chat, YouTube 1080p, YouTube/Netflix 4K, video call HD, game online, live streaming, download besar) dengan status **Lancar / Bisa, kadang tersendat / Kurang** dan alasannya. Daftarnya terisi **selama tes berjalan**, setelah tiap tahap, plus perkiraan waktu download game 50 GB |
| Nilai bufferbloat | Nilai **A+ sampai F** untuk seberapa besar ping naik saat koneksi sibuk, seperti di situs tes untuk gamer. Tampil di hasil, panel Gaming, Riwayat, dan kartu PNG |
| Paket ISP kamu | Isi kecepatan paket internetmu sekali saja (**Pengaturan ⚙**); hasil tes lalu menampilkan "62% dari paket", dan muncul saran kalau kecepatannya jauh di bawah paket |
| Mode Gaming | Menilai ping, jitter, loss, dan kestabilan ping, bukan besar bandwidth |
| Ping Cepat | Sekitar 5 detik: cek koneksi dan ping saja, untuk cek kilat sebelum main game |
| Hemat kuota | Penghitung **kuota terpakai yang naik langsung** selama tes, perkiraan kuota sebelum tes (dari tes terakhir di koneksi yang sama), total kuota tes per bulan di Riwayat, dan tombol **💾 Hemat kuota**: download/upload dites lebih singkat dan berhenti setelah sekitar 30 MB + 15 MB. Kalau aplikasi mendeteksi hotspot HP, data seluler, atau koneksi yang ditandai berkuota oleh Windows, muncul saran untuk menyalakannya |
| Ulangi 3× | Tombol **🔁 Ulangi 3×** menjalankan tes 3 kali berturut-turut lalu menampilkan nilai tengahnya (median) beserta tabel tiap putaran. Hasilnya lebih stabil dan jadi bukti yang lebih kuat untuk ISP. Disimpan sebagai satu entri riwayat |
| Server tes | Bawaannya Cloudflare; **Cari server terdekat** mengurutkan server LibreSpeed publik berdasarkan latensi, server pilihanmu dipakai duluan dan Cloudflare jadi cadangan |
| Bahasa | Bahasa Inggris dan Bahasa Indonesia (saat pertama dibuka mengikuti bahasa Windows; bisa diganti di **Pengaturan ⚙**) |
| Satuan | Mbps (bawaan) atau MB/s (80 Mbps = 10 MB/s) |
| Export | **Bagikan**: **Salin hasil** dan **Kartu PNG**, yang tidak pernah memuat nama Wi-Fi atau alamat IP. **Simpan sebagai file**: TXT, JSON dan CSV, yang menyamarkan data itu (`Ho*****`, `192.168.1.x`) kecuali kamu mencentang **Sertakan identitas jaringan**; keterangan di bawah tombol menunjukkan persis apa yang disembunyikan. TXT memuat nama Wi-Fi dan IP, JSON juga BSSID, CSV hanya nama Wi-Fi |
| Riwayat | Disimpan di komputer sendiri sebagai JSON di samping file exe; grafik tren kecepatan dan ping dari waktu ke waktu, rata-rata per jenis koneksi (Wi-Fi vs Ethernet…), % dari paket, filter 7/30 hari, buka ulang hasil, export ke CSV, hapus |
| Diagnosa otomatis | Tab **Diagnosa**: panduan langkah demi langkah untuk mencari tahu koneksi lambat karena **jangkauan Wi-Fi, Wi-Fi-nya, router, atau ISP**. Tes di posisi sekarang → tes dekat router (khusus Wi-Fi, dan dilewati otomatis kalau sinyal sudah kuat) → tes pakai kabel LAN (opsional), lalu kesimpulan yang jelas plus tombol ke alat yang cocok (Cek channel, Laporan ISP, Monitor). Langkah yang tidak perlu dilewati otomatis, dan tes diagnosa tidak masuk Riwayat |
| Laporan ISP | **Riwayat → Laporan ISP (PDF)** membuat laporan rapi untuk dikirim ke ISP: rata-rata, % dari paket, grafik kecepatan dan ping dari waktu ke waktu, rata-rata per jam, tes paling lambat, dan hasil monitor stabilitas. Isi yang sama juga tersedia sebagai **TXT**. Tidak memuat nama Wi-Fi atau alamat IP |
| Monitor stabilitas | **Alat** → ping ke satu target dan ke router setiap detik selama 5–30 menit untuk menangkap putus sesaat dan lonjakan lag, dengan grafik langsung, kesimpulan (stabil / ada lonjakan / putus; masalahnya di router atau setelahnya) dan laporan TXT. Hanya berjalan selama jendela aplikasi terbuka, dan berhenti saat kamu menekan Stop atau menutup aplikasi |
| Pencari sinyal Wi-Fi | **Alat** → angka sinyal besar (%, dBm) yang **diperbarui tiap detik**, grafik 60 detik terakhir, dan bunyi seperti detektor logam (opsional, makin kuat makin cepat). Bawa laptop berkeliling untuk mencari posisi terbaik, lalu **simpan titik** ("Kamar", "Ruang tamu") untuk melihat peringkatnya. Tidak memakai kuota sama sekali, dan berhenti sendiri saat kamu keluar dari alat ini |
| Ping ke region | **Alat** → latensi ke region cloud Jakarta, Singapura, Kuala Lumpur, Hong Kong, Tokyo, Mumbai, Sydney, Frankfurt dan California, muncul satu per satu begitu terukur, plus kesimpulan untuk game (sebagian besar server game Asia Tenggara ada di Singapura). Ini ping ke region-nya, bukan ke server game tertentu |
| Tes DNS | **Alat** → membandingkan seberapa cepat DNS kamu sekarang, Cloudflare, Google dan Quad9 menjawab, lalu bilang perlu ganti atau tidak |
| Cek channel Wi-Fi | **Alat** → menampilkan jaringan Wi-Fi di sekitar (hanya membaca, dari Windows) dan seberapa ramai tiap channel, plus saran channel untuk diatur di router |
| Salin hasil | Ringkasan singkat untuk WhatsApp, Discord atau di mana saja, tanpa nama Wi-Fi atau alamat IP |
| Update | Cek versi baru saat aplikasi dibuka (bisa dimatikan di **Pengaturan ⚙**). Mulai 2.1: tombol **Update sekarang** mengunduh versi baru dari rilis GitHub, mencocokkan SHA-256-nya dengan `SHA256SUMS.txt`, lalu mengganti exe; **Mulai ulang sekarang** membuka versi barunya. Riwayat, pengaturan, dan `WiFiSpeedTester.config.json` tidak disentuh. Kalau gagal (misalnya folder tidak bisa ditulisi), program lama tetap utuh dan link halaman download tetap ada |
| Error | Penyebab dan saran dengan bahasa sederhana, detail teknisnya bisa dibuka kalau perlu |
| Laporkan masalah | Di dialog **Tentang** dan di kartu error: **Buka GitHub Issues** dengan info aplikasi yang sudah terisi (versi aplikasi dan Windows, jenis koneksi, error terakhir), atau **Salin info** untuk dikirim lewat WhatsApp/email. Infonya ditampilkan dulu sebelum dikirim, dan tidak memuat alamat IP, nama Wi-Fi, atau nama perangkat |
| Batal | Menghentikan semua aktivitas jaringan saat itu juga (tombol Esc juga bisa) |
| Satu jendela | Kalau exe-nya dibuka lagi, jendela yang sudah terbuka langsung muncul ke depan, bukan membuka jendela kedua |
| Tampilan | Navigasi tab **Tes · Diagnosa · Riwayat · Alat** (juga lewat Ctrl+1/2/3/4), tema **Terang / Gelap / Ikuti Windows** di **Pengaturan ⚙** (title bar jendela ikut berubah), ukuran dan posisi jendela diingat, dan tombol taskbar berkedip saat tes, diagnosa atau monitor selesai sementara kamu membuka aplikasi lain |

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
- hanya saat kamu menjalankan **Ping ke region**: 5 koneksi TCP singkat tanpa
  data ke tiap region (bawaannya `dynamodb.<region>.amazonaws.com` port 443,
  milik Amazon Web Services);
- **Diagnosa otomatis** menjalankan speed test biasa ke server yang sama;
- hanya saat kamu menekan **Cari server terdekat**: daftar server publik dari
  `librespeed.org` dan satu permintaan kecil ke tiap server di daftar itu;
- hanya saat kamu menekan **Buka GitHub Issues** di **Laporkan masalah**:
  halaman issue baru di github.com terbuka di browser, berisi info yang sudah
  kamu lihat di dialog;
- hanya kalau **Cek update saat aplikasi dibuka** menyala (**Pengaturan ⚙**,
  menyala secara bawaan): satu permintaan ke `api.github.com` untuk melihat
  versi terbaru;
- hanya saat kamu menekan **Update sekarang**: unduhan zip versi baru dan
  `SHA256SUMS.txt` dari halaman rilis proyek ini di `github.com`.
  Checksum-nya melindungi dari unduhan yang rusak atau tidak lengkap; karena
  kedua file berasal dari rilis yang sama, ini bukan tanda tangan digital.

**Cek channel Wi-Fi** dan **Pencari sinyal Wi-Fi** tidak mengirim apa pun:
keduanya hanya membaca jaringan dan sinyal yang memang sudah terlihat oleh
Windows. Aplikasi ini tidak membaca
file, riwayat browser atau password, tidak punya telemetri, dan tidak berjalan
di latar belakang. Dialog **Tentang** (tombol `?`) menampilkan ke mana saja
lalu lintas dikirim sesuai pengaturan yang sedang dipakai.

File yang dibuat aplikasi:

- Folder `WiFiSpeedTester-data\` di samping exe berisi `history.json`,
  `settings.json` (bahasa, satuan, mode, tema, cek update, kecepatan paket),
  `window.json` (ukuran dan posisi jendela),
  `server.json` (server yang dipilih di aplikasi, kalau ada), `targets.json`
  (target ping tambahanmu), `monitors.json` (20 ringkasan monitor
  stabilitas terakhir) dan `spots.json` (titik sinyal yang kamu simpan,
  maksimal 20). Kalau folder itu tidak bisa ditulisi, aplikasi
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
- **Pencari sinyal Wi-Fi**: aplikasi membaca sinyal tiap detik, tapi Windows
  sendiri memperbarui nilainya kira-kira tiap 1–5 detik tergantung kartu
  Wi-Fi, jadi angkanya kadang tetap selama beberapa detik. Titik yang disimpan
  memakai rata-rata 5 bacaan terakhir. Sinyal 60% ke atas disebut kuat, 40–59%
  cukup, di bawah 40% lemah.
- **Ping ke region**: waktu membuka koneksi TCP (5 kali, diambil mediannya),
  sama seperti cadangan ping saat ICMP diblokir. Angkanya mendekati ping ke
  server game di region itu, tapi tiap game punya server dan rute sendiri.
- **Diagnosa otomatis**: aturan kesimpulannya ada di
  [docs/QUALITY.md](docs/QUALITY.md#automatic-diagnosis).
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
  "regions": [                       // Ping ke region (port 443 kalau tidak ditulis host:port), maks. 20
    { "label": "Jakarta",   "host": "dynamodb.ap-southeast-3.amazonaws.com" },
    { "label": "Singapore", "host": "dynamodb.ap-southeast-1.amazonaws.com" }
    // … Kuala Lumpur, Hong Kong, Tokyo, Mumbai, Sydney, Frankfurt, California
  ],
  "planLowPct": 50,                  // di bawah % paket ini, muncul saran yang bisa dilakukan
  "dataSaverSeconds": 5,             // Hemat kuota: lama maksimal tiap arah (detik)
  "dataSaverDownloadMB": 30, "dataSaverUploadMB": 15,   // dan batas kuota tiap arah
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
WiFiSpeedTester.exe --cli --repeat 3  jalankan tes 3 kali dan tampilkan nilai tengahnya
WiFiSpeedTester.exe --cli --data-saver  tes hemat kuota (lebih singkat, dibatasi kuotanya)
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
dan Windows (isi Laporan ISP PDF dicek dengan `pdftotext`), menjalankan tes UI
di Chromium, dan membuat kedua zip sebagai artifact. CI juga menjalankan smoke
test sungguhan untuk `--cli`, `--dns` dan `--wifi-scan` di runner Windows, dan
menguji jendela exe yang asli lewat port DevTools WebView2
(`ui-tests/windows.cjs`): bridge Go, tab, tema, Ping Cepat, ping ke region, dan
kode Wi-Fi Windows.

Website di folder `site/` diterbitkan ke GitHub Pages oleh
`.github/workflows/pages.yml` setiap ada perubahan di `main`. Screenshot-nya
diambil dari tes UI, jadi selalu sesuai tampilan terbaru. Sekali saja, di
repository: **Settings → Pages → Source: GitHub Actions**.

Untuk menerbitkan rilis: buka **Actions → build → Run workflow**, pilih
`main`, lalu isi versinya, misalnya `1.0.1`. Workflow akan membuat kedua zip
dan `SHA256SUMS.txt`, membuat tag `v1.0.1`, lalu menerbitkan rilis dengan
catatan dari `packaging/RELEASE_NOTES.md` (perbarui dulu bagian "Yang baru di …"-nya).
Push tag `v*` juga menghasilkan hal yang sama.

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
`ui-tests/wifi-scan.json`) untuk mencoba Cek channel Wi-Fi tanpa Wi-Fi, dan
`WST_FAKE_WIFI_SIGNAL` dengan daftar bacaan sinyal (contohnya
`ui-tests/wifi-signal.json`, dipakai bergiliran tiap detik) untuk mencoba
Pencari sinyal Wi-Fi.

`scripts/ui-test.sh` melakukan semua itu secara otomatis dan menguji
halamannya dengan Playwright (`ui-tests/run.cjs`): pengaturan, Ping Cepat dan
tes lengkap, export, riwayat dan Laporan ISP, diagnosa, semua alat, kedua bahasa, tema
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
internal/regions      ping ke region cloud lewat waktu koneksi TCP
internal/analysis     ConnectionQualityEngine: rating, catatan dan saran; kesimpulan DNS, channel Wi-Fi, monitor, region dan diagnosa
internal/i18n         teks Inggris / Indonesia untuk semua yang ditulis sisi Go
internal/servers      daftar server LibreSpeed dan pengurutan latensi
internal/update       cek rilis terbaru di GitHub, Update sekarang (unduh, cek SHA-256, ganti exe)
internal/export       TXT / JSON / CSV, Laporan ISP (TXT dan PDF, penulis PDF sendiri tanpa library) dan laporan monitor, dengan penyamaran data identitas
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
tidak ada mode latar belakang atau mode tray. Cek channel Wi-Fi dan Pencari
sinyal Wi-Fi hanya membaca apa yang sudah terlihat oleh Windows. Diagnosa
otomatis hanya membandingkan hasil tes; aplikasi tidak pernah mengubah
pengaturan router.

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
