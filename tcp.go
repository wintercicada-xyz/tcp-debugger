package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"sync"
)

func handleParseResult(parseRes ParseResult) error {
	for _, f := range parseRes.flags {
		switch f {
		case Flag(ClientMode):
			clientMode(parseRes.tcpAddr)

		case Flag(ServerMode):
			serverMode(parseRes.tcpAddr)
		}
	}
	return nil
}

func clientMode(address net.TCPAddr) {
	conn, err := net.DialTCP("tcp", nil, &address)
	if err != nil {
		fmt.Printf("\rTcp dial errors: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("\rConnected to Server with local addr %s\n\n", conn.LocalAddr())
	go func() {
		for {
			readFromConn("", conn)
		}
	}()

	go func() {
		for {
			var input string
			_, err := fmt.Scanln(&input)
			if err != nil {
				fmt.Printf("\rScan user input errors: %v\n", err)
			} else {
				conn.Write([]byte(input))
				fmt.Printf("> ")
			}
		}
	}()
}

func serverMode(address net.TCPAddr) {
	mutex := &sync.RWMutex{}
	var conns []*net.TCPConn
	l, err := net.ListenTCP("tcp", &address)
	fmt.Printf("\rWaiting for client...\n")
	if err != nil {
		fmt.Printf("\rTcp listen errors: %v\n", err)
	}
	defer l.Close()

	go func() {
		for {
			if len(conns) == 0 {
				continue
			}
			var input string
			_, err := fmt.Scanln(&input)
			if err != nil {
				fmt.Printf("\rScan user input errors: %v\n", err)
			} else {
				mutex.RLock()
				for _, conn := range conns {
					conn.Write([]byte(input))
				}
				mutex.RUnlock()
				fmt.Printf("> ")
			}
		}
	}()

	for {
		conn, err := l.AcceptTCP()
		if err != nil {
			fmt.Printf("\rTcp accept errors: %v\n", err)
			break
		}
		addr := conn.RemoteAddr().String()
		fmt.Printf("\rClient %v connect\n", addr)
		fmt.Print("> ")
		mutex.Lock()
		conns = append(conns, conn)
		mutex.Unlock()
		go func() {
			for {
				readFromConn(addr+"> ", conn)
			}
		}()
	}
}

func readFromConn(connMark string, conn *net.TCPConn) {
	buf := make([]byte, 1024)
	reqLen, err := conn.Read(buf)
	if err != nil {
		if err == net.ErrClosed || err == io.EOF {
			fmt.Printf("\rConnection closed, exiting...\n")
			os.Exit(0)
		}
	}
	if reqLen > 0 {
		fmt.Print("\r", "                                                              ")
		fmt.Printf("\r%s%s\n", connMark, buf[:reqLen])
		fmt.Print("> ")
	}
}
