package network

import (
	"ZShadowRotator/proxy"
	"ZShadowRotator/tor"
        "fmt"
	"net"
	"sync"
	"time"
)

type Tunnel struct {
	torManager   *tor.TorManager
	rotator      *proxy.Rotator
	connections  sync.Map
	stopChan     chan bool
}

func NewTunnel(tm *tor.TorManager, rot *proxy.Rotator) *Tunnel {
	return &Tunnel{
		torManager: tm,
		rotator:    rot,
		stopChan:   make(chan bool),
	}
}

func (t *Tunnel) Dial(dest string) (net.Conn, error) {
	// Get current chain
	chain := t.rotator.GetCurrentChain()
	if chain == nil || len(chain.Hops) == 0 {
		return nil, fmt.Errorf("no active proxy chain")
	}
	
	// First connect to Tor
	torAddr := t.torManager.GetSocksProxy()
	torConn, err := net.DialTimeout("tcp", torAddr, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Tor: %v", err)
	}
	
	// Use Tor to connect to first proxy in chain
	err = t.socks5Handshake(torConn, chain.Hops[0].Address)
	if err != nil {
		torConn.Close()
		return nil, fmt.Errorf("Tor to proxy handshake failed: %v", err)
	}
	
	// Now chain through remaining proxies
	currentConn := torConn
	for i := 1; i < len(chain.Hops); i++ {
		nextConn, err := t.chainProxy(currentConn, chain.Hops[i].Address)
		if err != nil {
			currentConn.Close()
			return nil, fmt.Errorf("chain failed at hop %d: %v", i, err)
		}
		currentConn = nextConn
	}
	
	// Finally connect to destination
	err = t.socks5Handshake(currentConn, dest)
	if err != nil {
		currentConn.Close()
		return nil, fmt.Errorf("final destination handshake failed: %v", err)
	}
	
	// Store connection
	t.connections.Store(currentConn.LocalAddr().String(), currentConn)
	
	return currentConn, nil
}

func (t *Tunnel) socks5Handshake(conn net.Conn, target string) error {
	// SOCKS5 handshake
	conn.Write([]byte{0x05, 0x01, 0x00})
	
	buf := make([]byte, 2)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, err := conn.Read(buf)
	if err != nil {
		return err
	}
	
	if buf[0] != 0x05 || buf[1] != 0x00 {
		return fmt.Errorf("SOCKS5 handshake failed")
	}
	
	// Connect command
	host, port, _ := net.SplitHostPort(target)
	ip := net.ParseIP(host)
	
	var atyp byte
	if ip == nil {
		atyp = 0x03 // Domain name
	} else if ip.To4() != nil {
		atyp = 0x01 // IPv4
	} else {
		atyp = 0x04 // IPv6
	}
	
	// Build request
	req := []byte{0x05, 0x01, 0x00, atyp}
	
	if atyp == 0x03 {
		req = append(req, byte(len(host)))
		req = append(req, []byte(host)...)
	} else if atyp == 0x01 {
		req = append(req, ip.To4()...)
	} else {
		req = append(req, ip.To16()...)
	}
	
	// Add port
	var portNum uint16
	fmt.Sscanf(port, "%d", &portNum)
	req = append(req, byte(portNum>>8), byte(portNum&0xFF))
	
	conn.Write(req)
	
	// Read response
	resp := make([]byte, 4)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, err = conn.Read(resp)
	if err != nil {
		return err
	}
	
	if resp[1] != 0x00 {
		return fmt.Errorf("SOCKS5 connect failed: %d", resp[1])
	}
	
	return nil
}

func (t *Tunnel) chainProxy(conn net.Conn, nextProxy string) (net.Conn, error) {
	// Connect to next proxy
	nextConn, err := net.DialTimeout("tcp", nextProxy, 10*time.Second)
	if err != nil {
		return nil, err
	}
	
	// Start relaying between connections
	go t.relay(conn, nextConn)
	go t.relay(nextConn, conn)
	
	return nextConn, nil
}

func (t *Tunnel) relay(dst, src net.Conn) {
	buf := make([]byte, 32768)
	for {
		src.SetReadDeadline(time.Now().Add(30 * time.Second))
		n, err := src.Read(buf)
		if err != nil {
			return
		}
		dst.SetWriteDeadline(time.Now().Add(30 * time.Second))
		dst.Write(buf[:n])
	}
}

func (t *Tunnel) CloseConnection(addr string) {
	if conn, ok := t.connections.Load(addr); ok {
		conn.(net.Conn).Close()
		t.connections.Delete(addr)
	}
}

func (t *Tunnel) Stop() {
	t.stopChan <- true
	
	// Close all connections
	t.connections.Range(func(key, value interface{}) bool {
		value.(net.Conn).Close()
		return true
	})
}
