package main

import (
	"net"
	"time"
	"yotamura/common"
)

func GetLocalServerAddress() string {
	yotamuraServerIp := "localhost" //default ip
	go common.ServeMulticastUDP(common.MulticastAddress, func(u *net.UDPAddr, i int, b []byte) bool {
		if string(b) == "yotamuraServer" {
			yotamuraServerIp = u.IP.String()
			return true //will stop the servemulticast function
		}
		return false
	})

	common.WriteUDP(common.MulticastAddress, "yotamuraClient")
	timer := time.NewTimer(time.Second)
	<-timer.C

	return yotamuraServerIp
}
