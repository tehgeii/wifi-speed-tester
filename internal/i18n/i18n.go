// Package i18n holds the user-facing text produced by the Go side (rating
// notes, tips, check details, error explanations, reports) in English and
// Indonesian. The web UI has its own table for its static labels.
package i18n

import (
	"fmt"
	"strings"
)

// Lang is a supported language.
type Lang string

const (
	EN Lang = "en"
	ID Lang = "id"
)

// Parse maps a language tag such as "id", "id-ID" or "en-US" to a Lang,
// defaulting to English.
func Parse(s string) Lang {
	if strings.HasPrefix(strings.ToLower(s), "id") || strings.HasPrefix(strings.ToLower(s), "in") {
		return ID
	}
	return EN
}

// T formats the message key in l. Unknown keys return the key itself so a
// missing translation is visible instead of silent.
func T(l Lang, key string, args ...any) string {
	m, ok := catalog[key]
	if !ok {
		return key
	}
	f := m[0]
	if l == ID && m[1] != "" {
		f = m[1]
	}
	if len(args) == 0 {
		return f
	}
	return fmt.Sprintf(f, args...)
}

// catalog maps a key to its {English, Indonesian} text.
var catalog = map[string][2]string{
	// Metric labels
	"metric.ping":          {"Ping", "Ping"},
	"metric.jitter":        {"Jitter", "Jitter"},
	"metric.packetLoss":    {"Packet Loss", "Packet Loss"},
	"metric.download":      {"Download", "Download"},
	"metric.upload":        {"Upload", "Upload"},
	"metric.loadedLatency": {"Latency Under Load", "Ping Saat Sibuk"},

	"level.EXCELLENT": {"EXCELLENT", "SANGAT BAGUS"},
	"level.GOOD":      {"GOOD", "BAGUS"},
	"level.FAIR":      {"FAIR", "CUKUP"},
	"level.POOR":      {"POOR", "BURUK"},
	"level.UNSTABLE":  {"UNSTABLE", "TIDAK STABIL"},
	"level.UNKNOWN":   {"UNKNOWN", "TIDAK DIKETAHUI"},

	"profile.general": {"General", "Umum"},
	"profile.gaming":  {"Gaming", "Gaming"},
	"profile.quick":   {"Quick Ping", "Ping Cepat"},

	// Summary
	"summary.head":       {"Connection quality: %s (%s profile).", "Kualitas koneksi: %s (profil %s)."},
	"summary.download":   {"measured download %s", "download terukur %s"},
	"summary.upload":     {"upload %s", "upload %s"},
	"summary.ping":       {"ping %s", "ping %s"},
	"summary.jitter":     {"jitter %s", "jitter %s"},
	"summary.loss":       {"packet loss %s", "packet loss %s"},
	"summary.disclaimer": {"Results reflect this moment and this test server, not a guaranteed ISP speed.", "Hasil ini mencerminkan kondisi saat ini dan server tes ini, bukan jaminan kecepatan dari ISP."},

	// Notes (observations)
	"note.pingMissing":     {"Ping could not be measured (ICMP may be blocked on this network).", "Ping tidak bisa diukur (ICMP mungkin diblokir di jaringan ini)."},
	"note.downloadMissing": {"Download was not measured; the rating uses the remaining results.", "Download tidak terukur; rating memakai hasil lainnya."},
	"note.uploadMissing":   {"Upload was not measured; the rating uses the remaining results.", "Upload tidak terukur; rating memakai hasil lainnya."},
	"note.bufferbloat":     {"Latency increases significantly under load. This can be an early sign of bufferbloat; a dedicated test is needed to confirm.", "Ping naik cukup tinggi saat koneksi dipakai penuh. Ini bisa jadi tanda awal bufferbloat; perlu tes khusus untuk memastikannya."},
	"note.unstable":        {"Packet loss or large latency variation makes this connection unstable.", "Packet loss atau ping yang naik-turun membuat koneksi ini tidak stabil."},
	"note.tooFew":          {"Too few results were measured to rate this connection reliably.", "Terlalu sedikit hasil yang terukur untuk memberi rating yang bisa diandalkan."},
	"note.tcpPing":         {"This network blocks ICMP ping, so latency was measured with TCP connections to the test server. These values can read slightly higher than a normal ping.", "Jaringan ini memblokir ping ICMP, jadi ping diukur lewat koneksi TCP ke server tes. Nilainya bisa sedikit lebih tinggi dari ping biasa."},
	"note.routerNoPing":    {"The router did not answer ping. Many routers block ping, so this alone is not a fault.", "Router tidak membalas ping. Banyak router memang memblokir ping, jadi ini saja belum tentu masalah."},
	"note.localProblem":    {"Delay or loss already appears between this device and the router (%s avg, %s loss). The local network or Wi-Fi is a likely cause.", "Delay atau loss sudah muncul antara perangkat ini dan router (rata-rata %s, loss %s). Kemungkinan penyebabnya jaringan lokal atau Wi-Fi."},
	"note.beyondRouter":    {"The local network looks healthy; the extra latency or loss is added beyond the router (ISP or internet route).", "Jaringan lokal terlihat sehat; tambahan ping atau loss terjadi setelah router (ISP atau rute internet)."},
	"note.weakSignal":      {"Wi-Fi signal is weak (%d%%).", "Sinyal Wi-Fi lemah (%d%%)."},
	"note.linkSpeed":       {"Wi-Fi link speed (%.0f Mbps) is the radio rate to the router, not your internet speed.", "Wi-Fi link speed (%.0f Mbps) adalah kecepatan radio ke router, bukan kecepatan internet kamu."},

	// Tips (actions)
	"tip.sqm":           {"Turn on SQM / Smart Queue / QoS in your router settings, or avoid big downloads while gaming or on calls.", "Aktifkan SQM / Smart Queue / QoS di pengaturan router, atau hindari download besar saat main game atau video call."},
	"tip.weakSignal":    {"Move closer to the router or reduce obstacles (walls, metal) between them.", "Dekatkan perangkat ke router atau kurangi penghalang (tembok, logam) di antaranya."},
	"tip.use5GHz":       {"You are on 2.4 GHz. If your router also has a 5 GHz network, use it for more speed and less interference.", "Kamu memakai 2,4 GHz. Kalau router punya jaringan 5 GHz, pakai itu agar lebih cepat dan minim gangguan."},
	"tip.tryCable":      {"Test once with a LAN cable. If the problem disappears, Wi-Fi is the cause.", "Coba tes sekali dengan kabel LAN. Kalau masalahnya hilang, berarti penyebabnya Wi-Fi."},
	"tip.restartRouter": {"Restart the router, and try another Wi-Fi channel if neighbouring networks are crowded.", "Restart router, dan coba ganti channel Wi-Fi kalau jaringan tetangga ramai."},
	"tip.vpn":           {"A VPN is active. VPNs usually lower speed and add latency; disconnect it to compare.", "VPN sedang aktif. VPN biasanya menurunkan kecepatan dan menambah ping; matikan dulu untuk membandingkan."},
	"tip.isp":           {"If this keeps happening, contact your ISP and send them an exported report.", "Kalau ini terus terjadi, hubungi ISP kamu dan kirimkan laporan hasil export."},
	"tip.nearerServer":  {"Latency to this test server is high. Try a nearer server (Server → Find nearby servers).", "Ping ke server tes ini tinggi. Coba server yang lebih dekat (Server → Cari server terdekat)."},
	"tip.otherTraffic":  {"Pause other downloads, streaming or cloud sync on this and other devices, then test again.", "Hentikan dulu download, streaming atau sinkronisasi cloud di perangkat ini dan perangkat lain, lalu tes ulang."},

	// Ping target labels
	"target.gateway": {"Local Gateway", "Gateway Lokal (router)"},
	"target.tcp":     {"Test server (TCP connect)", "Server tes (koneksi TCP)"},

	// ISP plan
	"plan.down":      {"Download is %.0f%% of your %s plan.", "Download %.0f%% dari paket %s kamu."},
	"plan.up":        {"Upload is %.0f%% of your %s plan.", "Upload %.0f%% dari paket %s kamu."},
	"plan.wifiLimit": {"Your Wi-Fi link speed (%.0f Mbps) is below your plan, so Wi-Fi itself limits the speed here.", "Wi-Fi link speed kamu (%.0f Mbps) di bawah paket, jadi Wi-Fi-nya sendiri yang membatasi kecepatan di sini."},
	"tip.belowPlan":  {"Speed is well below your plan. Test once with a LAN cable next to the router; if it stays low at different times of day, contact your ISP with the ISP report (History → ISP report).", "Kecepatan jauh di bawah paket. Coba tes sekali dengan kabel LAN di dekat router; kalau tetap rendah di jam yang berbeda, hubungi ISP dengan laporan ISP (Riwayat → Laporan ISP)."},

	// DNS test
	"dns.system":      {"Your DNS (%s)", "DNS kamu (%s)"},
	"dns.switch":      {"%s answered fastest (%s, your current DNS %s). Switching DNS can make websites start loading faster.", "%s paling cepat (%s, DNS kamu sekarang %s). Mengganti DNS bisa membuat website mulai terbuka lebih cepat."},
	"dns.howto":       {"Windows: Settings → Network & internet → Wi-Fi or Ethernet → Hardware properties → DNS server assignment → Edit. Or change it once in your router for every device.", "Windows: Settings → Network & internet → Wi-Fi atau Ethernet → Hardware properties → DNS server assignment → Edit. Atau ganti sekali di router supaya berlaku untuk semua perangkat."},
	"dns.keep":        {"Your current DNS is already among the fastest; no change needed.", "DNS kamu sekarang sudah termasuk yang tercepat; tidak perlu diganti."},
	"dns.systemFails": {"Your current DNS failed %d of %d lookups. Switching to %s may make browsing more reliable.", "DNS kamu sekarang gagal %d dari %d pencarian. Pindah ke %s bisa membuat browsing lebih andal."},
	"dns.allFailed":   {"No DNS server answered. Check your connection, or a firewall may block DNS.", "Tidak ada server DNS yang menjawab. Cek koneksi, atau firewall mungkin memblokir DNS."},

	// Wi-Fi scan
	"scan.recommend": {"Your network uses channel %d (%s), shared with about %d nearby networks. Channel %d looks less crowded; you can change it in your router's Wi-Fi settings.", "Jaringan kamu memakai channel %d (%s), berbagi dengan sekitar %d jaringan terdekat. Channel %d terlihat lebih sepi; kamu bisa menggantinya di pengaturan Wi-Fi router."},
	"scan.fine":      {"Your channel %d (%s) is not crowded compared with the alternatives. No change needed.", "Channel %d (%s) kamu tidak ramai dibanding pilihan lain. Tidak perlu diganti."},
	"scan.best24":    {"On 2.4 GHz, channel %d is the least crowded (use only 1, 6 or 11 so networks do not overlap).", "Di 2,4 GHz, channel %d paling sepi (pakai hanya 1, 6 atau 11 supaya tidak tumpang-tindih)."},
	"scan.best5":     {"On 5 GHz, channel %d is the least crowded.", "Di 5 GHz, channel %d paling sepi."},
	"scan.try5":      {"2.4 GHz is crowded here (%d networks). If your router has 5 GHz, use it.", "2,4 GHz di sini ramai (%d jaringan). Kalau router punya 5 GHz, pakai itu."},
	"scan.none":      {"No Wi-Fi networks found. Is Wi-Fi turned on?", "Tidak ada jaringan Wi-Fi yang terdeteksi. Wi-Fi sudah dinyalakan?"},

	"dur.min": {"%d min", "%d menit"},
	"dur.sec": {"%d s", "%d detik"},

	// Stability monitor
	"mon.stable":     {"Stable: no drops and no big spikes during %s.", "Stabil: tidak ada putus dan lonjakan besar selama %s."},
	"mon.spiky":      {"Mostly connected, but latency jumped %d times. Games and calls may stutter at those moments.", "Sebagian besar tersambung, tapi ping melonjak %d kali. Game dan video call bisa tersendat di momen itu."},
	"mon.drops":      {"The connection dropped %d times (longest %s). This is what causes sudden lag or disconnects.", "Koneksi putus %d kali (paling lama %s). Inilah yang bikin tiba-tiba lag atau terputus."},
	"mon.lossy":      {"%s of probes were lost, but without long drops.", "%s probe hilang, tapi tanpa putus yang lama."},
	"mon.local":      {"The router also missed replies (%s), so the problem is likely Wi-Fi or the local network.", "Router juga tidak membalas sebagian (%s), jadi masalahnya kemungkinan di Wi-Fi atau jaringan lokal."},
	"mon.isp":        {"The router answered every time, so the drops happen beyond it (ISP or route).", "Router selalu membalas, jadi putusnya terjadi setelah router (ISP atau rute)."},
	"mon.tcp":        {"ICMP ping is blocked on this network; the monitor used TCP connections to the test server.", "Ping ICMP diblokir di jaringan ini; monitor memakai koneksi TCP ke server tes."},
	"mon.cancelled":  {"Stopped early after %s.", "Dihentikan lebih awal setelah %s."},
	"rep.monTitle":   {"STABILITY MONITOR", "MONITOR STABILITAS"},
	"rep.duration":   {"Duration", "Durasi"},
	"rep.target":     {"Target", "Target"},
	"rep.probes":     {"Probes", "Probe"},
	"rep.probesVal":  {"%d sent, %d received, loss %s", "%d terkirim, %d diterima, loss %s"},
	"rep.latency":    {"Latency", "Latensi"},
	"rep.latencyVal": {"avg %s · min %s · max %s · 95%% under %s · jitter %s", "rata-rata %s · min %s · maks %s · 95%% di bawah %s · jitter %s"},
	"rep.spikes":     {"Spikes", "Lonjakan"},
	"rep.outages":    {"Drops", "Putus"},
	"rep.outagesVal": {"%d (longest %s)", "%d (paling lama %s)"},
	"rep.router":     {"Router lost", "Router hilang"},
	"rep.verdict":    {"Conclusion", "Kesimpulan"},

	// ISP report
	"isp.title":      {"INTERNET QUALITY REPORT", "LAPORAN KUALITAS INTERNET"},
	"isp.intro":      {"Summary of speed tests measured with WiFi Speed + Ping Tester.", "Ringkasan speed test yang diukur dengan WiFi Speed + Ping Tester."},
	"isp.period":     {"Period", "Periode"},
	"isp.tests":      {"Tests", "Jumlah tes"},
	"isp.connection": {"Connection", "Koneksi"},
	"isp.plan":       {"Plan", "Paket"},
	"isp.planVal":    {"%s down / %s up", "%s download / %s upload"},
	"isp.planNone":   {"not set", "belum diisi"},
	"isp.summary":    {"SUMMARY", "RINGKASAN"},
	"isp.avgMinMax":  {"avg %s · lowest %s · highest %s", "rata-rata %s · terendah %s · tertinggi %s"},
	"isp.ofPlan":     {"Average of plan", "Rata-rata dari paket"},
	"isp.ofPlanVal":  {"download %s · upload %s", "download %s · upload %s"},
	"isp.belowPlan":  {"Below %.0f%% of plan", "Di bawah %.0f%% paket"},
	"isp.belowVal":   {"%d of %d tests", "%d dari %d tes"},
	"isp.byTime":     {"BY TIME OF DAY", "BERDASARKAN JAM"},
	"isp.timeRow":    {"%s  tests: %d · download avg %s · ping avg %s", "%s  jumlah tes: %d · rata-rata download %s · rata-rata ping %s"},
	"isp.slot0":      {"00:00–06:00", "00.00–06.00"},
	"isp.slot1":      {"06:00–12:00", "06.00–12.00"},
	"isp.slot2":      {"12:00–18:00", "12.00–18.00"},
	"isp.slot3":      {"18:00–24:00", "18.00–24.00"},
	"isp.peak":       {"Average download in the slowest period (%s) is %.0f%% lower than in the fastest (%s), which points to congestion at busy hours.", "Rata-rata download di periode paling lambat (%s) %.0f%% lebih rendah dari periode tercepat (%s), tanda jaringan padat di jam sibuk."},
	"isp.worst":      {"SLOWEST TESTS", "TES PALING LAMBAT"},
	"isp.all":        {"ALL TESTS", "SEMUA TES"},
	"isp.row":        {"%s  %-12s  ↓ %-10s ↑ %-10s ping %-7s loss %s", "%s  %-12s  ↓ %-10s ↑ %-10s ping %-7s loss %s"},
	"isp.wifiNote":   {"Note: %d of these tests were over Wi-Fi. Wi-Fi can lower results; tests over a LAN cable are the strongest evidence.", "Catatan: %d dari tes ini memakai Wi-Fi. Wi-Fi bisa menurunkan hasil; tes dengan kabel LAN adalah bukti paling kuat."},
	"isp.none":       {"No speed tests in this period.", "Tidak ada speed test di periode ini."},

	// Checks
	"check.adapter":        {"Network adapter", "Adapter jaringan"},
	"check.gateway":        {"Gateway", "Gateway"},
	"check.dns":            {"DNS", "DNS"},
	"check.internet":       {"Internet", "Internet"},
	"check.noGateway":      {"No default gateway reported", "Tidak ada default gateway"},
	"check.noPing":         {"Ping is not available on this system", "Ping tidak tersedia di sistem ini"},
	"check.gatewayOK":      {"%s reachable", "%s terjangkau"},
	"check.gatewayNoReply": {"%s did not answer ping (many routers block it)", "%s tidak membalas ping (banyak router memblokirnya)"},
	"check.unavailable":    {"Unavailable", "Tidak tersedia"},
	"check.dnsOK":          {"Resolved in %s", "Berhasil dalam %s"},
	"check.internetOK":     {"Reached test server %s", "Terhubung ke server tes %s"},

	// Failures
	"fail.noNetwork.title":      {"No Network Connection", "Tidak Ada Koneksi Jaringan"},
	"fail.noNetwork.reason":     {"No active network adapter was found.", "Tidak ditemukan adapter jaringan yang aktif."},
	"fail.noNetwork.suggestion": {"Connect to Wi-Fi or plug in a network cable, then try again.", "Sambungkan ke Wi-Fi atau colok kabel LAN, lalu coba lagi."},
	"fail.offline.title":        {"Internet Unavailable", "Internet Tidak Tersedia"},
	"fail.dns.reason":           {"Names on the internet could not be looked up (DNS failed).", "Alamat di internet tidak bisa dicari (DNS gagal)."},
	"fail.dns.suggestion":       {"Speed test cannot continue. Check your internet connection or DNS settings, then try again.", "Speed test tidak bisa dilanjutkan. Periksa koneksi internet atau pengaturan DNS, lalu coba lagi."},
	"fail.offline.reasonPrefix": {"The internet could not be reached. ", "Internet tidak bisa dijangkau. "},
	"fail.offline.suggPrefix":   {"Speed test cannot continue. ", "Speed test tidak bisa dilanjutkan. "},
	"fail.speed.title":          {"Speed Test Failed", "Speed Test Gagal"},

	// Error explanations: reason / suggestion
	"err.cancelled.r":   {"The test was cancelled.", "Tes dibatalkan."},
	"err.cancelled.s":   {"Press Start Test to run it again.", "Tekan Mulai Tes untuk menjalankannya lagi."},
	"err.ratelimit.r":   {"The test server is limiting requests because too many tests were run recently.", "Server tes membatasi permintaan karena terlalu banyak tes dalam waktu singkat."},
	"err.ratelimit.s":   {"Wait a minute and try again, or choose another test server.", "Tunggu sebentar lalu coba lagi, atau pilih server tes lain."},
	"err.http.r":        {"The test server returned an error (%s).", "Server tes mengembalikan error (%s)."},
	"err.http.s":        {"Try again later or select another test server.", "Coba lagi nanti atau pilih server tes lain."},
	"err.proxy.r":       {"A proxy or firewall blocked the connection to the test server.", "Proxy atau firewall memblokir koneksi ke server tes."},
	"err.proxy.s":       {"Try another network, or ask your network administrator to allow the test server.", "Coba jaringan lain, atau minta admin jaringan mengizinkan server tes."},
	"err.dns.r":         {"The test server's name could not be looked up (DNS).", "Nama server tes tidak bisa dicari (DNS)."},
	"err.dns.s":         {"Check that you are connected to the internet and that your DNS server works. Try again in a moment.", "Pastikan kamu terhubung ke internet dan DNS berfungsi. Coba lagi sebentar lagi."},
	"err.tls.r":         {"A secure connection to the test server could not be established.", "Koneksi aman ke server tes tidak bisa dibuat."},
	"err.tls.s":         {"Check that the system date and time are correct. A proxy, captive portal or antivirus that inspects HTTPS can also cause this.", "Pastikan tanggal dan jam di komputer benar. Proxy, halaman login Wi-Fi publik, atau antivirus yang memeriksa HTTPS juga bisa jadi penyebabnya."},
	"err.reset.r":       {"The connection was reset during the test.", "Koneksi terputus di tengah tes."},
	"err.reset.s":       {"Your connection may be unstable, or a firewall/VPN interrupted it. Try again.", "Koneksi mungkin tidak stabil, atau firewall/VPN memutusnya. Coba lagi."},
	"err.refused.r":     {"The test server refused the connection.", "Server tes menolak koneksi."},
	"err.refused.s":     {"Try again or select another test server.", "Coba lagi atau pilih server tes lain."},
	"err.timeout.r":     {"The test server did not respond in time.", "Server tes tidak merespons tepat waktu."},
	"err.timeout.s":     {"Try again or select another test server. A firewall, VPN or a very slow connection can also cause this.", "Coba lagi atau pilih server tes lain. Firewall, VPN atau koneksi yang sangat lambat juga bisa jadi penyebabnya."},
	"err.estimated.r":   {"Upload is very slow: no data finished uploading in time, so this value is only an estimate.", "Upload sangat lambat: tidak ada data yang selesai terkirim tepat waktu, jadi angka ini hanya perkiraan."},
	"err.estimated.s":   {"Move closer to the router or try another network. A weak mobile signal often causes this.", "Dekatkan perangkat ke router atau coba jaringan lain. Sinyal seluler yang lemah sering jadi penyebabnya."},
	"err.nodata.r":      {"No data could be transferred to or from the test server.", "Tidak ada data yang bisa dikirim ke atau diterima dari server tes."},
	"err.nodata.s":      {"Try again. If it keeps failing, a firewall or proxy may be blocking speed tests.", "Coba lagi. Kalau terus gagal, firewall atau proxy mungkin memblokir speed test."},
	"err.unreachable.r": {"The network is unreachable.", "Jaringan tidak bisa dijangkau."},
	"err.unreachable.s": {"Check that Wi-Fi or the cable is connected, then try again.", "Pastikan Wi-Fi atau kabel tersambung, lalu coba lagi."},
	"err.other.r":       {"The test could not be completed.", "Tes tidak bisa diselesaikan."},
	"err.other.s":       {"Try again. If the problem continues, try another network or test server.", "Coba lagi. Kalau masalahnya berlanjut, coba jaringan atau server tes lain."},

	// Wi-Fi notes (model.WiFiInfo.NoteKey)
	"wifi.hotspot":     {"Gateway address is typical of a phone hotspot.", "Alamat gateway ini biasanya milik hotspot HP."},
	"wifi.ssidHidden":  {"Windows hides the Wi-Fi name. On Windows 11, turn on Settings → Privacy & security → Location → Let desktop apps access your location.", "Windows menyembunyikan nama Wi-Fi. Di Windows 11, nyalakan Settings → Privacy & security → Location → Let desktop apps access your location."},
	"wifi.unavailable": {"Wi-Fi details unavailable: %s", "Detail Wi-Fi tidak tersedia: %s"},

	// Report (TXT export)
	"rep.title":        {"WIFI SPEED + PING TEST", "WIFI SPEED + PING TEST"},
	"rep.date":         {"Date", "Tanggal"},
	"rep.mode":         {"Mode", "Mode"},
	"rep.connection":   {"Connection", "Koneksi"},
	"rep.adapter":      {"Adapter", "Adapter"},
	"rep.network":      {"Network", "Jaringan"},
	"rep.signal":       {"Signal", "Sinyal"},
	"rep.band":         {"Band", "Band"},
	"rep.channel":      {"Channel", "Channel"},
	"rep.linkSpeed":    {"Wi-Fi Link Speed", "Wi-Fi Link Speed"},
	"rep.linkSpeedVal": {"%.0f Mbps (radio rate, not internet speed)", "%.0f Mbps (kecepatan radio, bukan kecepatan internet)"},
	"rep.cancelled":    {"Test was cancelled; results are incomplete.", "Tes dibatalkan; hasil tidak lengkap."},
	"rep.reason":       {"Reason", "Penyebab"},
	"rep.suggestion":   {"Suggestion", "Saran"},
	"rep.notMeasured":  {"not measured", "tidak terukur"},
	"rep.failed":       {"failed (%s)", "gagal (%s)"},
	"rep.incomplete":   {"%s (incomplete: %s)", "%s (tidak lengkap: %s)"},
	"rep.pingDownload": {"Ping During Download", "Ping Saat Download"},
	"rep.pingUpload":   {"Ping During Upload", "Ping Saat Upload"},
	"rep.quality":      {"Quality", "Kualitas"},
	"rep.pingDetails":  {"PING DETAILS", "DETAIL PING"},
	"rep.pingLine":     {"%d sent, %d received, loss %s", "%d terkirim, %d diterima, loss %s"},
	"rep.pingStats":    {"min %s / avg %s / max %s, jitter %s", "min %s / rata-rata %s / maks %s, jitter %s"},
	"rep.noReplies":    {" (no replies; the target may block ping)", " (tidak ada balasan; target mungkin memblokir ping)"},
	"rep.notes":        {"NOTES", "CATATAN"},
	"rep.tips":         {"TIPS", "SARAN"},
	"rep.server":       {"Test server: %s", "Server tes: %s"},
	"rep.footer":       {"Measured values reflect this moment, this device and this test server.\nThey are not a guarantee of your ISP plan speed.", "Nilai yang terukur mencerminkan kondisi saat ini, perangkat ini dan server tes ini.\nBukan jaminan kecepatan paket ISP kamu."},

	// Short shareable text (Copy result)
	"share.head":    {"WiFi Speed + Ping Test", "Tes WiFi Speed + Ping"},
	"share.speeds":  {"↓ %s   ↑ %s", "↓ %s   ↑ %s"},
	"share.latency": {"Ping %s · Jitter %s · Loss %s", "Ping %s · Jitter %s · Loss %s"},
	"share.quality": {"Quality: %s", "Kualitas: %s"},

	// App errors
	"app.runFirst":     {"run a test first", "jalankan tes dulu"},
	"app.historyEmpty": {"history is empty", "riwayat masih kosong"},
	"app.busy":         {"a test is already running", "tes sedang berjalan"},
	"app.monBusy":      {"the stability monitor is running", "monitor stabilitas sedang berjalan"},
	"app.badTarget":    {"invalid host: use an IP address or a name like game.example.com", "host tidak valid: pakai alamat IP atau nama seperti game.example.com"},
	"app.tooMany":      {"at most %d custom targets", "maksimal %d target tambahan"},
	"app.noWifiScan":   {"Wi-Fi scanning is not available on this system", "Pemindaian Wi-Fi tidak tersedia di sistem ini"},
}
