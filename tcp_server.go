package main

import (
	"errors"
	"log"
	"net"
	"net/http"
	"net/url"
	"sync"
)

type tcpHandler func(*net.TCPAddr) error

type tcpServer struct {
	addr   string
	remote string
}

func (srv *tcpServer) run() {
	l, err := net.Listen("tcp", srv.addr)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("listening %s\n", l.Addr().String())
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println(err)
			return
		}
		go srv.serve(conn)
	}
}

// connect creates a client to handle an incoming connection.
func (srv *tcpServer) connect(handle tcpHandler) error {
	l, err := net.Listen("tcp", srv.addr)
	if err != nil {
		return err
	}
	if handle == nil {
		return errors.New("handle function should not be nil")
	}
	defer l.Close()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		wg.Done()
		conn, err := l.Accept()
		if err != nil {
			log.Fatalf("accept error: %v", err)
		}
		srv.serve(conn)
	}()
	wg.Wait()

	tcpAddr, _ := l.Addr().(*net.TCPAddr)
	return handle(tcpAddr)
}

func (srv *tcpServer) serve(c net.Conn) {
	defer c.Close()

	u, _ := url.Parse(srv.remote)

	log.Printf("connected %s => %s\n", c.RemoteAddr(), srv.remote)

	if u.Scheme == "ws" || u.Scheme == "wss" {
		conn1, resp, err := dialer.Dial(srv.remote, nil)
		if err != nil {
			log.Printf("websocket dial error: %v %s", err, resp.Status)
			return
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusSwitchingProtocols {
			log.Printf("unexpected HTTP status code: %d\n", resp.StatusCode)
			return
		}
		defer conn1.Close()

		forwardWS2TCP(conn1, c)
		return
	}

	if u.Scheme == "tcp" {
		conn1, err := net.Dial("tcp", u.Host)
		if err != nil {
			log.Println(err)
			return
		}
		defer conn1.Close()

		forwardTCP2TCP(c, conn1)
		return
	}

	log.Printf("unsupported scheme %s\n", u.Scheme)
}
