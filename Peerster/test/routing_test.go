package test

import (
	"go.dedis.ch/onet/log"
	"net"
	"strconv"
	"testing"
)

func TestUdpConn(t *testing.T) {
	//sanity check of if 05000 <=> 5000
	port := 1
	s := strconv.Itoa(port)
	addr1 := "127.0.0.1:"
	udpAddr, err := net.ResolveUDPAddr("udp4", addr1+s)
	if err != nil {
		log.Error(err)
		t.Fail()
		return
	}

	if udpAddr.Port != port {
		t.Error("Unexpected port")
		t.Fail()
		return
	}
}
