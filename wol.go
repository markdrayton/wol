package main

import (
	"bytes"
	"encoding/binary"
	"log"
	"net"
	"net/netip"
	"os"
	"path/filepath"
)

var preamble = []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

func broadcastAddr(prefix netip.Prefix) (netip.Addr, bool) {
	if prefix.Addr().Is4() {
		mask := uint32(1<<(32-prefix.Bits()) - 1)
		broadcast := make([]byte, 4)
		binary.BigEndian.PutUint32(broadcast, binary.BigEndian.Uint32(prefix.Masked().Addr().AsSlice())|mask)
		if addr, ok := netip.AddrFromSlice(broadcast); ok {
			return addr, true
		}
	}
	return netip.Addr{}, false
}

func allBroadcastAddrs() (map[netip.Addr]string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	broadcasts := map[netip.Addr]string{}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp != net.FlagUp || iface.Flags&net.FlagBroadcast != net.FlagBroadcast {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}

		for _, addr := range addrs {
			prefix, err := netip.ParsePrefix(addr.String())
			if err != nil {
				return nil, err
			}
			if bAddr, ok := broadcastAddr(prefix); ok {
				broadcasts[bAddr] = iface.Name
			}
		}
	}

	return broadcasts, nil
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("usage: %s mac-address\n", filepath.Base(os.Args[0]))
	}

	hwaddr, err := net.ParseMAC(os.Args[1])
	if err != nil {
		log.Fatalf("couldn't parse MAC: %v", err)
	}
	msg := append(preamble, bytes.Repeat(hwaddr, 16)...)

	addrs, err := allBroadcastAddrs()
	if err != nil {
		log.Fatalf("couldn't determine broadcast addresses: %v", err)
	}

	for addr, iface := range addrs {
		log.Printf("waking %s via %s (interface %s)\n", hwaddr, addr, iface)
		ua, err := net.ResolveUDPAddr("udp", netip.AddrPortFrom(addr, 9).String())
		if err != nil {
			log.Fatalf("couldn't resolve target address: %v", err)
		}

		conn, err := net.DialUDP("udp", nil, ua)
		if err != nil {
			log.Fatalf("couldn't connect: %v", err)
		}
		defer conn.Close()

		if _, err := conn.Write(msg); err != nil {
			log.Fatalf("couldn't write: %v", err)
		}
	}
}
