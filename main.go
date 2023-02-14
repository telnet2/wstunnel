package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var (
		remoteAddr, localAddr string
		daemon                bool
		logFile               string
	)

	flag.StringVar(&remoteAddr, "r", "", "remote WS url")
	flag.StringVar(&localAddr, "l", "tcp://127.0.0.1:60060", "listening address (e.g., tcp://127.0.0.1:60060)")
	flag.StringVar(&logFile, "log_file", "./wstunnel.log", "log file")
	flag.BoolVar(&daemon, "daemon", false, "run as a daemon mode")
	flag.Parse()

	if remoteAddr == "" {
		log.Fatalf("-r remoteAddr is missing")
	}

	lw, err := os.Create(logFile)
	if err != nil {
		log.Fatalln(err)
	}
	log.SetOutput(lw)

	if daemon {
		cfg := conf{
			ProxyConfig: []proxyItem{{Listen: localAddr, Remote: remoteAddr}},
		}

		makeServers(cfg)

		ch := make(chan os.Signal, 2)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		<-ch
	} else {
		err := connect(remoteAddr, flag.Args())
		if err != nil {
			fmt.Println("\n\n", logFile, ">>>>>>>>>>>>>>>")
			lr, _ := os.Open(logFile)
			_, _ = io.Copy(os.Stdout, lr)
		}
	}
}
