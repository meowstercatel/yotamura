package common

import (
	"log"
	"net"
)

const (
	MulticastAddress = "224.0.0.2:9999"
	maxDatagramSize  = 8192
)

func WriteUDP(address string, content string) {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		log.Fatal(err)
	}
	c, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Println("failed to write to udp", err)
		return
	}
	c.Write([]byte(content))
}

func ServeMulticastUDP(a string, h func(*net.UDPAddr, int, []byte) bool) {
	addr, err := net.ResolveUDPAddr("udp", a)
	if err != nil {
		log.Fatal(err)
	}
	l, err := net.ListenMulticastUDP("udp", nil, addr)
	if err != nil {
		log.Println("failed to listen to udp", err)
		return
	}
	l.SetReadBuffer(maxDatagramSize)
	for {
		b := make([]byte, maxDatagramSize)
		n, src, err := l.ReadFromUDP(b)
		if err != nil {
			log.Fatal("ReadFromUDP failed:", err)
		}
		if h(src, n, b) {
			break
		}
	}
}
