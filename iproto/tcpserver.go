package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"

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
		// '\n' must be trimmed from netData because ReadString() doesn't strip
		// the EOL character.
		cmd := strings.Split(strings.TrimRight(netData, "\n"), " ")

		switch len(cmd) {
		case 2:
			amt, thisErr2 := strconv.Atoi(cmd[1])
			if thisErr2 != nil {
				_, _ = c.Write([]byte("iproto: second parameter must be a number.\n"))
				break
			}
			switch cmd[0] {
			case "buy":
				logMsg := fmt.Sprintf("%s -> %d IC", cmd[0], amt)
				t.Reset()
				loggerCH <- loggerMSG{msg: logMsg, color: cell.ColorYellow}
				_, _ = c.Write([]byte("iproto: bought coins.\n"))
			case "sell":
				logMsg := fmt.Sprintf("%s -> %d IC", cmd[0], amt)
				t.Reset()
				loggerCH <- loggerMSG{msg: logMsg, color: cell.ColorYellow}
				_, _ = c.Write([]byte("iproto: you sold coins.\n"))
			default:
				_, _ = c.Write([]byte("iproto: invalid command: must be 'buy' or 'sell'\n"))

			}
		case 1:
			switch cmd[0] {
			case "bal":
				_, _ = c.Write([]byte("iproto: current balance is ... .\n"))
			case "reload":
				_, _ = c.Write([]byte("iproto: account reloaded.\n"))
			default:
				_, _ = c.Write([]byte("iproto: invalid command.\n"))
			}
		default:
			_, _ = c.Write([]byte("iproto: too many arguments.\n"))
		}
	}
}
