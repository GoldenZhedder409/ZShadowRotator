package network

import (
	"io"
	"net"
	"sync"
	"time"
)

type Forwarder struct {
	connections sync.Map
}

func NewForwarder() *Forwarder {
	return &Forwarder{}
}

func (f *Forwarder) Forward(src, dst net.Conn) {
	defer src.Close()
	defer dst.Close()
	
	// Bidirectional copy
	go f.copyAndClose(dst, src)
	f.copyAndClose(src, dst)
}

func (f *Forwarder) copyAndClose(dst, src net.Conn) {
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

func (f *Forwarder) ForwardWithTimeout(src, dst net.Conn, timeout time.Duration) {
	defer src.Close()
	defer dst.Close()
	
	// Set deadlines
	src.SetDeadline(time.Now().Add(timeout))
	dst.SetDeadline(time.Now().Add(timeout))
	
	// Forward
	go io.Copy(dst, src)
	io.Copy(src, dst)
}

func (f *Forwarder) StoreConnection(key string, conn net.Conn) {
	f.connections.Store(key, conn)
}

func (f *Forwarder) CloseConnection(key string) {
	if val, ok := f.connections.Load(key); ok {
		if conn, ok := val.(net.Conn); ok {
			conn.Close()
		}
		f.connections.Delete(key)
	}
}

func (f *Forwarder) CloseAll() {
	f.connections.Range(func(key, value interface{}) bool {
		if conn, ok := value.(net.Conn); ok {
			conn.Close()
		}
		return true
	})
}
