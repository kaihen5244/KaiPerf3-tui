package iperf

import (
	"fmt"
	"net"
	"strings"
)

// LocalAddr is one IP address bound to a local network interface.
type LocalAddr struct {
	Iface string
	IP    string
}

// LocalIPs enumerates this host's non-loopback IP addresses (IPv4 and
// global-unicast IPv6) across all "up" interfaces, so the UI can tell the
// user what address to hand out without them running ifconfig/ipconfig.
func LocalIPs() ([]LocalAddr, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var out []LocalAddr
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
				continue
			}
			out = append(out, LocalAddr{Iface: iface.Name, IP: ip.String()})
		}
	}
	return out, nil
}

// Summary renders local addresses as a compact one-line string, e.g.
// "en0 192.168.1.23 · en0 fd00::1 · utun4 10.8.0.2", or a placeholder if
// none were found.
func LocalIPsSummary() string {
	addrs, err := LocalIPs()
	if err != nil || len(addrs) == 0 {
		return "no network addresses found"
	}
	parts := make([]string, len(addrs))
	for i, a := range addrs {
		parts[i] = fmt.Sprintf("%s %s", a.Iface, a.IP)
	}
	return strings.Join(parts, " · ")
}
