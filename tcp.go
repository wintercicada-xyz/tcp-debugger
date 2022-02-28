package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
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
		clientMode(parseRes.tcpAddr, flag)
	} else if serverFlag {
		serverMode(parseRes.tcpAddr, flag)
	}
	return nil
}

func clientMode(address net.TCPAddr, mode Flag) {
	conn, err := net.DialTCP("tcp", nil, &address)
	if err != nil {
		fmt.Printf("\rTcp dial errors: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("\rConnected to Server with local addr %s\n\n", conn.LocalAddr())
	fmt.Printf("> ")
	go func() {
		for {
			if readFromConn("", conn, mode == HexFlag) != nil {
				os.Exit(1)
			}
		}
	}()

	go func() {
		for {
			input, err := scan(mode == HexFlag)
			if err != nil {
				fmt.Printf("\rScan user input errors: %v\n", err)
			} else {
				conn.Write(input)
			}
			fmt.Printf("> ")
		}
	}()
}

func serverMode(address net.TCPAddr, mode Flag) {
	mutex := &sync.RWMutex{}
	conns := make(map[*net.TCPConn]struct{})
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
			input, err := scan(mode == HexFlag)
			if err != nil {
				fmt.Printf("\rScan user input errors: %v\n", err)
			} else {
				mutex.RLock()
				for conn := range conns {
					conn.Write(input)
				}
				mutex.RUnlock()
			}
			fmt.Printf("> ")
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
		conns[conn] = struct{}{}
		mutex.Unlock()
		go func() {
			for {
				err = readFromConn(addr+"> ", conn, mode == HexFlag)
				if err != nil {
					mutex.Lock()
					delete(conns, conn)
					if len(conns) != 0 {
						fmt.Print("> ")
					}
					mutex.Unlock()
					break
				}

			}
		}()
	}
}

func readFromConn(connMark string, conn *net.TCPConn, isHexMode bool) error {
	buf := make([]byte, 1024)
	reqLen, err := conn.Read(buf)
	if err != nil {
		if err == net.ErrClosed || err == io.EOF {
			fmt.Printf("\rConnection to %s closed\n", conn.RemoteAddr())
			return err
		}
	}
	if reqLen > 0 {
		fmt.Print("\r", "                                                              ")
		if isHexMode {
			fmt.Printf("\r%s% X\n", connMark, buf[:reqLen])
		} else {
			fmt.Printf("\r%s%s\n", connMark, buf[:reqLen])
		}
		fmt.Print("> ")
	}
	return nil
}

func scan(isHexMode bool) ([]byte, error) {
	var input string
	_, err := fmt.Scanln(&input)
	if isHexMode {
		res, err := hex.DecodeString(input)
		return res, err
	} else {
		res := []byte(input)
		return res, err
	}

}
