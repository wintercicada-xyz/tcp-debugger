package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"net"
	"os"
)

const HelpText = `
Tcp debugger

Usage: td [options][<host>:<port>]

Options
  -c                   Client mode
  -s                   Server mode
  -h                   Print usage
  -H                   HEX mode
  -m threadcount       Multithreaded mode (Client mode only)

`

type Mode bool

const (
	Client Mode = true
	Server Mode = false
)

type ParseResult struct {
	tcpAddr net.TCPAddr
	mode    Mode
	HEX     bool
	thread  int
}

func (parseRes ParseResult) Handle() {
	if parseRes.mode == Client {
		clientMode(parseRes.tcpAddr, parseRes.HEX, parseRes.thread)
	} else {
		serverMode(parseRes.tcpAddr, parseRes.HEX)
	}
}

func flagParse() ParseResult {
	res := ParseResult{
		mode:   Client,
		HEX:    false,
		thread: 1,
	}
	isClientMode := flag.Bool("c", false, "Client mode")
	isServerMode := flag.Bool("s", false, "Server mode")
	isHEX := flag.Bool("H", false, "HEX mode")
	isHelp := flag.Bool("h", false, "Print usage")
	multithread := flag.Int("m", 1, "Multithreaded mode (Client mode only)")
	flag.Parse()
	if isHelp != nil && *isHelp {
		fmt.Print(HelpText)
		os.Exit(0)
	}
	if isServerMode == nil || !*isServerMode {
		res.mode = Client
	} else if *isServerMode && isClientMode == nil || !*isClientMode {
		res.mode = Server
	} else {
		fmt.Println("Can't run server mode and client mode in the same time")
		os.Exit(1)
	}
	if isHEX != nil {
		res.HEX = *isHEX
	}
	if multithread != nil {
		if *multithread <= 0 {
			fmt.Println("Thread count must be positive")
			os.Exit(1)
		}
		res.thread = *multithread
	}
	if res.mode == Server && res.thread > 1 {
		fmt.Println("Multithreaded mode can only be used in client mode")
		os.Exit(1)
	}

	// try to resolve last arg to TCPAddr
	arg := flag.Arg(0)
	tcpAddr, err := net.ResolveTCPAddr("tcp", arg)
	if err != nil {
		fmt.Println("Illegal tcp address format: " + arg)
	}
	res.tcpAddr = *tcpAddr
	return res
}

func readInput(ch chan []byte, isHexMode bool, doAfterInput func()) {
	for {
		var input string
		_, err := fmt.Scanln(&input)
		if err != nil {
			fmt.Println("\rFailed to scan user input")
		} else {
			if isHexMode {
				res, err := hex.DecodeString(input)
				if err != nil {
					fmt.Println(err)
				} else {
					ch <- res
				}
			} else {
				ch <- []byte(input)
			}
		}
		doAfterInput()
	}
}

func writeMessage(ch chan Message, isHexMode bool, handleConnCloseFn func(string)) {
	for {
		msg := <-ch
		if len(msg.msg) == 0 { // conn close
			handleConnCloseFn(msg.addr)
		} else {
			if isHexMode {
				fmt.Printf("\r%s> % X\n> ", msg.addr, msg.msg)
			} else {
				fmt.Printf("\r%s> %s\n> ", msg.addr, msg.msg)
			}
		}
	}
}
