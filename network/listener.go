package network

import (
	"ZShadowRotator/utils"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

type Listener struct {
	port        int
	tunnel      *Tunnel
	listener    net.Listener
	connections sync.Map
	stopChan    chan bool
}

func NewListener(port int, tunnel *Tunnel) *Listener {
	return &Listener{
		port:     port,
		tunnel:   tunnel,
		stopChan: make(chan bool),
	}
}

func (l *Listener) Start() error {
	addr := fmt.Sprintf("127.0.0.1:%d", l.port)
	var err error
	l.listener, err = net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start listener: %v", err)
	}
	
	utils.PrintSuccess(fmt.Sprintf("Local proxy listening on %s", addr))
	
	go l.acceptConnections()
	
	return nil
}

func (l *Listener) acceptConnections() {
	for {
		select {
		case <-l.stopChan:
			return
		default:
			conn, err := l.listener.Accept()
			if err != nil {
				continue
			}
			
			go l.handleConnection(conn)
		}
	}
}

func (l *Listener) handleConnection(clientConn net.Conn) {
	defer clientConn.Close()
	
	// Set deadline untuk baca pertama
	clientConn.SetReadDeadline(time.Now().Add(10 * time.Second))
	
	// Read first byte to determine protocol
	buf := make([]byte, 1)
	_, err := clientConn.Read(buf)
	if err != nil {
		return
	}
	
	// Reset deadline
	clientConn.SetReadDeadline(time.Time{})
	
	var target string
	
	// Handle different protocols
	if buf[0] == 0x05 { // SOCKS5
		target, err = l.handleSOCKS5(clientConn)
	} else if buf[0] == 0x04 { // SOCKS4
		target, err = l.handleSOCKS4(clientConn)
	} else {
		// Assume HTTP CONNECT
		target, err = l.handleHTTP(clientConn, buf)
	}
	
	if err != nil || target == "" {
		return
	}
	
	// Dial through tunnel
	remoteConn, err := l.tunnel.Dial(target)
	if err != nil {
		utils.PrintDebug(fmt.Sprintf("Failed to dial %s: %v", target, err))
		return
	}
	defer remoteConn.Close()
	
	// Store connection
	l.connections.Store(clientConn.RemoteAddr().String(), remoteConn)
	defer l.connections.Delete(clientConn.RemoteAddr().String())
	
	// Relay data - bidirectional
	go l.relay(remoteConn, clientConn)
	l.relay(clientConn, remoteConn)
}

func (l *Listener) handleSOCKS5(conn net.Conn) (string, error) {
	// Read auth methods
	buf := make([]byte, 256)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	n, err := conn.Read(buf)
	if err != nil {
		return "", err
	}
	
	// Send auth response (no auth)
	conn.Write([]byte{0x05, 0x00})
	
	// Read request
	n, err = conn.Read(buf)
	if err != nil {
		return "", err
	}
	
	if buf[1] != 0x01 { // Only support CONNECT
		return "", fmt.Errorf("unsupported command")
	}
	
	// Parse address
	var host string
	switch buf[3] {
	case 0x01: // IPv4
		host = net.IP(buf[4:8]).String()
	case 0x03: // Domain name
		addrLen := buf[4]
		host = string(buf[5 : 5+addrLen])
	case 0x04: // IPv6
		host = net.IP(buf[4:20]).String()
	default:
		return "", fmt.Errorf("unsupported address type")
	}
	
	port := int(buf[n-2])<<8 | int(buf[n-1])
	target := fmt.Sprintf("%s:%d", host, port)
	
	// Send success response
	conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	
	return target, nil
}

func (l *Listener) handleSOCKS4(conn net.Conn) (string, error) {
	// Simplified SOCKS4 handling
	return "", fmt.Errorf("SOCKS4 not implemented")
}

func (l *Listener) handleHTTP(conn net.Conn, firstByte []byte) (string, error) {
	// Read the rest of the request
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		return "", err
	}
	
	// Parse HTTP CONNECT
	request := string(append(firstByte, buf[:n]...))
	
	// Check if it's CONNECT method
	if strings.HasPrefix(request, "CONNECT") {
		parts := strings.Split(request, " ")
		if len(parts) >= 2 {
			return parts[1], nil
		}
	}
	
	return "", fmt.Errorf("not a CONNECT request")
}

func (l *Listener) relay(dst, src net.Conn) {
	buf := make([]byte, 32768)
	for {
		src.SetReadDeadline(time.Now().Add(30 * time.Second))
		n, err := src.Read(buf)
		if err != nil {
			return
		}
		dst.SetWriteDeadline(time.Now().Add(30 * time.Second))
		_, err = dst.Write(buf[:n])
		if err != nil {
			return
		}
	}
}

func (l *Listener) Stop() {
	l.stopChan <- true
	if l.listener != nil {
		l.listener.Close()
	}
	
	// Close all connections
	l.connections.Range(func(key, value interface{}) bool {
		if conn, ok := value.(net.Conn); ok {
			conn.Close()
		}
		return true
	})
}
