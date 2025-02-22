package main

import (
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"

	log "github.com/fangdingjun/go-log/v5"
	"github.com/gorilla/websocket"
	"gopkg.in/yaml.v2"
)

type testWSHandler struct{}

func (ts *testWSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	defer conn.Close()

	for {
		typ, data, err := conn.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, 1006) {
				log.Println(err)
			}
			return
		}
		conn.WriteMessage(typ, data)
	}

}

func echoServerWS(addr string) {
	srv := &http.Server{Addr: addr, Handler: &testWSHandler{}}
	if err := srv.ListenAndServe(); err != nil {
		log.Println(err)
	}
}

func echoServer(addr string) {
	l1, err := net.Listen("tcp", addr)
	if err != nil {
		log.Println(err)
		return
	}
	defer l1.Close()

	for {
		c1, err := l1.Accept()
		if err != nil {
			log.Println(err)
			return
		}
		go func(c net.Conn) {
			defer c.Close()
			data := make([]byte, 1024)
			for {
				n, err := c.Read(data)
				if err != nil {
					if err != io.EOF {
						log.Println(err)
					}
					break
				}
				c.Write(data[:n])
				log.Infof("%s receive: %s", addr, data[:n])
			}
		}(c1)
	}
}

func sendAndRecv(addr string, msg string) string {
	u, _ := url.Parse(addr)
	if u.Scheme == "ws" || u.Scheme == "wss" {
		return _sendAndRecvWS(addr, msg)
	}
	if u.Scheme == "tcp" {
		return _sendAndRecvTCP(addr, msg)
	}
	return ""
}

func _sendAndRecvTCP(addr string, msg string) string {
	u, _ := url.Parse(addr)
	c, err := net.Dial("tcp", u.Host)
	if err != nil {
		log.Println(err)
		return ""
	}

	defer c.Close()

	_, err = c.Write([]byte(msg))
	if err != nil {
		log.Println(err)
		return ""
	}

	data := make([]byte, 100)
	n, err := c.Read(data)
	if err != nil {
		log.Println(err)
		return ""
	}
	return string(data[:n])
}

func _sendAndRecvWS(addr string, msg string) string {
	c1, resp, err := dialer.Dial(addr, nil)
	if err != nil {
		log.Println(err)
		return ""
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusSwitchingProtocols {
		log.Errorf("dial ws code %d", resp.StatusCode)
		return ""
	}

	defer c1.Close()

	err = c1.WriteMessage(websocket.BinaryMessage, []byte(msg))
	if err != nil {
		log.Println(err)
		return ""
	}

	_, d, err := c1.ReadMessage()
	if err != nil {
		log.Println(err)
		return ""
	}
	return string(d)
}

func TestServer(t *testing.T) {
	cfgfile := "config.example.yaml"

	log.Default.Level = log.DEBUG

	data, err := ioutil.ReadFile(cfgfile)
	if err != nil {
		log.Fatal(err)
	}
	var cfg conf
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatal(err)
	}

	makeServers(cfg)

	go echoServer("127.0.0.1:2903")
	go echoServer("127.0.0.1:2904")
	go echoServerWS("127.0.0.1:2907")
	go echoServerWS("127.0.0.1:2908")

	time.Sleep(time.Second)

	testdata := []struct {
		addr string
		msg  string
	}{
		{"ws://127.0.0.1:2901/p1", "p1"},
		{"ws://127.0.0.1:2902/p2", "p2"},
		{"tcp://127.0.0.1:2905", "c3"},
		{"tcp://127.0.0.1:2906", "c4"},
	}
	for _, tt := range testdata {
		_m := sendAndRecv(tt.addr, tt.msg)
		if _m != tt.msg {
			t.Errorf("expected %s, got %s", tt.msg, _m)
		}
	}
}
