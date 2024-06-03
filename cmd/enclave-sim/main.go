// Copyright 2019 by Donald Wilson. All rights reserved.
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

// enclave-sim runs a simulation of a blockchain
// using an enclave to protect the components.
package main

import (
	"context"
	"fmt"
	"github.com/donaldww/idemo2/internal/blockchain"
	"github.com/donaldww/idemo2/internal/logger"
	"github.com/donaldww/idemo2/internal/tcp"
	"github.com/donaldww/idemo2/internal/term"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	cr "github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/gauge"
	"github.com/mum4k/termdash/widgets/text"

	"github.com/donaldww/idemo2/internal/config"
	"github.com/donaldww/idemo2/internal/consensus"
)

// playType indicates how to play a gauge.
type playType int

const (
	version                  = "v1.0.1"
	playTypePercent playType = iota
	playTypeAbsolute
)

// writeConsensus generates a randomized consensus group every 3 seconds.
func writeConsensus(ctx context.Context, t *text.Text, trig chan string, waitForGaugeCH chan bool, cf *config.Config) {
	var (
		ctr       = 0
		theLeader = ""
	)
	for {
		t.Reset()
		ctr++
		term.WriteColorf(t, cell.ColorBlue, "\n CONSENSUS GROUP WAITING FOR BLOCK: ")
		term.WriteColorf(t, cell.ColorRed, "%d\n\n", ctr)
		select {
		default:
			nodes := consensus.NewGroup(cf.GetInt("numberOfNodes"))
			for _, x := range *nodes {
				format := fmt.Sprintf(" %s\n", x.Node)
				if x.IsLeader {
					theLeader = x.Node
				}
				err := t.Write(format)
				if err != nil {
					panic(err)
				}
			}
		case <-ctx.Done():
			return
		}
		term.WriteColorf(t, cell.ColorBlue, "\n CONSENSUS GROUP LEADER: ")
		term.WriteColorf(t, cell.ColorRed, "\n %s\n", theLeader)
		select {
		case <-waitForGaugeCH:
			break
		}
		term.WriteColorf(t, cell.ColorBlue, "\n VERIFYING BLOCK TRANSACTIONS ")
		term.WriteColorf(t, cell.ColorRed, "%d ", ctr)
		term.WriteColorf(t, cell.ColorRed, "-->\n ")
		for i := 0; i < cf.GetInt("numberOfMoneyBags"); i++ {
			term.WriteColorf(t, cell.ColorRed, "💰")
			time.Sleep(cf.GetMilliseconds("moneyBagsDelay"))
		}
		trig <- theLeader
	}
}

func maxTransactionsAdjust(cf *config.Config) int {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return r1.Intn(cf.GetInt("randFactor"))
}

var maxT int

// playGauge continuously changes the displayed percent value on the
// gauge by the step once every delay. Exits when the context expires.
func playGauge(ctx context.Context, g *gauge.Gauge, pt playType, waitForGaugeCH chan bool, cf *config.Config) {
	prog := 0
	maxT = cf.GetInt("maxTransactions") - maxTransactionsAdjust(cf)
	ticker := time.NewTicker(cf.GetMilliseconds("gaugeDelay"))
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C: // The delay.
			switch pt {
			case playTypePercent:
				if err := g.Percent(prog); err != nil {
					panic(err)
				}
			case playTypeAbsolute:
				if err := g.Absolute(prog, maxT); err != nil {
					panic(err)
				}
			default:
				panic("unhandled default case")
			}
			prog += cf.GetInt("gaugeInterval")
			if prog > maxT {
				prog = 0
				maxT = cf.GetInt("maxTransactions") - maxTransactionsAdjust(cf)
				waitForGaugeCH <- true
				time.Sleep(cf.GetMilliseconds("endGaugeWait"))
			}
		case <-ctx.Done():
			return
		}
	}
}

func main() {
	cf := config.NewConfig("enclave_config")
	// Connect to listening port before writing to the terminal box,
	// to avoid a `hung` terminal in the case of log.Fatal(err).
	l, err := net.Listen("tcp", cf.GetString("TCPconnect"))
	if err != nil {
		log.Fatal(err)
	}
	// termbox.New returns a 'termbox' based on
	// the user's default terminal: (e.g. Terminal or iTerm on macOS)
	t, err := termbox.New(termbox.ColorMode(terminalapi.ColorMode256))
	if err != nil {
		panic(err)
	}
	defer t.Close()
	// Adds cancel function to context, used by the quitter function.
	ctx, cancel := context.WithCancel(context.Background())
	balanceWindow, err := text.New(text.WrapAtWords())
	if err != nil {
		panic(err)
	}
	balanceLogger, err := text.New()
	if err != nil {
		panic(err)
	}
	// Consensus Generator Window.
	consensusWindow, err := text.New(text.RollContent(), text.WrapAtWords())
	if err != nil {
		panic(err)
	}
	// Gauge: Transaction Generator Window
	transactionGauge, err := gauge.New(
		gauge.Height(1),
		gauge.Color(cell.ColorBlue),
		gauge.Border(linestyle.Light),
		gauge.BorderTitle(" Collecting Trades "),
	)
	if err != nil {
		panic(err)
	}
	// Pre-Consensus Transaction Monitor
	blockWriteWindow, err := text.New(text.WrapAtWords(), text.RollContent())
	if err != nil {
		panic(err)
	}
	// SGX Monitor Window
	softwareMonitorWindow, err := text.New(text.RollContent(), text.WrapAtWords())
	if err != nil {
		panic(err)
	}
	title := fmt.Sprintf(" ENCLAVE SIMULATER %s - PRESS Q TO QUIT ", version)
	// Container Layout.
	c := container(err, t, title, transactionGauge, consensusWindow, cf, balanceWindow,
		balanceLogger, blockWriteWindow, softwareMonitorWindow)
	// GOROUTINES
	var (
		loggerCH       = make(chan logger.MSG, 10)
		loggerCH2      = make(chan logger.MSG, 10)
		blockCH        = make(chan string)
		waitForGaugeCH = make(chan bool)
	)
	// Display randomly generated nodes in the 'consensusWindow'.
	go writeConsensus(ctx, consensusWindow, blockCH, waitForGaugeCH, cf)
	// Play the transaction gathering gauge.
	go playGauge(ctx, transactionGauge, playTypeAbsolute, waitForGaugeCH, cf)
	go logger.WriteLogger(ctx, softwareMonitorWindow, loggerCH, cf)
	go logger.ScanEnclave(loggerCH, cf)
	go logger.WriteLogger(ctx, balanceLogger, loggerCH2, cf)
	go blockchain.HandleBlockchain(blockWriteWindow, blockCH, maxT)
	go tcp.Server(l, balanceLogger, balanceWindow, loggerCH2, cf)
	// Define the exit handler.
	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == 'q' || k.Key == 'Q' {
			cancel() // generated by contextWithCancel()
		}
	}
	// Run the program.
	if thisErr := termdash.Run(ctx, t, c, termdash.KeyboardSubscriber(quitter)); thisErr != nil {
		panic(thisErr)
	}
}

func container(err error, t *termbox.Terminal, title string, transactionGauge *gauge.Gauge,
	consensusWindow *text.Text, cf *config.Config, balanceWindow *text.Text,
	balanceLogger *text.Text, blockWriteWindow *text.Text, softwareMonitorWindow *text.Text) *cr.Container {
	c, err := cr.New(t,
		cr.Border(linestyle.Light),
		cr.BorderColor(cell.ColorDefault),
		cr.BorderTitleAlignCenter(),
		cr.BorderTitle(title),
		cr.SplitHorizontal(
			cr.Top(
				cr.PlaceWidget(transactionGauge),
			),
			cr.Bottom(
				cr.SplitHorizontal(
					cr.Top(
						cr.SplitVertical(
							cr.Left(
								cr.Border(linestyle.Light),
								cr.BorderTitle(" Consensus Group Randomizer "),
								cr.PlaceWidget(consensusWindow),
							),
							cr.Right(
								cr.SplitHorizontal(
									cr.Top(
										cr.Border(linestyle.Light),
										cr.BorderColor(cell.ColorCyan),
										cr.BorderTitle(
											" Account: "+cf.GetString("accountID")+" "),
										cr.SplitHorizontal(
											cr.Top(
												cr.PlaceWidget(balanceWindow),
											),
											cr.Bottom(
												cr.PlaceWidget(balanceLogger),
											),
											cr.SplitPercent(cf.GetInt("inputBlock")), // the input field
										),
									),
									cr.Bottom(
										cr.Border(linestyle.Light),
										cr.BorderTitle(" Blockchain Tail Monitor "),
										cr.PlaceWidget(blockWriteWindow),
									),
									cr.SplitPercent(cf.GetInt("inputButtons")),
								),
							),
							cr.SplitPercent(40),
						),
					),
					cr.Bottom(
						cr.Border(linestyle.Light),
						cr.BorderTitle(" Enclave Monitor "),
						cr.PlaceWidget(softwareMonitorWindow),
					),
					cr.SplitPercent(cf.GetInt("consensusSGXmonitor")),
				),
			),
			cr.SplitPercent(cf.GetInt("gaugeConsensus")),
		),
	)
	if err != nil {
		panic(err)
	}
	return c
}
