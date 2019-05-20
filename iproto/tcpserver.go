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

var openBalance = cf.GetInt("openBal")
var balance int

// Reset the balance before updating the balance window.
func reload(t *text.Text) {
	balance = openBalance
	update(t)
}

func update(t *text.Text) {
	t.Reset()
	writeColorf(t, cell.ColorCyan, "\n Balance: ")
	writeColorf(t, cell.ColorRed, "%d", balance)
}

func tcpServer(t *text.Text, b *text.Text, loggerCH chan loggerMSG) {
	reload(b)
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
		loggerCH <- loggerMSG{"Problem with client connection.", cell.ColorYellow}
		goto WAITING
	}

	t.Reset()
	loggerCH <- loggerMSG{"Client connected.", cell.ColorYellow}

	for {
		netData, thisErr := bufio.NewReader(c).ReadString('\n')
		if thisErr == io.EOF {
			t.Reset()
			loggerCH <- loggerMSG{"Client connection closed.", cell.ColorYellow}
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
			case "sell":
				if balance-amt < 0 {
					_, _ = c.Write([]byte("iproto: trade blocked: insufficient funds!\n"))
					logMsg := fmt.Sprintf("%s order: %d IC: BLOCKED!", cmd[0], amt)
					t.Reset()
					loggerCH <- loggerMSG{msg: logMsg, color: cell.ColorRed}
				} else {
					balance -= amt
					update(b)
					logMsg := fmt.Sprintf("%s order: %d IC.", cmd[0], amt)
					t.Reset()
					loggerCH <- loggerMSG{msg: logMsg, color: cell.ColorYellow}
					logMsg = fmt.Sprintf("iproto: sold: %d coins.\n", amt)
					_, _ = c.Write([]byte(logMsg))
				}
			case "buy":
				balance += amt
				update(b)
				logMsg := fmt.Sprintf("%s order: %d IC.", cmd[0], amt)
				t.Reset()
				loggerCH <- loggerMSG{msg: logMsg, color: cell.ColorYellow}
				logMsg = fmt.Sprintf("iproto: bought: %d coins.\n", amt)
				_, _ = c.Write([]byte(logMsg))
			default:
				_, _ = c.Write([]byte("iproto: invalid command: must be 'buy' or 'sell'.\n"))

			}
		case 1:
			switch cmd[0] {
			case "bal":
				msg := fmt.Sprintf("iproto: current balance: %d IC.\n", balance)
				_, _ = c.Write([]byte(msg))
			case "reload":
				reload(b)
				t.Reset()
				loggerCH <- loggerMSG{msg: "reload.",
					color: cell.ColorYellow}
				_, _ = c.Write([]byte("iproto: account reloaded.\n"))
			default:
				_, _ = c.Write([]byte("iproto: invalid command.\n"))
			}
		default:
			_, _ = c.Write([]byte("iproto: too many parameters.\n"))
		}
	}
}
