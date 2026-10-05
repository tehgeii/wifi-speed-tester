# WiFi Speed + Ping Tester

> **A lightweight Windows utility for quickly measuring internet speed, latency, jitter, packet loss, and basic connection quality.**

WiFi Speed + Ping Tester adalah aplikasi desktop Windows portable yang dirancang untuk membantu pengguna mengetahui **seberapa cepat dan stabil koneksi internet mereka saat ini**.

Aplikasi berfokus pada dua hal utama:

1. **Speed Testing**
2. **Latency / Ping Testing**

Tujuan aplikasi adalah membuat pengujian internet yang mudah dipahami oleh pengguna awam tanpa harus membuka website speed test yang penuh elemen tambahan.

---

# 1. Product Definition

Aplikasi harus memungkinkan pengguna melakukan:

```text
Internet Connection
        ↓
Detect Network
        ↓
Measure Latency
        ↓
Measure Download
        ↓
Measure Upload
        ↓
Measure Jitter
        ↓
Measure Packet Loss
        ↓
Generate Connection Summary
```

Hasil akhirnya harus mudah dipahami:

```text
Download
48.6 Mbps

Upload
11.2 Mbps

Ping
18 ms

Jitter
3 ms

Packet Loss
0%

Connection
GOOD
```

---

# 2. Apa yang Ingin Diselesaikan?

Banyak pengguna hanya mengetahui:

> "Internetku lambat."

Tetapi tidak mengetahui apakah masalahnya:

- download speed rendah
- upload speed rendah
- ping tinggi
- jitter tinggi
- packet loss
- Wi-Fi tidak stabil
- server tujuan bermasalah
- koneksi ISP bermasalah
- local network bermasalah

Aplikasi ini bertujuan memberikan **pengukuran dasar dan ringkasan sederhana**, bukan sekadar angka.

---

# 3. Positioning

Aplikasi ini adalah:

> **Simple Internet Connection Testing Utility**

Bukan:

- ISP monitoring platform
- enterprise network monitoring
- advanced packet analyzer
- Wi-Fi hacking tool
- router configuration tool
- speed guarantee tool
- professional network analyzer

---

# 4. Target Users

### Casual User

Ingin tahu:

> "Internet saya cepat atau tidak?"

### Gamer

Ingin mengetahui:

- ping
- jitter
- packet loss
- latency stability

### Student

Untuk tugas jaringan / praktikum.

### Technician

Untuk quick troubleshooting.

### Home User

Untuk membandingkan koneksi:

- Wi-Fi
- Ethernet
- hotspot

---

# 5. Portable Distribution

## Primary Recommendation

Aplikasi harus didistribusikan sebagai portable application.

Contoh:

```text
WiFiSpeedTester-Portable.zip
└── WiFiSpeedTester.exe
```

User:

```text
Download
↓
Extract
↓
Run
↓
Test
```

Tidak membutuhkan:

- installer
- registry installation
- background service
- startup entry

---

# 6. Kenapa Portable?

Utility seperti ini sering digunakan:

- sekali-sekali
- untuk troubleshooting
- di laptop teman
- untuk mengecek Wi-Fi
- untuk membandingkan koneksi
- oleh teknisi

Karena itu friction harus seminimal mungkin.

User cukup mendapatkan:

> `WiFiSpeedTester.exe`

---

# 7. Installer Future Option

Installer dapat ditambahkan kemudian.

Contoh:

```text
WiFiSpeedTester-Setup.exe
```

Installer baru bermanfaat jika ditambahkan:

- automatic monitoring
- system tray
- startup
- scheduled tests
- test history database
- notification
- auto diagnostics
- Windows integration

Untuk MVP:

> Portable is preferred.

---

# 8. Main User Flow

```text
Launch
  ↓
Detect Network
  ↓
Select Test Mode
  ↓
Latency Test
  ↓
Download Test
  ↓
Upload Test
  ↓
Jitter / Packet Loss Analysis
  ↓
Generate Result
```

---

# 9. Dashboard

Contoh:

```text
WIFI SPEED + PING TESTER

Connection
Wi-Fi

Network
Home Wi-Fi

────────────────────

DOWNLOAD
48.6 Mbps

UPLOAD
11.2 Mbps

PING
18 ms

JITTER
3 ms

PACKET LOSS
0%

────────────────────

Connection Quality
GOOD

[ START TEST ]

[ DETAILS ]
```

Dashboard harus sederhana.

---

# 10. Network Detection

Aplikasi dapat mendeteksi:

- active network interface
- Wi-Fi
- Ethernet
- hotspot/tethering
- adapter name
- local IP
- gateway
- DNS
- IPv4
- IPv6 jika tersedia

Contoh:

```text
Connection:
Wi-Fi

Adapter:
Intel Wi-Fi 6 AX201

IPv4:
192.168.1.23

Gateway:
192.168.1.1
```

Informasi sensitif harus ditampilkan secara aman apabila diekspor.

---

# 11. Wi-Fi Specific Information

Jika tersedia melalui Windows API:

- SSID
- signal strength
- link speed
- channel
- band
- adapter
- connection type

Contoh:

```text
Wi-Fi

SSID:
Home_5G

Signal:
88%

Band:
5 GHz

Link Speed:
866 Mbps
```

Link speed bukan internet speed.

Aplikasi harus membedakan:

> **Wi-Fi Link Speed**

dengan:

> **Internet Download Speed**

---

# 12. Internet Connectivity Test

Sebelum speed test:

```text
Internet Check

Network adapter ✓
Gateway ✓
DNS ✓
Internet ✓
```

Jika internet tidak tersedia:

```text
Internet
✗ Unavailable

Speed test cannot continue.
```

---

# 13. Ping Test

Ping adalah komponen inti.

Aplikasi harus dapat mengukur:

- minimum latency
- average latency
- maximum latency
- packets sent
- packets received
- packet loss

Contoh:

```text
Packets:
20 sent
20 received

Min:
15 ms

Average:
18 ms

Max:
27 ms

Packet Loss:
0%
```

---

# 14. Multiple Ping Targets

Jangan hanya menggunakan satu target.

Contoh konsep:

```text
Local Gateway
Internet DNS
Regional Server
```

Contoh:

```text
Gateway
2 ms

Public Internet
18 ms

Regional Endpoint
22 ms
```

Tujuannya membantu membedakan:

```text
Local network issue
```

vs

```text
Internet latency issue
```

Endpoint aktual harus configurable.

Jangan mengandalkan satu server selamanya.

---

# 15. Download Speed Test

Download test harus menggunakan test resource yang cukup besar sehingga hasil tidak terlalu dipengaruhi oleh startup overhead.

Hasil:

```text
Download
48.6 Mbps
```

Gunakan pendekatan streaming/download sehingga pengukuran dapat dilakukan secara bertahap.

Perlu dipertimbangkan:

- test duration
- test payload size
- connection warm-up
- measurement interval
- cancellation
- server response
- retry

---

# 16. Upload Speed Test

Upload harus menggunakan endpoint yang benar-benar menerima data untuk testing.

Hasil:

```text
Upload
11.2 Mbps
```

Jangan membuat fake upload measurement berdasarkan local processing.

Upload membutuhkan destination endpoint yang mendukung upload testing.

---

# 17. Speed Test Accuracy

Aplikasi harus jujur mengenai batas pengukuran.

Hasil speed test dipengaruhi:

- test server
- routing
- ISP
- Wi-Fi signal
- concurrent traffic
- VPN
- browser/app activity
- server congestion
- device limitations

Karena itu jangan tampilkan:

> "Your ISP speed is exactly 50 Mbps"

Lebih baik:

> "Measured download speed: 48.6 Mbps"

---

# 18. Unit

Support:

- Mbps
- MB/s

Default:

> Mbps

Konversi harus benar.

Contoh:

```text
80 Mbps ≈ 10 MB/s
```

Jangan membingungkan:

```text
Mb
```

dengan:

```text
MB
```

---

# 19. Ping During Speed Test

Optional advanced feature:

Pantau ping selama download/upload.

Tujuan:

> Apakah latency meningkat drastis saat koneksi dipakai penuh?

Contoh:

```text
Idle Ping
18 ms

During Download
74 ms

During Upload
91 ms
```

Aplikasi dapat memberikan:

> "Latency increases significantly under load."

Ini dapat menjadi indikator awal bufferbloat, tetapi jangan menyatakan diagnosis absolut tanpa pengujian yang sesuai.

---

# 20. Jitter

Jitter menunjukkan perubahan latency antar-paket.

Contoh:

```text
Ping:
18
19
17
21
18
20

Jitter:
3 ms
```

Jangan menjelaskan jitter sebagai:

> "kecepatan internet."

Jitter adalah **stability of latency**.

---

# 21. Packet Loss

Packet loss:

```text
0%
```

bagus untuk basic connectivity.

Contoh:

```text
Packets:
100 sent
100 received

Packet Loss:
0%
```

Jika:

```text
100 sent
96 received
```

hasil:

```text
Packet Loss
4%
```

---

# 22. Connection Quality

Buat klasifikasi sederhana.

Contoh:

```text
EXCELLENT
GOOD
FAIR
POOR
UNSTABLE
```

Kriteria harus deterministic.

Contoh konsep:

```text
Excellent
Low latency
Low jitter
No packet loss
Strong speed

Good
Minor latency/jitter

Fair
Moderate issues

Poor
High latency or low speed

Unstable
Packet loss / large jitter
```

Threshold harus configurable dan didokumentasikan.

---

# 23. Gamer Mode

Optional.

User memilih:

```text
[ General ]
[ Gaming ]
```

Gaming mode memprioritaskan:

- ping
- jitter
- packet loss
- latency under load

Contoh:

```text
Gaming Connection

Ping
18 ms ✓

Jitter
3 ms ✓

Packet Loss
0% ✓

Latency Stability
Excellent
```

Jangan menjanjikan:

> "Your game will have 18 FPS ping."

Ping bukan FPS.

---

# 24. Test Progress

Saat testing:

```text
CHECKING CONNECTION...

✓ Network detected
✓ Gateway reachable
✓ Internet reachable

PING TEST
████████░░ 80%

DOWNLOAD
████░░░░░░ 40%

UPLOAD
░░░░░░░░░░ 0%
```

UI tidak boleh freeze.

---

# 25. Cancel Test

User harus dapat membatalkan.

```text
[ CANCEL ]
```

Cancellation harus:

- stop network operation
- release resources
- restore UI
- not crash

---

# 26. Test History

Optional MVP+ feature.

Simpan:

```text
Date
Connection
Download
Upload
Ping
Jitter
Packet Loss
```

Contoh:

```text
Oct 04
48.6 / 11.2 Mbps
18 ms
0%

Oct 03
42.1 / 9.8 Mbps
23 ms
1%
```

Untuk portable mode:

- local JSON
- SQLite
- local database

jika diperlukan.

---

# 27. Export Result

Support:

### TXT

### JSON

### CSV

Optional:

### PNG

Contoh report:

```text
WIFI SPEED + PING TEST

Date:
04 October 2026

Connection:
Wi-Fi

SSID:
Home_5G

Download:
48.6 Mbps

Upload:
11.2 Mbps

Ping:
18 ms

Jitter:
3 ms

Packet Loss:
0%

Quality:
GOOD
```

---

# 28. Shareable Result Card

Future feature.

Contoh:

```text
┌─────────────────────────────┐
│ INTERNET TEST               │
│                             │
│ ↓ 48.6 Mbps                 │
│ ↑ 11.2 Mbps                 │
│ ● 18 ms                     │
│                             │
│ Jitter 3 ms                 │
│ Loss 0%                     │
│                             │
│ CONNECTION: GOOD            │
└─────────────────────────────┘
```

Export PNG.

---

# 29. Privacy

Default:

> Testing should be local except for traffic required to measure internet connectivity.

The application must not:

- upload personal files
- inspect browser history
- inspect passwords
- scan arbitrary files
- collect unrelated information

If testing uses external servers, explain this to the user.

---

# 30. No Background Monitoring

MVP tidak membutuhkan:

- tray application
- background monitoring
- auto-start
- scheduled tests

This keeps portable utility lightweight.

---

# 31. Technical Architecture

Suggested:

```text
WiFiSpeedTester
│
├── UI
│
├── Network
│   ├── AdapterDetector
│   ├── WifiInfoProvider
│   ├── GatewayChecker
│   └── InternetChecker
│
├── Testing
│   ├── PingTester
│   ├── DownloadTester
│   ├── UploadTester
│   ├── JitterCalculator
│   └── PacketLossCalculator
│
├── Analysis
│   └── ConnectionQualityEngine
│
├── Export
│
├── History
│
└── Infrastructure
    ├── Logging
    └── Configuration
```

---

# 32. Important Technical Principle

Jangan menempatkan network measurement logic di UI.

Gunakan:

```text
Network Test
↓
Test Result Model
↓
Analysis
↓
ViewModel
↓
UI
```

---

# 33. Error Handling

Contoh:

```text
Speed Test Failed

Reason:
Test server did not respond.

Suggestion:
Try again or select another test server.
```

Jangan:

```text
Error 500 Internal Server Error
```

saja.

User harus tahu apa yang bisa dilakukan.

---

# 34. Server Strategy

Jangan membuat arsitektur yang hanya bergantung pada satu endpoint.

Pertimbangkan:

```text
Primary Test Server
Backup Test Server
```

atau configurable endpoint.

Pada fase awal, gunakan endpoint yang stabil dan legal untuk testing.

Aplikasi harus menangani:

- timeout
- connection reset
- HTTP error
- unavailable server
- slow server
- cancellation

---

# 35. MVP

MVP minimal:

```text
✓ Network detection
✓ Wi-Fi / Ethernet detection
✓ Internet connectivity
✓ Ping
✓ Average ping
✓ Packet loss
✓ Basic jitter
✓ Download test
✓ Upload test
✓ Connection rating
✓ Dashboard
✓ Portable EXE
✓ TXT/JSON export
```

Future:

```text
Wi-Fi details
History
PNG card
Gaming mode
Latency under load
Multiple servers
AI explanation
```

---

# 36. What It Should NOT Become

Jangan mengubahnya menjadi:

- Wireshark clone
- enterprise monitoring
- packet analyzer
- Wi-Fi hacking software
- router admin panel
- vulnerability scanner

Keep it:

> Fast. Simple. Reliable.

---

# 37. Definition of Done

A release dianggap selesai apabila:

```text
Network detection works
+
Speed measurement works
+
Ping measurement works
+
Packet loss calculation works
+
Jitter calculation works
+
UI does not freeze
+
Cancellation works
+
Errors are understandable
+
Export works
+
Portable EXE works
```

---

# 38. Final Product Statement

> **WiFi Speed + Ping Tester lets users quickly measure how fast and stable their internet connection really is.**

Short description:

> **Check your speed. Measure your ping. Understand your connection.**

**End of Product Brief**