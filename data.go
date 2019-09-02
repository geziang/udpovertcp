package udpovertcp

import "net"

type Handshake struct {
	Secret   string
	ClientId string
}

type UDPPacket struct {
	ClientId string
	SrcAddr  *net.UDPAddr
	DestAddr *net.UDPAddr
	Data     []byte
}
