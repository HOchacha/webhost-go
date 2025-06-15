package libvirt

import (
	"encoding/xml"
	"fmt"
	"net"
)

type networkXML struct {
	IP struct {
		Address string `xml:"address,attr"`
		Netmask string `xml:"netmask,attr"`
	} `xml:"ip"`
}

// example
//
// ipnet, err := manager.GetDefaultNetworkCIDR()
// if err != nil {
//    log.Fatalf("CIDR 조회 실패: %v", err)
// }
// fmt.Println("Libvirt 네트워크 CIDR:", ipnet.String()) // 예: 192.168.122.0/24
//

type LibvirtNetworkConfig struct {
	CIDR    *net.IPNet
	Gateway net.IP
}

// GetDefaultNetworkCIDR retrieves the CIDR range of the default libvirt NAT network.
func (m *LibvirtManager) GetDefaultNetworkConfig() (*LibvirtNetworkConfig, error) {
	network, err := m.conn.NetworkLookupByName("default")
	if err != nil {
		return nil, fmt.Errorf("default 네트워크 조회 실패: %w", err)
	}

	xmlDesc, err := m.conn.NetworkGetXMLDesc(network, 0)
	if err != nil {
		return nil, fmt.Errorf("네트워크 XML 조회 실패: %w", err)
	}

	var netConf networkXML
	if err := xml.Unmarshal([]byte(xmlDesc), &netConf); err != nil {
		return nil, fmt.Errorf("XML 파싱 실패: %w", err)
	}

	ip := net.ParseIP(netConf.IP.Address)
	mask := net.IPMask(net.ParseIP(netConf.IP.Netmask).To4())

	cidr := &net.IPNet{
		IP:   ip.Mask(mask),
		Mask: mask,
	}

	return &LibvirtNetworkConfig{
		CIDR:    cidr,
		Gateway: ip,
	}, nil
}

// GetUsableIPs returns all usable (assignable) IPs from the default libvirt network
func (m *LibvirtManager) GetUsableIPs(used []net.IP) ([]net.IP, error) {
	netConf, err := m.GetDefaultNetworkConfig()
	if err != nil {
		return nil, fmt.Errorf("libvirt 네트워크 설정 조회 실패: %w", err)
	}

	cidr := netConf.CIDR
	gateway := netConf.Gateway
	broadcastIP := lastIP(cidr)

	// 이미 사용 중인 IP → map으로 빠르게 검사
	usedMap := make(map[string]bool)
	for _, ip := range used {
		usedMap[ip.String()] = true
	}

	var usableIPs []net.IP
	for ip := nextIP(cidr.IP.Mask(cidr.Mask)); compareIP(ip, broadcastIP) < 0; ip = nextIP(ip) {
		if ip.Equal(gateway) || usedMap[ip.String()] {
			continue
		}
		usableIPs = append(usableIPs, ip)
	}

	return usableIPs, nil
}

func nextIP(ip net.IP) net.IP {
	ip = ip.To4()
	next := make(net.IP, len(ip))
	copy(next, ip)
	for i := len(next) - 1; i >= 0; i-- {
		next[i]++
		if next[i] != 0 {
			break
		}
	}
	return next
}

func lastIP(n *net.IPNet) net.IP {
	ip := n.IP.To4()
	mask := n.Mask

	broadcast := make(net.IP, len(ip))
	for i := range ip {
		broadcast[i] = ip[i] | ^mask[i]
	}
	return broadcast
}

func compareIP(a, b net.IP) int {
	a = a.To4()
	b = b.To4()
	for i := 0; i < len(a); i++ {
		if a[i] < b[i] {
			return -1
		} else if a[i] > b[i] {
			return 1
		}
	}
	return 0
}
