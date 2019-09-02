package main

import (
	"flag"
	"github.com/geziang/udpovertcp/server"
	"log"
	"net"
)

func main() {
	servAddr := flag.String("listen", ":10032", "server listen addr")
	sec := flag.String("secret", "", "secret")
	ovDest := flag.String("overrideDest", "", "override destAddr")
	flag.Parse()

	var overrideDestAddr *net.UDPAddr
	if *ovDest != "" {
		var err error
		overrideDestAddr, err = net.ResolveUDPAddr("udp", *ovDest)
		if err != nil {
			log.Println(err)
			return
		}
	}

	server.NewUDPOverTCPServer(*servAddr, *sec, overrideDestAddr).Serve()
}
