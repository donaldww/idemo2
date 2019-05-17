package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/donaldww/ig"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/widgets/text"
)

func tcpServer(t *text.Text, loggerCH chan loggerMSG) {
	PORT := ig.NewConfig("iproto_config").GetString("TCPconnect")
	l, err := net.Listen("tcp", PORT)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer l.Close()

WAITING:
	t.Reset()
	loggerCH <- loggerMSG{"Waiting for client connection...", cell.ColorYellow}

	c, err := l.Accept()
	if err != nil {
		t.Reset()
		loggerCH <- loggerMSG{"Problem with client connection...", cell.ColorYellow}
		goto WAITING
	}
	t.Reset()
	loggerCH <- loggerMSG{"Client connected...", cell.ColorYellow}

	for {
		netData, thisErr := bufio.NewReader(c).ReadString('\n')
		if thisErr == io.EOF {
			t.Reset()
			loggerCH <- loggerMSG{"Client connection closed...", cell.ColorYellow}
			goto WAITING
		}
		
		t.Reset()
		loggerCH <- loggerMSG{string(netData), cell.ColorYellow}
		t := time.Now()
		myTime := t.Format(time.RFC3339) + "\n"
		_, _ = c.Write([]byte(myTime))
	}
}
