package rudp

import (
	"net"
	"sync"
)

func NewListener(conn *net.UDPConn) *RudpListener {
	listen := &RudpListener{conn: conn,
		newRudpConn: make(chan *Conn, 1024),
		newRudpErr:  make(chan error, 12),
		rudpConnMap: make(map[string]*Conn)}
	go listen.run()
	return listen
}

type RudpListener struct {
	conn *net.UDPConn
	lock sync.RWMutex
	
	newRudpConn chan *Conn
	newRudpErr  chan error
	rudpConnMap map[string]*Conn
}

// net listener interface
func (rl *RudpListener) Accept() (net.Conn, error) { return rl.AcceptRudp() }
func (rl *RudpListener) Close() error {
	rl.CloseAllRudp()
	return rl.conn.Close()
}
func (rl *RudpListener) Addr() net.Addr { return rl.conn.LocalAddr() }

func (rl *RudpListener) CloseRudp(addr string) {
	rl.lock.Lock()
	delete(rl.rudpConnMap, addr)
	rl.lock.Unlock()
}

func (rl *RudpListener) CloseAllRudp() {
	rl.lock.Lock()
	for _, rconn := range rl.rudpConnMap {
		rconn.closef = nil
		rconn.Close()
	}
	rl.lock.Unlock()
}
func (rl *RudpListener) AcceptRudp() (*Conn, error) {
	select {
	case c := <-rl.newRudpConn:
		return c, nil
	case e := <-rl.newRudpErr:
		return nil, e
	}
}
func (rl *RudpListener) run() {
	data := make([]byte, MAX_PACKAGE)
	for {
		n, remoteAddr, err := rl.conn.ReadFromUDP(data)
		if err != nil {
			rl.CloseAllRudp()
			rl.newRudpErr <- err
			return
		}
		rl.lock.RLock()
		rudpConn, ok := rl.rudpConnMap[remoteAddr.String()]
		rl.lock.RUnlock()
		if !ok {
			rudpConn = NewUnConn(rl.conn, remoteAddr, New(), rl.CloseRudp)
			rl.lock.Lock()
			rl.rudpConnMap[remoteAddr.String()] = rudpConn
			rl.lock.Unlock()
			rl.newRudpConn <- rudpConn
		}
		bts := make([]byte, n)
		copy(bts, data[:n])
		rudpConn.in <- bts
	}
}
