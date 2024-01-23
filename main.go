package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"

	"darvaza.org/darvaza/shared/tls/sni"
)

func main() {
	lis, err := net.Listen("tcp", ":443")
	if err != nil {
		panic(err)
	}
	slog.Info("run", "listener", ":443")
	for {
		conn, _ := lis.Accept()
		go handle(conn)
	}
}

func handle(conn net.Conn) {
	defer conn.Close()
	tmp := make([]byte, 1024)
	n, _ := conn.Read(tmp)

	s, err := sni.ReadClientHelloInfo(context.Background(), bytes.NewReader(tmp[:n]))
	if err != nil {
		// 不是https
		return
	}

	ips, err := QueryDns(s.ServerName)
	if err != nil {
		slog.Error("QueryDns", "err", err)
	}
	// 转发
	ip := ips[0].String()
	newConn, err := net.Dial("tcp", ip+":443")
	if err != nil {
		slog.Error("new conn", "err", err, "remote", conn.RemoteAddr(), "dst", ip, "domain", s.ServerName)
		return
	}
	defer newConn.Close()
	signal := make(chan struct{}, 1)

	newConn.Write(tmp[:n])
	slog.Info(fmt.Sprintf("%s <-> %s:443", conn.RemoteAddr(), ip), "remote", conn.RemoteAddr(), "dst", ip, "domain", s.ServerName)
	go f(newConn, conn, signal)
	go f(conn, newConn, signal)
	<-signal
}

func f(dst io.WriteCloser, src io.ReadCloser, signal chan struct{}) {
	io.Copy(dst, src)
	signal <- struct{}{}
}
