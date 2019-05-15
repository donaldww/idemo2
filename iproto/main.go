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
	"unicode"

	"github.com/donaldww/ig"
	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/button"
	"github.com/mum4k/termdash/widgets/gauge"
	"github.com/mum4k/termdash/widgets/text"
	"github.com/mum4k/termdash/widgets/textinput"

	"github.com/donaldww/idemo/internal/consensus"
	"github.com/donaldww/idemo/internal/sgx"
)

// playType indicates how to play a gauge.
type playType int

const (
	version                  = "v0.5.0"
	playTypePercent playType = iota
	playTypeAbsolute
)

var (
	config       = ig.NewConfig("iproto_config")
	waitForGauge = make(chan bool)
)

var (
	buttonHeight = config.GetInt("buttonHeight")

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

	// SGX monitor widget (logger)
	loggerDelay   = config.GetMilliseconds("loggerDelay")
	loggerRefresh = config.GetInt("loggerRefresh")

	// Gauge widget
	gaugeDelay    = config.GetMilliseconds("gaugeDelay")
	endGaugeWait  = config.GetMilliseconds("endGaugeWait")
	gaugeInterval = config.GetInt("gaugeInterval")

	maxTransactions = config.GetInt("maxTransactions")
	randFactor      = config.GetInt("randFactor")
)

//TODO: Implement auto-load function for config file values.
//TODO: Add pre-consensus check into the remaining two windows.
//TODO: Re-write writeLogger as a general purpose logger that receives
// messages using buffered channels.

// writeLogger logs messages into the SGX monitor widget.
func writeLogger(_ context.Context, t *text.Text, delay_ time.Duration) {
	counter := 0
	for {
		sgx.Scan()
		tNow := time.Now()
		err := sgx.IsValid()
		if counter >= loggerRefresh {
			t.Reset()
			counter = 0
		}
		if err != nil {
			writeColorf(t, cell.ColorRed, " %v\n", err)
		} else {
			writeColorf(t, cell.ColorGreen, " %s: %s\n", time.Date(
				tNow.Year(), tNow.Month(), tNow.Day(), tNow.Hour(), tNow.Minute(),
				tNow.Second(), tNow.Nanosecond(),
				tNow.Location()),
				"IG17-SGX ENCLAVE: Verified.")
		}
		counter++
		sgx.Reset()
		time.Sleep(delay_)
	}
}

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
		writeColorf(t, cell.ColorRed, "%s\n", ldr)

		select {
		case <-waitForGauge:
			break
		}

		writeColorf(t, cell.ColorBlue, "\n WRITING BLOCK ")
		writeColorf(t, cell.ColorRed, "%d ", ctr)
		writeColorf(t, cell.ColorRed, "--> ")

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

	// The input field.
	inputT, err := textinput.New(
		textinput.Label(" Sell: ", cell.FgColor(cell.ColorBlue)),
		textinput.MaxWidthCells(10),
		textinput.Filter(unicode.IsDigit),
	)
	if err != nil {
		panic(err)
	}

	// The Buttons.
	sendB, err := button.New("Send", func() error {
		//TODO: add submit action here
		// updateText <- input.ReadAndClear()
		return nil
	},
		button.Height(buttonHeight),
		button.GlobalKey(keyboard.KeyEnter),
		button.FillColor(cell.ColorNumber(220)),
	)

	reloadB, err := button.New("Reload", func() error {
		inputT.ReadAndClear()
		//TODO: what does the clear button do?
		// updateText <- ""
		return nil
	},
		button.Height(buttonHeight),
		button.WidthFor("Submit"),
		button.FillColor(cell.ColorNumber(220)),
	)

	// quitB, err := button.New("Quit", func() error {
	// 	cancel()
	// 	return nil
	// },
	// 	button.Height(buttonHeight),
	// 	button.WidthFor("Submit"),
	// 	button.FillColor(cell.ColorNumber(196)),
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
		gauge.BorderTitle(" Processing Infinicoin Transactions "),
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
	c, err := container.New(
		t,
		container.Border(linestyle.Light),
		container.BorderColor(cell.ColorDefault),
		container.BorderTitle(title),
		container.SplitHorizontal(
			container.Top(
				container.PlaceWidget(transactionGauge),
			),
			container.Bottom(
				container.SplitHorizontal(
					container.Top(
						container.SplitVertical(
							container.Left(
								container.Border(linestyle.Light),
								container.BorderTitle(" IG17 Consensus Group Randomizer "),
								container.PlaceWidget(consensusWindow),
							),
							container.Right(
								container.SplitHorizontal(
									container.Top(
										container.SplitVertical(
											container.Left(
												container.PlaceWidget(inputT),
											),
											container.Right(
												container.SplitVertical(
													container.Left(
														container.PlaceWidget(sendB),
													),
													container.Right(
														container.PlaceWidget(reloadB),
													),
												),
											),
											container.SplitPercent(inputBlock), // the imput field
										),
									),
									container.Bottom(
										container.Border(linestyle.Light),
										container.BorderTitle(" Block Creation Monitor "),
										container.PlaceWidget(blockWriteWindow),
									),
									container.SplitPercent(inputButtons),
								),
							),
						),
					),
					container.Bottom(
						container.Border(linestyle.Light),
						container.BorderTitle(" SGX Security Monitor "),
						container.PlaceWidget(softwareMonitorWindow),
					),
					container.SplitPercent(consensusSGXmonitor),
				),
			),
			container.SplitPercent(gaugeConsensus),
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
	go writeLogger(ctx, softwareMonitorWindow, loggerDelay)

	// Exit handler.
	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == 'q' || k.Key == 'Q' {
			cancel() // generated by contextWithCancel()
		}
	}

	if thisErr := termdash.Run(ctx, t, c, termdash.KeyboardSubscriber(
		quitter)); thisErr != nil {
		panic(thisErr)
	}
}

// writeColorf adds terminal Color and Sprintf parameters to the Write method.
//
// Params:
//  color: a cell.Color, such as cell.ColorRed, cell.ColorDefault,
//  ... [termdash/cell/color.go]
//  format: a Printf/Sprintf-style format string
//  args: an optional list of comma-separated arguments (varags)
//
func writeColorf(t *text.Text, color cell.Color, format string, args ...interface{}) {
	_ = t.Write(fmt.Sprintf(format, args...), text.WriteCellOpts(cell.FgColor(color)))
}
