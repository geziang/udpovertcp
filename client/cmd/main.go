package main

import (
	"flag"
	"github.com/geziang/udpovertcp"
	"github.com/geziang/udpovertcp/client"
	"log"
	"net"
	"sync"
	"time"
)

var (
	mConn sync.Map
)

func main() {
	lisAddr := flag.String("listen", "127.0.0.1:1194", "udp listen addr")
	servAddr := flag.String("server", "", "server addr")
	sec := flag.String("secret", "", "secret")
	flag.Parse()

	addr, err := net.ResolveUDPAddr("udp", *lisAddr)
	if err != nil {
		log.Fatalln(err)
		return
	}

	listener, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalln(err)
		return
	}
	defer listener.Close()

	cl, err := client.NewUDPOverTCPClient(*servAddr, *sec, listener, 2)
	if err != nil {
		log.Fatalln(err)
		return
	}
	err = cl.FetchClientId()
	if err != nil {
		log.Fatalln(err)
		return
	}

	go cleanLoop()

	buf := make([]byte, 2048)
	for {
		n, rAddr, err := listener.ReadFromUDP(buf)
		if err != nil {
			log.Println(err)
			continue
		}

		var conn *client.ProxyTCPConn
		key := rAddr.String()
		v, ok := mConn.Load(key)
		if ok {
			conn = v.(*client.ProxyTCPConn)
		} else {
			conn = client.NewProxyTCPConn(cl, rAddr)
			mConn.Store(key, conn)
		}

		nbuf := make([]byte, n)
		copy(nbuf, buf[:n])

		conn.Send(&udpovertcp.UDPPacket{
			ClientId: cl.ClientId,
			SrcAddr:  rAddr,
			DestAddr: nil,
			Data:     nbuf,
		})
	}
}

func cleanLoop() {
	for {
		time.Sleep(time.Second)
		mConn.Range(func(key, value interface{}) bool {
			c := value.(*client.ProxyTCPConn)
			if c.Closed {
				mConn.Delete(key)
			}
			return true
		})
	}
}