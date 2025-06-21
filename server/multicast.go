package main

import (
	"net"
	"yotamura/common"
)

func StartMulticastListener() {
	common.ServeMulticastUDP(common.MulticastAddress, func(u *net.UDPAddr, i int, b []byte) bool {
		if string(b) == "yotamuraClient" {
			common.WriteUDP(common.MulticastAddress, "yotamuraServer")
		}
		return false
	})
}
