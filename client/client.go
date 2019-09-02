package client

import (
	"encoding/gob"
	"github.com/geziang/udpovertcp"
	"net"
	"sync"
)

type UDPOverTCPClient struct {
	ServerAddr      string
	ClientId        string
	Secret          string
	UdpConn         *net.UDPConn
	ConcurrentCount int
	mConn           sync.Map
	MyDial          func(network, address string) (net.Conn, error)
}

/*
 serverAddr 服务器tcp地址
 secret 简易鉴权密码
 listenAddr 监听udp地址，用于发送返回数据的socket
*/
func NewUDPOverTCPClient(serverAddr, secret string, udpConn *net.UDPConn, concurrentCount int) (*UDPOverTCPClient, error) {
	client := &UDPOverTCPClient{
		ServerAddr:      serverAddr,
		Secret:          secret,
		UdpConn:         udpConn,
		ConcurrentCount: concurrentCount,
		MyDial:          net.Dial,
	}

	return client, nil
}

func (cl *UDPOverTCPClient) FetchClientId() error {
	c, err := cl.MyDial("tcp", cl.ServerAddr)
	if err != nil {
		return err
	}
	defer c.Close()
	enc := gob.NewEncoder(c)
	D := gob.NewDecoder(c)

	req := &udpovertcp.Handshake{
		Secret:   cl.Secret,
		ClientId: cl.ClientId,
	}
	if err := enc.Encode(req); err != nil {
		return err
	}

	var handshake udpovertcp.Handshake
	err = D.Decode(&handshake)
	if err != nil {
		return err
	}

	cl.ClientId = handshake.ClientId
	return nil
}
