## Yang baru di 1.3

- **Lisensi baru.** Aplikasi tetap gratis untuk di-download, dipakai, dan dibagikan apa adanya. Menyalin kode, menyebarkan versi yang sudah diubah, atau menjual aplikasinya sekarang perlu izin dari pembuatnya. Ketentuan lengkapnya, dalam bahasa Inggris dan Bahasa Indonesia, ada di `LICENSE.txt`. Versi 1.2.1 dan sebelumnya tetap memakai Lisensi MIT.
- Zip sekarang menyertakan `THIRD_PARTY_NOTICES.txt`, berisi daftar komponen dari pembuat lain beserta lisensinya.
- Hasil tes, cara mengukur, dan fiturnya sama persis dengan 1.2.1.

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
