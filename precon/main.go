package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/donaldww/idemo/internal/conf"
	"github.com/donaldww/idemo/internal/env"
)

func main() {
	negI := flag.String("i", "localhost", "Optional IP address")
	flag.Parse()
	
	CONNECT := func() string {
		i := conf.NewConfig("iproto_config", env.Config())
		if *negI != "localhost" {
			return *negI + ":" + i.GetString("TCPport")
		} else {
			return i.GetString("TCPconnect")
		}
	}()
	
	
	c, err := net.Dial("tcp", CONNECT)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	myTime := time.Now().Format(time.RFC3339)
	fmt.Println("precon version 0.1.0", myTime)
	fmt.Println("Enter 'help' for usage hints.")
	fmt.Println("Connected to IG17 demo server.")

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("precon> ")
		text, _ := reader.ReadString('\n')
		switch strings.TrimSpace(string(text)) {
		case "quit", "q":
			fmt.Println("TCP client exiting...")
			os.Exit(0)
		case "help", "h":
			printHelp()
			break
		default:
			// Send message to server.
			_, _ = fmt.Fprintf(c, text+"\n")
			// Receive message from server.
			message, _ := bufio.NewReader(c).ReadString('\n')
			// Print response.
			fmt.Print("response> " + message)
		}
	}
}

func printHelp() {
	msg := `precon commands:
"buy or sell" <amount>
"reload" to replenish account
"bal" to retrieve current balance
"q or quit" to exit`

	fmt.Println(msg)
}
