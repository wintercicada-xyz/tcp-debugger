package main

import (
	"fmt"
	"os"
)

func main() {
	parseRes, err := parseArgs(os.Args)
	if err != nil {
		fmt.Println("Failed to parse arguments: ", err.Error())
	}
	handleParseResult(parseRes)
	for {
		select {}
	}

}
