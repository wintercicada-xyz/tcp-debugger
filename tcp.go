package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"sync"
)

func clientMode(address net.TCPAddr, isHEX bool, connCount int) {
	pool := CreatePool()
	rch := make(chan Message)
	ich := make(chan []byte)
	go pool.HandleWriteToAll(ich)
	go writeMessage(rch, isHEX, func(addr string) {
		pool.DeleteConn(addr)
		if !pool.IsEmpty() {
			fmt.Print("\r> ")
		} else {
			fmt.Print("\r \r")
			os.Exit(0)
		}
	})
	for i := 0; i < connCount; i++ {
		conn, err := net.DialTCP("tcp", nil, &address)
		if err != nil {
			fmt.Printf("\rTcp dial errors: %v\n", err)
			continue
		}
		addr := conn.LocalAddr().String()
		fmt.Printf("\rConnect to server as %v\n", addr)
		fmt.Print("\r> ")
		myConn := MyConn{conn, make(chan []byte)}
		go myConn.HandleWrite()
		mark := conn.RemoteAddr().String()
		if connCount > 1 {
			mark = conn.LocalAddr().String()
		}
		go pool.AddConn(myConn, mark)
		go myConn.HandleReceive(rch, mark)
	}
	go readInput(ich, isHEX, func() {
		if !pool.IsEmpty() {
			fmt.Print("\r> ")
		} else {
			fmt.Print("\r \r")
			os.Exit(0)
		}
	})
}

func serverMode(address net.TCPAddr, isHEX bool) {
	l, err := net.ListenTCP("tcp", &address)
	if err != nil {
		fmt.Printf("\rTcp listen errors: %v\n", err)
	}
	defer l.Close()

	pool := CreatePool()
	rch := make(chan Message)
	ich := make(chan []byte)
	go readInput(ich, isHEX, func() {
		if !pool.IsEmpty() {
			fmt.Print("\r> ")
		} else {
			fmt.Print("\r")
		}
	})
	go pool.HandleWriteToAll(ich)
	go writeMessage(rch, isHEX, func(addr string) {
		pool.DeleteConn(addr)
		if !pool.IsEmpty() {
			fmt.Print("\r> ")
		} else {
			fmt.Print("\r  \r")
		}
	})
	for {
		conn, err := l.AcceptTCP()
		if err != nil {
			fmt.Printf("\rTcp accept errors: %v\n", err)
			continue
		}
		addr := conn.RemoteAddr().String()
		fmt.Printf("\rClient %v connect\n", addr)
		fmt.Print("\r> ")
		myConn := MyConn{conn, make(chan []byte)}
		go myConn.HandleWrite()
		mark := conn.RemoteAddr().String()
		go pool.AddConn(myConn, mark)
		go myConn.HandleReceive(rch, mark)
	}
}

type Pool struct {
	conns map[string]MyConn
	lock  *sync.RWMutex
}

func CreatePool() Pool {
	return Pool{
		make(map[string]MyConn),
		&sync.RWMutex{},
	}

}

func (pool Pool) IsEmpty() bool {
	pool.lock.RLock()
	res := len(pool.conns) == 0
	pool.lock.RUnlock()
	return res
}

func (pool Pool) AddConn(conn MyConn, mark string) {
	pool.lock.Lock()
	pool.conns[mark] = conn
	pool.lock.Unlock()
}
func (pool Pool) DeleteConn(addr string) {
	pool.lock.Lock()
	pool.conns[addr].c.Close()
	delete(pool.conns, addr)
	pool.lock.Unlock()
}

func (pool Pool) HandleWriteToAll(ch chan []byte) {
	for {
		msg := <-ch
		pool.lock.RLock()
		for _, c := range pool.conns {
			c.ch <- msg
		}
		pool.lock.RUnlock()
	}
}

type MyConn struct {
	c  *net.TCPConn
	ch chan []byte
}

type Message struct {
	addr string
	msg  []byte
}

func (conn MyConn) HandleWrite() {
	for msg := range conn.ch {
		conn.c.Write(msg)
	}
}

func (conn MyConn) HandleReceive(ch chan Message, mark string) {
	buf := make([]byte, 1024)
	for {
		reqLen, err := conn.c.Read(buf)
		if err != nil {
			if err == net.ErrClosed || err == io.EOF {
				fmt.Printf("\rConnection to %s closed\n", mark)
				ch <- Message{mark, []byte{}}
				break
			}
		}
		if reqLen > 0 {
			ch <- Message{mark, buf[:reqLen]}
		}
	}
}
