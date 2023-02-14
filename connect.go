package main

import (
	"net"
	"strconv"

	"github.com/telnet2/wstunnel/term"
)

func connect(remoteAddr string, args []string) error {
	ts := tcpServer{addr: "localhost:", remote: remoteAddr}
	return ts.connect(func(a *net.TCPAddr) error {
		args := append([]string{
			"-o", "UserKnownHostsFile=/dev/null",
			"-o", "StrictHostKeyChecking=no",
			"-p", strconv.Itoa(a.Port),
			"localhost",
		}, args...)
		return term.ExecSSH(args)
	})
}
