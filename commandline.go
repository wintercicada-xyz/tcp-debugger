package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"os"
)

const HelpText = `
Tcp debugger

Usage: td [options][<host>:<port>]

Options
  -c                       Client mode
  -l                       Server mode
  -h        --help         Print usage
            --hex          HEX mode

`

type Flag int

const (
	ClientFlag Flag = iota
	ServerFlag
	HexFlag
)

type ParseResult struct {
	tcpAddr net.TCPAddr
	flags   map[Flag]string
}

func parseArgs(args []string) (ParseResult, error) {
	parseResult := ParseResult{}

	// check args len to ensure tcp address exist
	if len(args) < 2 {
		return parseResult, errors.New("tcp address not found")
	}

	// help flag
	if args[1] == "-h" {
		fmt.Print(HelpText)
		os.Exit(0)
	}

	// try to resolve last arg to TCPAddr
	tcpAddr, err := net.ResolveTCPAddr("tcp", args[len(args)-1])
	if err != nil {
		return parseResult, errors.New("illegal tcp address format: " + args[len(args)-1])
	}
	parseResult.tcpAddr = *tcpAddr

	// scan flag in the args
	flags := make(map[Flag]string)
	for _, arg := range args[1 : len(args)-1] {
		switch arg {
		case "-c":
			flags[ClientFlag] = ""
		case "-l":
			flags[ServerFlag] = ""
		case "--hex":
			flags[HexFlag] = ""
		default:
			return parseResult, errors.New("illegal flag: " + arg)
		}
	}
	parseResult.flags = flags

	return parseResult, nil
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
