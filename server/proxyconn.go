package server

import (
	"github.com/geziang/udpovertcp"
	"io"
	"log"
	"net"
	"sync/atomic"
)

type UDPProxyConn struct {
	srcAddr       *net.UDPAddr
	clientId      string
	conn          *net.UDPConn
	chRecv        chan *udpovertcp.UDPPacket
	accessCounter int32
}

func NewUDPProxyConn(srcAddr *net.UDPAddr, clientId string) (*UDPProxyConn, error) {
	ch := make(chan *udpovertcp.UDPPacket)
	addr, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}
	proxyConn := &UDPProxyConn{
		srcAddr:       srcAddr,
		clientId:      clientId,
		conn:          conn,
		chRecv:        ch,
		accessCounter: 0,
	}

	go proxyConn.recvLoop()

	return proxyConn, nil
}

func (c *UDPProxyConn) recvLoop() {
	defer close(c.chRecv)
	defer c.Close()

	buf := make([]byte, 2048)
	for {
		n, addr, err := c.conn.ReadFromUDP(buf)
		if err != nil {
			if err != io.EOF {
				log.Println(err)
			}
			break
		}

		nbuf := make([]byte, n)
		copy(nbuf, buf[:n])

		c.chRecv <- &udpovertcp.UDPPacket{
			ClientId: c.clientId,
			SrcAddr:  addr,
			DestAddr: c.srcAddr,
			Data:     nbuf,
		}
	}
}

func (c *UDPProxyConn) Close() error {
	return c.conn.Close()
}

func (c *UDPProxyConn) Recv() (*udpovertcp.UDPPacket, bool) {
	packet, ok := <-c.chRecv
	return packet, ok
}

func (c *UDPProxyConn) Send(b []byte, addr *net.UDPAddr) (int, error) {
	return c.conn.WriteToUDP(b, addr)
}

func (c *UDPProxyConn) IncAccessCounter() {
	atomic.AddInt32(&c.accessCounter, 1)
}

func (c *UDPProxyConn) DecAccessCounter() {
	atomic.AddInt32(&c.accessCounter, -1)
}

func (c *UDPProxyConn) GetAccessCounter() int32 {
	return atomic.LoadInt32(&c.accessCounter)
}
