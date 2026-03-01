package validator

import (
	"net"
	"strings"
	"time"
)

type ProxyInfo struct {
	Address  string
	Country  string
	Flag     string
	Code     string
	Protocol string
	Latency  time.Duration
	Alive    bool
}

func ValidateProxy(address string, protocol string) (bool, time.Duration) {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	latency := time.Since(start)
	
	if err != nil {
		return false, 0
	}
	defer conn.Close()

	// Test protocol
	if protocol == "SOCKS5" {
		return testSOCKS5(conn), latency
	}
	
	return true, latency
}

func testSOCKS5(conn net.Conn) bool {
	// SOCKS5 handshake
	conn.Write([]byte{0x05, 0x01, 0x00})
	buf := make([]byte, 2)
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, _ := conn.Read(buf)
	return (n == 2 && buf[0] == 0x05)
}

func IsValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

func ExtractHostPort(addr string) (string, string) {
	parts := strings.Split(addr, ":")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", ""
}
