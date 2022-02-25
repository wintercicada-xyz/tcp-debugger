package main

import (
	"fmt"
	"io"
	"net"
	"os"
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
	fmt.Printf("\rConnected to Server, ready to send message\n\n")
	go func() {
		for {
			readFromConn("", conn)
		}
	}()

	go func() {
		for {
			scanfInput(&[]*net.TCPConn{conn})
		}
	}()
}

func serverMode(address net.TCPAddr) {
	var conns []*net.TCPConn
	//mutex := &sync.Mutex{}
	l, err := net.ListenTCP("tcp", &address)
	fmt.Printf("\rWaiting for client...\n")
	if err != nil {
		fmt.Printf("\rTcp listen errors: %v\n", err)
	}
	defer l.Close()
	for {
		conn, err := l.AcceptTCP()
		if err != nil {
			fmt.Printf("\rTcp accept errors: %v\n", err)
			break
		}
		addr := conn.RemoteAddr().String()
		fmt.Printf("Client %v connect\n", addr)
		conns = append(conns, conn)
		go func() {
			for {
				scanfInput(&conns)
			}
		}()
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
			fmt.Printf("\rConnection closed by server, exiting...\n")
			os.Exit(0)
		}
	}
	if reqLen > 0 {
		fmt.Print("\r", "                                                              ")

		fmt.Printf("\r%s%s\n", connMark, buf[:reqLen])
		fmt.Print("> ")
	}
}

func scanfInput(conns *[]*net.TCPConn) {
	if len(*conns) < 1 {
		return
	}
	var input string
	fmt.Printf("> ")
	_, err := fmt.Scanf("%s\n", &input)
	if err != nil {
		fmt.Printf("\rScan user input errors: %v\n", err)
	}
	for _, conn := range *conns {
		conn.Write([]byte(input))
	}
}
