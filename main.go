package main

func main() {
	flagParse().Handle()
	for {
		select {}
	}

}
