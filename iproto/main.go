// Copyright 2019 by Donald Wilson. All rights reserved.
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

// iproto runs a demostration of IG17 blockchain in operation.
package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"
	
	"github.com/donaldww/ig"
	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	cr "github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/gauge"
	"github.com/mum4k/termdash/widgets/text"
	
	"github.com/donaldww/idemo/internal/consensus"
)

// playType indicates how to play a gauge.
type playType int

const (
	version                  = "v0.7.0"
	playTypePercent playType = iota
	playTypeAbsolute
)

var (
	config       = ig.NewConfig("iproto_config")
	waitForGauge = make(chan bool)
)

var (
	//TODO: remove buttonHeight from iproto_config
	// buttonHeight = config.GetInt("buttonHeight")

	// Relative sizes of windows
	gaugeConsensus      = config.GetInt("gaugeConsensus")
	consensusSGXmonitor = config.GetInt("consensusSGXmonitor")
	inputBlock          = config.GetInt("inputBlock")
	inputButtons        = config.GetInt("inputButtons")

	// Consensus widget
	numberOfNodes     = config.GetInt("numberOfNodes")
	numberOfMoneyBags = config.GetInt("numberOfMoneyBags")
	consensusDelay    = config.GetMilliseconds("consensusDelay")
	moneyBagsDelay    = config.GetMilliseconds("moneyBagsDelay")

	// Gauge widget
	gaugeDelay    = config.GetMilliseconds("gaugeDelay")
	endGaugeWait  = config.GetMilliseconds("endGaugeWait")
	gaugeInterval = config.GetInt("gaugeInterval")

	maxTransactions = config.GetInt("maxTransactions")
	randFactor      = config.GetInt("randFactor")
)

//TODO: Implement auto-load function for config file values.
//TODO: Add pre-consensus check into the remaining two windows.

// writeConsensus generates a randomized consensus group every 3 seconds.
func writeConsensus(ctx context.Context, t *text.Text, _ time.Duration) {
	var (
		ctr = 0
		ldr = ""
	)

	for {
		t.Reset()
		ctr++

		writeColorf(t, cell.ColorBlue, "\n CONSENSUS GROUP WAITING FOR BLOCK: ")
		writeColorf(t, cell.ColorRed, "%d\n\n", ctr)

		select {
		default:
			nodes := consensus.NewGroup(numberOfNodes)
			for _, x := range *nodes {
				format := fmt.Sprintf(" %s\n", x.Node)
				if x.IsLeader {
					ldr = x.Node
				}
				err := t.Write(format)
				if err != nil {
					panic(err)
				}
			}
		case <-ctx.Done():
			return
		}

		writeColorf(t, cell.ColorBlue, "\n CONSENSUS GROUP LEADER: ")
		writeColorf(t, cell.ColorRed, "\n %s\n", ldr)

		select {
		case <-waitForGauge:
			break
		}

		writeColorf(t, cell.ColorBlue, "\n WRITING BLOCK ")
		writeColorf(t, cell.ColorRed, "%d ", ctr)
		writeColorf(t, cell.ColorRed, "--> \n")

		for i := 0; i < numberOfMoneyBags; i++ {
			writeColorf(t, cell.ColorRed, "💰")
			time.Sleep(moneyBagsDelay)
		}
	}
}

func maxTransactionsAdjust() int {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return r1.Intn(randFactor)
}

// playGauge continuously changes the displayed percent value on the
// gauge by the step once every delay. Exits when the context expires.
func playGauge(ctx context.Context, g *gauge.Gauge, step int,
	delay time.Duration, pt playType) {
	prog := 0
	var maxT = maxTransactions - maxTransactionsAdjust()

	ticker := time.NewTicker(delay)
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
			}
			prog += step
			if prog > maxT {
				prog = 0
				maxT = maxTransactions - maxTransactionsAdjust()
				waitForGauge <- true
				time.Sleep(endGaugeWait)
			}
		case <-ctx.Done():
			return
		}
	}
}



func main() {
	var err error

	// termbox.New returns a 'termbox' based on
	// the user's default terminal: Terminal or iTerm.
	t, err := termbox.New(termbox.ColorMode(terminalapi.ColorMode256))
	if err != nil {
		panic(err)
	}
	defer t.Close()

	// Returns a context and cancel function.
	ctx, cancel := context.WithCancel(context.Background())

	// Display an account name and balance.
	balanceWindow, err := text.New(text.WrapAtWords())
	if err != nil {
		panic(err)
	}

	// Display an account name and balance.
	balanceLogger, err := text.New()
	if err != nil {
		panic(err)
	}

	// // reloadB is a button that will reload the account with starting balance.
	// reloadB, err := button.New("Reload", func() error {
	// 	reload(balanceWindow)
	// 	return nil
	// },
	// 	button.Height(buttonHeight),
	// 	button.WidthFor("Submit"),
	// 	button.FillColor(cell.ColorNumber(220)),
	// )

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
		gauge.BorderTitle(" Collecting Infinicoin Transactions "),
	)
	if err != nil {
		panic(err)
	}

	// Pre Consensus Transaction Monitor
	blockWriteWindow, err := text.New(text.WrapAtWords())
	if err != nil {
		panic(err)
	}

	// SGX Monitor Window
	softwareMonitorWindow, err := text.New(text.RollContent(),
		text.WrapAtWords())
	if err != nil {
		panic(err)
	}

	title := fmt.Sprintf(" IG17 DEMO %s - PRESS Q TO QUIT ", version)

	// Container Layout.
	c, err := cr.New(
		t,
		cr.Border(linestyle.Light),
		cr.BorderColor(cell.ColorDefault),
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
								cr.BorderTitle(" IG17 Consensus Group Randomizer "),
								cr.PlaceWidget(consensusWindow),
							),
							cr.Right(
								cr.SplitHorizontal(
									cr.Top(
										cr.Border(linestyle.Light),
										cr.BorderColor(cell.ColorCyan),
										cr.BorderTitle(
											" Account: "+config.GetString("accountID")+" "),
										cr.SplitHorizontal(
											cr.Top(
												// cr.SplitVertical(
												// 	cr.Left(
												// 		cr.PlaceWidget(balanceWindow),
												// 	),
												// 	cr.Right(
												// 		cr.PlaceWidget(reloadB),
												// 	),
												// 	cr.SplitPercent(70),
												// ),
												cr.PlaceWidget(balanceWindow),
											),
											cr.Bottom(
												cr.PlaceWidget(balanceLogger),
											),
											cr.SplitPercent(inputBlock), // the imput field
										),
									),
									cr.Bottom(
										cr.Border(linestyle.Light),
										cr.BorderTitle(" Block Monitor "),
										cr.PlaceWidget(blockWriteWindow),
									),
									cr.SplitPercent(inputButtons),
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
					cr.SplitPercent(consensusSGXmonitor),
				),
			),
			cr.SplitPercent(gaugeConsensus),
		),
	)
	if err != nil {
		panic(err)
	}

	// **********
	// GOROUTINES
	// **********

	// Display randomly generated nodes in the 'consensusWindow'.
	go writeConsensus(ctx, consensusWindow, consensusDelay)
	// Play the transaction gathering gauge.
	go playGauge(ctx, transactionGauge, gaugeInterval, gaugeDelay,
		playTypeAbsolute)
	// Logger

	var (
		loggerCH  = make(chan loggerMSG, 10)
		loggerCH2 = make(chan loggerMSG, 10)
	)

	go writeLogger(ctx, softwareMonitorWindow, loggerCH)
	go enclaveScan(loggerCH)

	go writeLogger(ctx, balanceLogger, loggerCH2)
	// go preconScan(loggerCH2)
	go tcpServer(balanceLogger, balanceWindow, loggerCH2)

	

	// Register the exit handler.
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
