package ipam

import (
	"fmt"
	"math/big"
	"net"
)

func ipToInt(ip net.IP) (*big.Int, int) {
	val := &big.Int{}
	val.SetBytes([]byte(ip))
	if len(ip) == net.IPv4len {
		return val, 32
	} else if len(ip) == net.IPv6len {
		return val, 128
	} else {
		panic(fmt.Errorf("unsupported address length %d", len(ip)))
	}
}

func intToIP(ipInt *big.Int, bits int) net.IP {
	ipBytes := ipInt.Bytes()
	ret := make([]byte, bits/8)
	// Pack our IP bytes into the end of the return array,
	// since big.Int.Bytes() removes front zero padding.
	for i := 1; i <= len(ipBytes); i++ {
		ret[len(ret)-i] = ipBytes[len(ipBytes)-i]
	}
	return net.IP(ret)
}

func addressRange(network *net.IPNet) (net.IP, net.IP) {
	// the first IP is easy
	firstIP := network.IP

	// the last IP is the network address OR NOT the mask address
	prefixLen, bits := network.Mask.Size()
	if prefixLen == bits {
		// Easy!
		// But make sure that our two slices are distinct, since they
		// would be in all other cases.
		lastIP := make([]byte, len(firstIP))
		copy(lastIP, firstIP)
		return firstIP, lastIP
	}

	firstIPInt, bits := ipToInt(firstIP)
	hostLen := uint(bits) - uint(prefixLen)
	lastIPInt := big.NewInt(1)
	lastIPInt.Lsh(lastIPInt, hostLen)
	lastIPInt.Sub(lastIPInt, big.NewInt(1))
	lastIPInt.Or(lastIPInt, firstIPInt)

	return firstIP, intToIP(lastIPInt, bits)
}

func nextSubnet(network *net.IPNet, prefixLen int) (*net.IPNet, bool) {
	_, currentLast := addressRange(network)
	mask := net.CIDRMask(prefixLen, 8*len(currentLast))
	currentSubnet := &net.IPNet{IP: currentLast.Mask(mask), Mask: mask}
	_, last := addressRange(currentSubnet)
	last = incIP(last)
	next := &net.IPNet{IP: last.Mask(mask), Mask: mask}
	if last.Equal(net.IPv4zero) || last.Equal(net.IPv6zero) {
		return next, true
	}
	return next, false
}

func incIP(IP net.IP) net.IP {
	IP = checkIPv4(IP)
	incIP := make([]byte, len(IP))
	copy(incIP, IP)
	for j := len(incIP) - 1; j >= 0; j-- {
		incIP[j]++
		if incIP[j] > 0 {
			break
		}
	}
	return incIP
}

func checkIPv4(ip net.IP) net.IP {
	// Go for some reason allocs IPv6len for IPv4 so we have to correct it
	if v4 := ip.To4(); v4 != nil {
		return v4
	}
	return ip
}

func isTheNextIP(ipToCheck string, previousIP string) bool {
	nextIP := incIP(net.ParseIP(previousIP))
	return nextIP.Equal(net.ParseIP(ipToCheck))
}
