package server

import (
	"encoding/gob"
	"fmt"
	"github.com/geziang/udpovertcp"
	"github.com/google/uuid"
	"log"
	"net"
)

type UDPOverTCPServer struct {
	mConn            *ConnMap
	secret           string
	overrideDestAddr *net.UDPAddr
	serverAddr       string
}

/*
 serverAddr 监听tcp地址
 secret 简易鉴权密码
 overrideDestAddr 覆盖目标地址，设置此项将忽略客户端传来的目标地址
*/
func NewUDPOverTCPServer(serverAddr, secret string, overrideDestAddr *net.UDPAddr) *UDPOverTCPServer {
	return &UDPOverTCPServer{
		mConn:            NewConnMap(),
		secret:           secret,
		overrideDestAddr: overrideDestAddr,
		serverAddr:       serverAddr,
	}
}

func (s *UDPOverTCPServer) Serve() {
	l, err := net.Listen("tcp", s.serverAddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		go s.handleConnection(c)
	}
}

func (s *UDPOverTCPServer) handleConnection(c net.Conn) {
	defer c.Close()
	enc := gob.NewEncoder(c)
	D := gob.NewDecoder(c)

	var handshake udpovertcp.Handshake
	err := D.Decode(&handshake)
	if err != nil {
		return
	}

	if handshake.Secret == s.secret {
		if handshake.ClientId == "" {
			id, err := uuid.NewUUID()
			if err != nil {
				log.Println(err)
				return
			}
			resp := &udpovertcp.Handshake{
				ClientId: id.String(),
			}
			if err := enc.Encode(resp); err != nil {
				log.Println(err)
				return
			}
		}
	} else {
		log.Println("invalid secret:", handshake.Secret)
		return
	}

	var proxyConn *UDPProxyConn
	defer func() {
		if proxyConn != nil {
			s.mConn.ReturnProxyConn(proxyConn)
		}
	}()

	for {
		var packet udpovertcp.UDPPacket
		err := D.Decode(&packet)
		if err != nil {
			break
		}

		if proxyConn == nil {
			proxyConn = s.mConn.LoadOrNewProxyConn(packet.ClientId, packet.SrcAddr)
			go func() {
				for {
					udpPacket, ok := proxyConn.Recv()
					if !ok {
						break
					}

					if err := enc.Encode(udpPacket); err != nil {
						log.Println(err)
						break
					}
				}
			}()
		}

		destAddr := packet.DestAddr
		if s.overrideDestAddr != nil {
			destAddr = s.overrideDestAddr
		}
		_, err = proxyConn.Send(packet.Data, destAddr)
		if err != nil {
			log.Println(err)
			break
		}
	}
}
