## Yang baru di 2.1

- **Hotspot HP dikenali dengan benar.** Hotspot Android sekarang memakai alamat acak (misalnya 10.54.x.x), jadi dulu terbaca sebagai Wi-Fi rumah. Sekarang aplikasi membaca status *berkuota* yang dikirim HP ke Windows, lalu menampilkannya sebagai **Hotspot / Tethering**. Hasilnya:
  - ping ke HP sampai 60 ms dianggap wajar (HP menghemat daya di antara paket), jadi tidak ada lagi "router bermasalah, restart router";
  - saran khusus HP: taruh HP di tempat sinyal 4G/5G kuat, ubah hotspot ke 5 GHz, pakai USB tethering supaya ping stabil;
  - **Diagnosa** langsung menyimpulkan "Lewat hotspot HP" setelah langkah 1, tanpa menyuruh jalan ke router;
  - **Hemat kuota** disarankan, dan hasilnya tidak dibandingkan dengan paket internet rumah;
  - Riwayat dan Laporan ISP memisahkan tes hotspot dari Wi-Fi rumah.

  Kalau hotspot-mu masih terbaca Wi-Fi, tandai sendiri di **Pengaturan Windows → Jaringan → Wi-Fi → (nama hotspot) → Koneksi berkuota**.
- **Dibanding tes sebelumnya**: di bawah hasil ada baris seperti "Download +18% · Upload −5% · Ping +12 ms" terhadap tes terakhir di koneksi yang sama. Hijau lebih baik, merah lebih buruk.
- **Update sekarang**: kalau ada versi baru, cukup klik **Update sekarang** lalu **Mulai ulang sekarang**. Aplikasi mengunduh zip dari rilis GitHub, mencocokkan checksum-nya, lalu mengganti exe; riwayat dan pengaturanmu tetap. Fitur ini dipakai mulai update **dari 2.1 ke versi berikutnya**; versi 2.1 sendiri di-download manual seperti biasa.

## Yang baru di 2.0.1

Perbaikan kecil hasil tes di laptop asli:

- **Diagnosa** tidak lagi menyuruh kamu jalan ke dekat router kalau sinyal Wi-Fi di posisimu sudah kuat (70% ke atas), karena pindah lebih dekat tidak akan mengubah hasilnya. Langsung ke langkah kabel (opsional), dan kesimpulannya menjelaskan kenapa.
- Hasil tiap langkah diagnosa sekarang berwarna sesuai kualitasnya: hijau untuk bagus, kuning untuk cukup, merah untuk buruk. Sebelumnya selalu hijau.
- Kartu **Cocok untuk apa?** tetap rapi saat jendela dimaksimalkan: statusnya ada di baris sendiri, di bawah nama aktivitas.
- Daftar cek tidak lagi menulis "Wi-Fi — Wi-Fi" kalau nama adapter sama dengan jenis koneksinya.

## Yang baru di 2.0

Versi besar: tampilannya dirombak dan ada banyak alat baru. Semuanya tetap portable, tanpa install, dan tidak ada yang berjalan di latar belakang.

### Tampilan
- **Navigasi tab baru**: **Tes · Diagnosa · Riwayat · Alat** (juga lewat Ctrl+1/2/3/4).
- **Tema Terang / Gelap / Ikuti Windows** di **Pengaturan ⚙**; title bar jendela ikut berubah.
- **Ukuran dan posisi jendela diingat**, dan tombol taskbar berkedip saat tes, diagnosa, atau monitor selesai sementara kamu membuka aplikasi lain.

### Hasil tes yang lebih mudah dipahami
- **Cocok untuk apa?**: browsing, YouTube 1080p, Netflix 4K, video call, game online, live streaming, dan download besar, masing-masing dengan status **Lancar / Bisa, kadang tersendat / Kurang** beserta alasannya. Daftarnya terisi **selama tes berjalan**.
- **Nilai bufferbloat A+ sampai F**: seberapa naik ping saat koneksi sedang dipakai banyak, seperti di situs tes untuk gamer.

### Kuota dan akurasi
- **Penghitung kuota langsung** selama tes, perkiraan kuota sebelum tes, dan total kuota tes per bulan di Riwayat.
- **💾 Hemat kuota**: tes lebih singkat dan dibatasi sekitar 45 MB. Kalau kamu memakai hotspot HP, aplikasi menyarankannya.
- **🔁 Ulangi 3×**: tes 3 kali berturut-turut, lalu hasilnya diambil dari nilai tengah (median).

### Alat baru
- **🩺 Diagnosa otomatis**: tes di posisi sekarang, dekat router, lalu pakai kabel. Aplikasi menyimpulkan apakah masalahnya di jangkauan Wi-Fi, Wi-Fi-nya, router, atau ISP.
- **📡 Pencari sinyal Wi-Fi**: sinyal diperbarui tiap detik, bisa dengan bunyi seperti detektor logam, dan titik-titik di rumah bisa disimpan untuk melihat mana yang terbaik.
- **🗺️ Ping ke region**: latensi ke Jakarta, Singapura, Kuala Lumpur, Hong Kong, Tokyo, Mumbai, Sydney, Frankfurt, dan California, plus kesimpulan untuk server game Asia Tenggara.

### Laporan dan bantuan
- **📄 Laporan ISP dalam PDF**: lengkap dengan grafik kecepatan dan ping, % dari paket, dan jam-jam lambat. Versi TXT tetap ada.
- **Laporkan masalah** (dialog **Tentang** dan kartu error): buka GitHub Issues yang sudah terisi info aplikasi, atau salin infonya untuk dikirim lewat WhatsApp/email. Info ini tidak memuat alamat IP atau nama Wi-Fi.
- **Website**: https://tehgeii.github.io/wifi-speed-tester/

### Privasi
Lalu lintas baru hanya terjadi saat kamu memakai alatnya: **Ping ke region** membuka beberapa koneksi TCP tanpa data ke server region Amazon Web Services, dan **Laporkan masalah** membuka GitHub di browser. Pencari sinyal Wi-Fi tidak mengirim apa pun. Daftar lengkapnya ada di dialog **Tentang**.

## Download

| PC kamu | File |
|---|---|
| Kebanyakan PC/laptop Windows 10/11 (Intel / AMD) | **`WiFiSpeedTester-Portable-…-win-x64.zip`** |
| Windows on ARM (Snapdragon) | `WiFiSpeedTester-Portable-…-win-arm64.zip` |

1. Download zip-nya, lalu ekstrak di mana saja.
2. Klik dua kali **`WiFiSpeedTester.exe`**. Tidak ada yang perlu diinstal.
3. Kalau Windows menampilkan **"Windows protected your PC"**, klik **More info → Run anyway** (di Windows berbahasa Indonesia: **Info selengkapnya → Tetap jalankan**). Peringatan SmartScreen ini muncul karena aplikasinya belum punya tanda tangan digital (code signing).

Aplikasi ini membutuhkan Microsoft Edge WebView2 Runtime, yang sudah ada di Windows 11 dan kebanyakan Windows 10. Kalau belum ada, aplikasi akan menawarkan link download-nya.

`SHA256SUMS.txt` berisi checksum untuk memastikan file yang kamu download asli dan tidak rusak (`certutil -hashfile <zip> SHA256`).
