package server

import (
	"log"
	"net"
	"sync"
)

type ConnMap struct {
	mProxyConn sync.Map
}

func NewConnMap() *ConnMap {
	connMap := &ConnMap{}
	return connMap
}

func (m *ConnMap) LoadOrNewProxyConn(clientId string, srcAddr *net.UDPAddr) *UDPProxyConn {
	key := clientId + srcAddr.String()
	v, ok := m.mProxyConn.Load(key)
	var conn *UDPProxyConn
	if ok {
		conn = v.(*UDPProxyConn)
	} else {
		var err error
		conn, err = NewUDPProxyConn(srcAddr, clientId)
		if err != nil {
			log.Println(err)
			return nil
		}
		m.mProxyConn.Store(key, conn)
	}

	conn.IncAccessCounter()
	return conn
}

func (m *ConnMap) ReturnProxyConn(conn *UDPProxyConn) {
	key := conn.clientId + conn.srcAddr.String()

	conn.DecAccessCounter()

	count := conn.GetAccessCounter()
	if count <= 0 {
		conn.Close()
		m.mProxyConn.Delete(key)
	}

	conn = nil
}
