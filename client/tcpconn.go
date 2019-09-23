package client

import (
	"encoding/gob"
	"github.com/geziang/udpovertcp"
	"log"
	"net"
	"time"
)

type ProxyTCPConn struct {
	Closed  bool
	chSend  chan *udpovertcp.UDPPacket
	chIdle  chan int
	chClose chan int
	client  *UDPOverTCPClient
	SrcAddr *net.UDPAddr
}

func NewProxyTCPConn(client *UDPOverTCPClient, srcAddr *net.UDPAddr) *ProxyTCPConn {
	ch := make(chan *udpovertcp.UDPPacket, 10)
	c := &ProxyTCPConn{
		Closed:  false,
		chSend:  ch,
		chIdle:  make(chan int),
		chClose: make(chan int),
		client:  client,
		SrcAddr: srcAddr,
	}

	go c.idleLoop()

	go c.sendLoop()
	//这一段用于多线程发送
	if c.client.ConcurrentCount > 1 {
		go func() {
			count := c.client.ConcurrentCount - 1
			time.Sleep(5 * time.Second)
			for i := 0; i < count; i++ {
				go c.sendLoop()
			}
		}()
	}

	return c
}

func (c *ProxyTCPConn) sendLoop() {
	conn, err := c.client.MyDial("tcp", c.client.ServerAddr)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()
	enc := gob.NewEncoder(conn)
	D := gob.NewDecoder(conn)

	hs := &udpovertcp.Handshake{
		Secret:   c.client.Secret,
		ClientId: c.client.ClientId,
	}
	if err := enc.Encode(hs); err != nil {
		log.Println(err)
		return
	}

	go func() {
		var packet udpovertcp.UDPPacket
		for !c.Closed {
			err := D.Decode(&packet)
			if err != nil {
				break
			}

			_, err = c.client.UdpConn.WriteToUDP(packet.Data, packet.DestAddr)
			if err != nil {
				log.Println(err)
			}
		}
	}()

	var packet *udpovertcp.UDPPacket
	var ok bool
	for !c.Closed {
		select {
		case _, ok = <-c.chClose:
		case packet, ok = <-c.chSend:
		}
		
		if !ok {
			break
		}

		if err := enc.Encode(packet); err != nil {
			log.Println(err)
			break
		}
	}
}

func (c *ProxyTCPConn) Close() {
	if !c.Closed {
		c.Closed = true
		close(c.chClose)
	}
}

func (c *ProxyTCPConn) Send(packet *udpovertcp.UDPPacket) {
	select {
	case <-c.chClose:
	case c.chIdle <- 0:
	}
	
	select {
	case <-c.chClose:
	case c.chSend <- packet:
	}
}

func (c *ProxyTCPConn) idleLoop() {
	running := true
	for running {
		select {
		case _, ok := <-c.chIdle:
			if !ok {
				running = false
			}
		case <-time.After(1 * time.Minute):
			c.Close()
			running = false
		}
	}
}
