// Command wifispeedtester is the WiFi Speed + Ping Tester.
//
// With no arguments it opens the desktop window (Windows). Other modes:
//
//	--cli            run one test in the console and print a report
//	--gaming         use the gaming profile (with --cli)
//	--quick          ping only, no speed test (with --cli)
//	--lang id|en     report language (with --cli)
//	--list-servers   find public LibreSpeed servers and rank them by latency
//	--dns            compare DNS servers
//	--wifi-scan      list nearby Wi-Fi networks and recommend a channel
//	--json           print the result as JSON (with --cli)
//	--serve [port]   serve the UI on http://127.0.0.1:port for development
//	--print-config   print the default configuration (JSON)
//	--write-config   write the default config file next to the executable
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/tehgeii/wifi-speed-tester/internal/analysis"
	"github.com/tehgeii/wifi-speed-tester/internal/app"
	"github.com/tehgeii/wifi-speed-tester/internal/config"
	"github.com/tehgeii/wifi-speed-tester/internal/dnsbench"
	"github.com/tehgeii/wifi-speed-tester/internal/engine"
	"github.com/tehgeii/wifi-speed-tester/internal/export"
	"github.com/tehgeii/wifi-speed-tester/internal/i18n"
	"github.com/tehgeii/wifi-speed-tester/internal/network"
	"github.com/tehgeii/wifi-speed-tester/internal/servers"
	"github.com/tehgeii/wifi-speed-tester/internal/ui"
)

func main() {
	cli := flag.Bool("cli", false, "run a test in the console")
	gaming := flag.Bool("gaming", false, "use the gaming profile (with --cli)")
	asJSON := flag.Bool("json", false, "print JSON instead of text (with --cli)")
	quick := flag.Bool("quick", false, "ping only, no download/upload (with --cli)")
	lang := flag.String("lang", "en", "report language: en or id (with --cli)")
	listServers := flag.Bool("list-servers", false, "find public LibreSpeed servers and rank them by latency")
	dnsTest := flag.Bool("dns", false, "compare how fast DNS servers answer")
	wifiScan := flag.Bool("wifi-scan", false, "list nearby Wi-Fi networks and recommend a channel")
	serve := flag.Int("serve", 0, "serve the UI on 127.0.0.1:`port` (development)")
	writeCfg := flag.Bool("write-config", false, "write "+config.FileName+" with the defaults next to the executable")
	printCfg := flag.Bool("print-config", false, "print the default configuration as JSON")
	version := flag.Bool("version", false, "print the version")
	prepareConsole()
	flag.Parse()

	exeDir := executableDir()
	switch {
	case *version:
		fmt.Println("WiFi Speed + Ping Tester", app.Version)
	case *printCfg:
		os.Stdout.Write(config.ExampleJSON())
	case *writeCfg:
		p := filepath.Join(exeDir, config.FileName)
		if _, err := os.Stat(p); err == nil {
			fatalf("%s already exists", p)
		}
		if err := os.WriteFile(p, config.ExampleJSON(), 0o644); err != nil {
			fatalf("%v", err)
		}
		fmt.Println("Wrote", p)
	case *listServers:
		os.Exit(runListServers(exeDir))
	case *dnsTest:
		os.Exit(runDNS(exeDir, i18n.Parse(*lang)))
	case *wifiScan:
		os.Exit(runWiFiScan(i18n.Parse(*lang)))
	case *cli:
		mode := "general"
		if *gaming {
			mode = "gaming"
		}
		if *quick {
			mode = "quick"
		}
		os.Exit(runCLI(exeDir, mode, i18n.Parse(*lang), *asJSON))
	case *serve > 0:
		runServer(exeDir, *serve)
	default:
		runGUI(exeDir)
	}
}

func executableDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	if p, err := filepath.EvalSymlinks(exe); err == nil {
		exe = p
	}
	return filepath.Dir(exe)
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

func runCLI(exeDir, mode string, lang i18n.Lang, asJSON bool) int {
	cfg, path, err := config.Load(exeDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "config:", err, "(using defaults)")
	} else if path != "" {
		fmt.Fprintln(os.Stderr, "config:", path)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	last := ""
	res := engine.New(cfg).Run(ctx, engine.Options{Mode: mode, Lang: lang}, func(ev engine.Event) {
		switch ev.Type {
		case "check":
			if ev.Check.Status != "pending" {
				fmt.Fprintf(os.Stderr, "  [%s] %s  %s\n", ev.Check.Status, ev.Check.Name, ev.Check.Detail)
			}
		case "stage":
			if ev.Stage != last {
				last = ev.Stage
				fmt.Fprintf(os.Stderr, "%s...\n", map[string]string{
					engine.StageConnection: "Checking connection", engine.StagePing: "Ping test",
					engine.StageDownload: "Download test", engine.StageUpload: "Upload test",
				}[ev.Stage])
			}
		case "progress":
			if ev.Stage == engine.StageDownload || ev.Stage == engine.StageUpload {
				fmt.Fprintf(os.Stderr, "\r  %3.0f%%  %8.1f Mbps", ev.Progress*100, ev.Mbps)
			}
		case "speed":
			fmt.Fprintln(os.Stderr)
		}
	})
	opts := export.Options{Unit: "Mbps", IncludeSensitive: true, Lang: lang}
	if asJSON {
		b, _ := export.JSON(res, opts)
		fmt.Println(string(b))
	} else {
		fmt.Println()
		os.Stdout.Write(export.TXT(res, opts))
	}
	if res.Failure != nil || res.Cancelled {
		return 1
	}
	return 0
}

func runListServers(exeDir string) int {
	cfg, _, _ := config.Load(exeDir)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	fmt.Fprintln(os.Stderr, "Fetching", cfg.ServerListURL)
	list, err := servers.Fetch(ctx, &http.Client{Transport: &http.Transport{Proxy: http.ProxyFromEnvironment}}, cfg.ServerListURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	fmt.Fprintf(os.Stderr, "Measuring latency to %d servers...\n", len(list))
	ranked := servers.Probe(ctx, list, 8, nil)
	ok := 0
	for _, c := range ranked {
		if c.Error != "" {
			continue
		}
		ok++
		if ok <= 20 {
			fmt.Printf("%7.1f ms  %-45s %s\n", c.LatencyMs, c.Server.Name, c.Server.URL)
		}
	}
	fmt.Fprintf(os.Stderr, "%d of %d servers reachable\n", ok, len(ranked))
	if ok == 0 {
		return 1
	}
	return 0
}

func runDNS(exeDir string, lang i18n.Lang) int {
	cfg, _, _ := config.Load(exeDir)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	var system []string
	if info, err := network.Detect(ctx); err == nil {
		system = info.DNS
	}
	list := dnsbench.Servers(system, cfg.DNSServers, lang)
	res := dnsbench.Run(ctx, list, dnsbench.Domains, dnsbench.Options{})
	ok := 0
	for _, r := range res {
		if r.OK > 0 {
			ok++
			fmt.Printf("%7.1f ms  %-36s %d/%d\n", r.MedianMs, r.Label, r.OK, r.OK+r.Failed)
		} else {
			fmt.Printf("     fail  %-36s %s\n", r.Label, r.Error)
		}
	}
	for _, a := range analysis.DNSAdvice(res, lang) {
		fmt.Println("-", a)
	}
	if ok == 0 {
		return 1
	}
	return 0
}

func runWiFiScan(lang i18n.Lang) int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	nets, err := network.ScanWiFi(ctx, true)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	rep := analysis.AnalyzeScan(nets, lang)
	for _, n := range rep.Networks {
		mark := " "
		if n.Connected {
			mark = "*"
		}
		name := n.SSID
		if name == "" {
			name = "(hidden)"
		}
		fmt.Printf("%s %-32s ch %-3d %-7s %4d dBm\n", mark, name, n.Channel, n.Band, n.RSSI)
	}
	for _, a := range rep.Advice {
		fmt.Println("-", a)
	}
	return 0
}

func runServer(exeDir string, port int) {
	srv := ui.NewDevServer()
	srv.App = app.New(exeDir, app.Host{Emit: srv.Emit}, log.Printf)
	ln, err := ui.ListenLoopback(port)
	if err != nil {
		fatalf("%v", err)
	}
	log.Printf("UI available at http://%s (exports are written to %s)", ln.Addr(), srv.App.ExportsDir())
	log.Fatal(serveHTTP(ln, srv))
}
