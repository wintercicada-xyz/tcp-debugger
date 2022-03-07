package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
)

func handleParseResult(parseRes ParseResult) error {
	flags := parseRes.flags

	if len(flags) == 0 {
		return errors.New("td should run in client mode or server mode")
	}

	_, hexFlag := flags[HexFlag]
	_, clientFlag := flags[ClientFlag]
	_, serverFlag := flags[ServerFlag]

	if clientFlag && serverFlag {
		return errors.New("can't run in client mode and server mode in the same time")
	}

	var flag Flag
	if hexFlag {
		flag = HexFlag
	}
	if clientFlag {
		clientM(parseRes.tcpAddr, flag, 1)
	} else if serverFlag {
		serverM(parseRes.tcpAddr, flag)
	}
	return nil
}

func clientM(address net.TCPAddr, mode Flag, connCount int) {
	pool := CreatePool()
	rch := make(chan Message)
	ich := make(chan []byte)
	go readInput(ich, mode == HexFlag, func() {
		if !pool.IsEmpty() {
			fmt.Print("\r> ")
		} else {
			fmt.Print("\r \r")
		}
	})
	go pool.HandleWriteToAll(ich)
	go writeMessage(rch, mode == HexFlag, func(addr string) {
		pool.DeleteConn(addr)
		if !pool.IsEmpty() {
			fmt.Print("\r> ")
		} else {
			fmt.Print("\r \r")
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
		go pool.AddConn(myConn, conn.LocalAddr().String())
		go myConn.HandleReceive(rch)
	}
	for {
		select {}
	}
}

func serverM(address net.TCPAddr, mode Flag) {
	l, err := net.ListenTCP("tcp", &address)
	if err != nil {
		fmt.Printf("\rTcp listen errors: %v\n", err)
	}
	defer l.Close()

	pool := CreatePool()
	rch := make(chan Message)
	ich := make(chan []byte)
	go readInput(ich, mode == HexFlag, func() {
		if !pool.IsEmpty() {
			fmt.Print("\r> ")
		} else {
			fmt.Print("\r")
		}
	})
	go pool.HandleWriteToAll(ich)
	go writeMessage(rch, mode == HexFlag, func(addr string) {
		pool.DeleteConn(addr)
		if !pool.IsEmpty() {
			fmt.Print("\r> ")
		} else {
			fmt.Print("\r  ")
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
		go pool.AddConn(myConn, conn.RemoteAddr().String())
		go myConn.HandleReceive(rch)
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
	return len(pool.conns) == 0
}

func (pool Pool) AddConn(conn MyConn, mark string) {
	pool.lock.Lock()
	pool.conns[mark] = conn
	pool.lock.Unlock()
}
func (pool Pool) DeleteConn(addr string) {
	pool.lock.Lock()
	delete(pool.conns, addr)
	pool.lock.Unlock()
}

func (pool Pool) HandleWriteToAll(ch chan []byte) {
	for {
		msg := <-ch
		pool.lock.RLock()
		//fmt.Println(pool.conns)
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

func (conn MyConn) HandleReceive(ch chan Message) {
	addr := conn.c.RemoteAddr().String()
	buf := make([]byte, 1024)
	for {
		reqLen, err := conn.c.Read(buf)
		if err != nil {
			if err == net.ErrClosed || err == io.EOF {
				fmt.Printf("\rConnection to %s closed\n", addr)
				ch <- Message{addr, []byte{}}
				break
			}
		}
		if reqLen > 0 {
			ch <- Message{addr, buf[:reqLen]}
		}
	}
}
