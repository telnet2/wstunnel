package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	log "github.com/fangdingjun/go-log/v5"
)

func main() {
	var loglevel string
	var remoteAddr, localAddr string

	flag.StringVar(&remoteAddr, "r", "", "remote WS url")
	flag.StringVar(&localAddr, "l", "tcp://127.0.0.1:60060", "listening address (e.g., tcp://127.0.0.1:60060)")
	flag.StringVar(&loglevel, "log_level", "INFO", "log level")
	flag.Parse()

	if remoteAddr == "" {
		log.Fatalf("-r remoteAddr is missing")
	}

	cfg := conf{
		ProxyConfig: []proxyItem{{Listen: localAddr, Remote: remoteAddr}},
	}

	if lv, err := log.ParseLevel(loglevel); err == nil {
		log.Default.Level = lv
	}

	makeServers(cfg)

	ch := make(chan os.Signal, 2)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	select {
	case s := <-ch:
		log.Printf("received signal %s, exit.", s)
	}
}
