// Copyright 2019 by Donald Wilson. All rights reserved.
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

// iproto runs a demostration of IG17 bc in operation.
package main

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"os"
	"time"
	
	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	cr "github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/gauge"
	"github.com/mum4k/termdash/widgets/text"
	
	"idemo/internal/conf"
	"idemo/internal/consensus"
	"idemo/internal/env"
)

// playType indicates how to play a gauge.
type playType int

const (
	version                  = "v1.0.1"
	playTypePercent playType = iota
	playTypeAbsolute
)

var (
	waitForGauge = make(chan bool)
	
	cf = conf.NewConfig("iproto_config", env.Config())
	
	// Relative size of windows
	gaugeConsensus      = cf.GetInt("gaugeConsensus")
	consensusSGXmonitor = cf.GetInt("consensusSGXmonitor")
	inputBlock          = cf.GetInt("inputBlock")
	inputButtons        = cf.GetInt("inputButtons")
	
	// Consensus window.
	numberOfNodes     = cf.GetInt("numberOfNodes")
	numberOfMoneyBags = cf.GetInt("numberOfMoneyBags")
	consensusDelay    = cf.GetMilliseconds("consensusDelay")
	moneyBagsDelay    = cf.GetMilliseconds("moneyBagsDelay")
	
	// Gauge window
	gaugeDelay    = cf.GetMilliseconds("gaugeDelay")
	endGaugeWait  = cf.GetMilliseconds("endGaugeWait")
	gaugeInterval = cf.GetInt("gaugeInterval")
	
	maxTransactions = cf.GetInt("maxTransactions")
	randFactor      = cf.GetInt("randFactor")
	
	// preconPORT is used by the TCP connect window.
	preconPORT = cf.GetString("TCPconnect")
)

// writeConsensus generates a randomized consensus group every 3 seconds.
func writeConsensus(ctx context.Context, t *text.Text, _ time.Duration, trig chan string) {
	var (
		ctr       = 0
		theLeader = ""
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
		
		writeColorf(t, cell.ColorBlue, "\n CONSENSUS GROUP LEADER: ")
		writeColorf(t, cell.ColorRed, "\n %s\n", theLeader)
		
		select {
		case <-waitForGauge:
			break
		}
		
		writeColorf(t, cell.ColorBlue, "\n VERIFYING BLOCK TRANSACTIONS ")
		writeColorf(t, cell.ColorRed, "%d ", ctr)
		writeColorf(t, cell.ColorRed, "-->\n ")
		
		for i := 0; i < numberOfMoneyBags; i++ {
			writeColorf(t, cell.ColorRed, "💰")
			time.Sleep(moneyBagsDelay)
		}
		trig <- theLeader
	}
}

func maxTransactionsAdjust() int {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return r1.Intn(randFactor)
}

var maxT int

// playGauge continuously changes the displayed percent value on the
// gauge by the step once every delay. Exits when the context expires.
func playGauge(ctx context.Context, g *gauge.Gauge, step int,
		delay time.Duration, pt playType) {
	prog := 0
	maxT = maxTransactions - maxTransactionsAdjust()
	
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
	// Try to connect to listening port before opening the terminal box.
	l, err := net.Listen("tcp", preconPORT)
	if err != nil {
		fmt.Printf("iproto connection error: %v\n", err)
		os.Exit(1)
	}
	
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
		gauge.BorderTitle(" Collecting Infinicoin Trades "),
	)
	if err != nil {
		panic(err)
	}
	
	// Pre Consensus Transaction Monitor
	blockWriteWindow, err := text.New(text.WrapAtWords(), text.RollContent())
	if err != nil {
		panic(err)
	}
	
	// SGX Monitor Window
	softwareMonitorWindow, err := text.New(text.RollContent(), text.WrapAtWords())
	if err != nil {
		panic(err)
	}
	
	title := fmt.Sprintf(" IG17 BLOCKCHAIN DEMO %s - PRESS Q TO QUIT ", version)
	
	// Container Layout.
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
											cr.SplitPercent(inputBlock), // the imput field
										),
									),
									cr.Bottom(
										cr.Border(linestyle.Light),
										cr.BorderTitle(" Blockchain Tail Monitor "),
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
	
	var (
		loggerCH  = make(chan loggerMSG, 10)
		loggerCH2 = make(chan loggerMSG, 10)
		blockCH   = make(chan string)
	)
	
	// Display randomly generated nodes in the 'consensusWindow'.
	go writeConsensus(ctx, consensusWindow, consensusDelay, blockCH)
	// Play the transaction gathering gauge.
	go playGauge(ctx, transactionGauge, gaugeInterval, gaugeDelay, playTypeAbsolute)
	// Logger
	
	go writeLogger(ctx, softwareMonitorWindow, loggerCH)
	go scanEnclave(loggerCH)
	
	go writeLogger(ctx, balanceLogger, loggerCH2)
	go handleBlockchain(blockWriteWindow, blockCH)
	
	go tcpServer(l, balanceLogger, balanceWindow, loggerCH2)
	
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
