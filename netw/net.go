package netw

import (
	"net/netip"
	"encoding/binary"
)

//	Converts an IPv4 string to a uint32
func Ipv4_int(s string) uint32 {
	addr, err := netip.ParseAddr(s)
	if err != nil || !addr.Is4() {
		return 0
	}
	b := addr.As4()
	return binary.BigEndian.Uint32(b[:])
}

//	Converts a uint32 to an IPv4 string
func Int_ipv4(i uint32) string {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], i)
	return netip.AddrFrom4(b).String()
}