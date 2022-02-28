package main

import (
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
