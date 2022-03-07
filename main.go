package main

import (
	"fmt"
	"os"
)

func main() {
	parseRes, err := parseArgs(os.Args)
	if err != nil {
		fmt.Printf("\r%s\n", err.Error())
		os.Exit(1)
	}
	err = handleParseResult(parseRes)
	if err != nil {
		fmt.Printf("\r%s\n", err.Error())
		os.Exit(1)
	}
	for {
		select {}
	}

}
