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
  -c                       client mode
  -l                       listener mode

`

type Flag int

const (
	ClientMode int = iota
	ServerMode
)

type ParseResult struct {
	tcpAddr net.TCPAddr
	flags   []Flag
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
	flags := []Flag{}
	for _, arg := range args[1 : len(args)-1] {
		switch arg {
		case "-c":
			flags = append(flags, Flag(ClientMode))
		case "-l":
			flags = append(flags, Flag(ServerMode))
		}
	}
	if len(flags) > 1 { // "-c" flag and "-l" flag can't coexist
		return parseResult, errors.New("can't run in client mode and server mode in the same time")
	}
	if len(flags) == 0 {
		return parseResult, errors.New("td should run in client mode or server mode")
	}
	parseResult.flags = flags

	return parseResult, nil
}
