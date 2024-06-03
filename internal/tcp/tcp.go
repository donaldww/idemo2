package tcp

import (
	"bufio"
	"fmt"
	"github.com/donaldww/idemo2/internal/config"
	"github.com/donaldww/idemo2/internal/logger"
	"github.com/donaldww/idemo2/internal/term"
	"io"
	"net"
	"strconv"
	"strings"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/widgets/text"
)

var balance int

// Reset the balance before updating the balance window.
func reload(t *text.Text, bal int) {
	balance = bal
	update(t)
}

func update(t *text.Text) {
	t.Reset()
	term.WriteColorf(t, cell.ColorCyan, "\n Balance: ")
	term.WriteColorf(t, cell.ColorRed, "%d", balance)
}

func Server(l net.Listener, t *text.Text, b *text.Text, loggerCH chan logger.MSG, cf *config.Config) {
	openBalance := cf.GetInt("openBal")
	reload(b, openBalance)
	defer func(l net.Listener) {
		err := l.Close()
		if err != nil {
			panic(err)
		}
	}(l)
WAITING:
	t.Reset()
	msg := fmt.Sprintf("Waiting for connection...")
	loggerCH <- logger.MSG{Msg: msg, Color: cell.ColorYellow}
	c, err := l.Accept()
	if err != nil {
		t.Reset()
		loggerCH <- logger.MSG{Msg: "Problem with node connection.", Color: cell.ColorYellow}
		goto WAITING
	}
	t.Reset()
	loggerCH <- logger.MSG{Msg: "Node connected.", Color: cell.ColorYellow}
	for {
		netData, thisErr := bufio.NewReader(c).ReadString('\n')
		if thisErr == io.EOF {
			t.Reset()
			loggerCH <- logger.MSG{Msg: "Node connection closed.", Color: cell.ColorYellow}
			goto WAITING
		}
		// '\n' must be trimmed from netData because ReadString() doesn't strip
		// the EOL character for you.
		cmd := strings.Split(strings.TrimRight(netData, "\n"), " ")
		switch len(cmd) {
		case 2:
			amt, thisErr2 := strconv.Atoi(cmd[1])
			if thisErr2 != nil {
				_, _ = c.Write([]byte("enclave-sim: second parameter must be a number.\n"))
				break
			}
			switch cmd[0] {
			case "sell":
				if balance-amt < 0 {
					_, _ = c.Write([]byte("enclave-sim: trade blocked: insufficient funds!\n"))
					logMsg := fmt.Sprintf("%s order: %d IC: BLOCKED!", cmd[0], amt)
					t.Reset()
					loggerCH <- logger.MSG{Msg: logMsg, Color: cell.ColorRed}
				} else {
					balance -= amt
					update(b)
					logMsg := fmt.Sprintf("%s order: %d IC.", cmd[0], amt)
					t.Reset()
					loggerCH <- logger.MSG{Msg: logMsg, Color: cell.ColorYellow}
					logMsg = fmt.Sprintf("enclave-sim: sold: %d coins.\n", amt)
					_, _ = c.Write([]byte(logMsg))
				}
			case "buy":
				balance += amt
				update(b)
				logMsg := fmt.Sprintf("%s order: %d IC.", cmd[0], amt)
				t.Reset()
				loggerCH <- logger.MSG{Msg: logMsg, Color: cell.ColorYellow}
				logMsg = fmt.Sprintf("enclave-sim: bought: %d coins.\n", amt)
				_, _ = c.Write([]byte(logMsg))
			default:
				_, _ = c.Write([]byte("enclave-sim: invalid command: must be 'buy' or 'sell'.\n"))
			}
		case 1:
			switch cmd[0] {
			case "bal":
				balMsg := fmt.Sprintf("enclave-sim: current balance: %d IC.\n", balance)
				_, _ = c.Write([]byte(balMsg))
			case "reload":
				reload(b, openBalance)
				t.Reset()
				loggerCH <- logger.MSG{Msg: "reload.",
					Color: cell.ColorYellow}
				_, _ = c.Write([]byte("enclave-sim: account reloaded.\n"))
			default:
				_, _ = c.Write([]byte("enclave-sim: invalid command.\n"))
			}
		default:
			_, _ = c.Write([]byte("enclave-sim: too many parameters.\n"))
		}
	}
}
