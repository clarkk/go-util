package netw

import (
	"net"
	"encoding/binary"
)

func Ipv4_int(s string) uint32 {
	ipaddr := net.ParseIP(s).To4()
	if ipaddr == nil {
		return 0
	}
	return binary.BigEndian.Uint32(ipaddr)
}

func Int_ipv4(i uint32) string {
	ip := make(net.IP, net.IPv4len)
	binary.BigEndian.PutUint32(ip, i)
	return ip.String()
}