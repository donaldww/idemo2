package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
	
	"idemo/internal/conf"
	"idemo/internal/env"
)

func main() {
	flagI := flag.String("i", "localhost", "Optional IP address")
	flag.Parse()
	
	tcpConnectString := func() string {
		// If the user has entered an IP address on the commandline, then
		// combine that address with the port found in the config file.
		// If the user hasn't over-ridden the default ('localhost'), then
		// use the connect string found in the confin file.
		config := conf.NewConfig("iproto_config", env.Config())
		if *flagI != "localhost" {
			return *flagI + ":" + config.GetString("TCPport")
		} else {
			return config.GetString("TCPconnect")
		}
	}()
	
	connection, err := net.Dial("tcp", tcpConnectString)
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
		switch strings.TrimSpace(text) {
		case "quit", "q":
			fmt.Println("TCP client exiting...")
			os.Exit(0)
		case "help", "h":
			printHelp()
			break
		default:
			// Send message to server.
			_, _ = fmt.Fprintf(connection, text+"\n")
			// Receive response from server.
			message, _ := bufio.NewReader(connection).ReadString('\n')
			// Print response.
			fmt.Print("response> " + message)
		}
	}
}

func printHelp() {
	const msg = `precon commands:
"buy or sell" <amount>
"reload" to replenish account
"bal" to retrieve current balance
"q or quit" to exit`
	
	fmt.Println(msg)
}
