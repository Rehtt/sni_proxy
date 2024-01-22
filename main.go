package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net"

	"darvaza.org/darvaza/shared/tls/sni"
	"github.com/miekg/dns"
)

func main() {
	lis, err := net.Listen("tcp", ":443")
	if err != nil {
		panic(err)
	}
	fmt.Println("run")
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
		log.Println("not https")
		return
	}
	c := new(dns.Client)
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(s.ServerName), dns.TypeA)
	r, _, err := c.Exchange(m, "8.8.8.8:53")
	if err != nil {
		panic(err)
	}
	ip := r.Answer
	if len(ip) == 0 {
		log.Println("not find ip")
		return
	}

	// 转发
	newConn, err := net.Dial("tcp", ip[0].(*dns.A).A.String()+":443")
	if err != nil {
		fmt.Println("newConn", err)
		return
	}
	defer newConn.Close()
	signal := make(chan struct{}, 1)

	newConn.Write(tmp[:n])
	go f(newConn, conn, signal)
	go f(conn, newConn, signal)
	<-signal
}

func f(dst io.WriteCloser, src io.ReadCloser, signal chan struct{}) {
	io.Copy(dst, src)
	signal <- struct{}{}
}
