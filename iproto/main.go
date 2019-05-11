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

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/gauge"
	"github.com/mum4k/termdash/widgets/text"

	"github.com/donaldww/idemo/internal/consensus"
	"github.com/donaldww/idemo/internal/sgx"
)

const version = "v0.3.2"

// playType indicates how to play a gauge.
type playType int

const (
	playTypePercent playType = iota
	playTypeAbsolute
)

const (
	// Relative sizes of windows on left side of screen
	splitPercent = 15

	// Consensus widget
	numberOfNodes     = 21
	numberOfMoneyBags = 17
	consensusDelay    = 1500 * time.Millisecond

	// SGX monitor widget (logger)
	loggerDelay   = 2000 * time.Millisecond
	loggerRefresh = 5

	// Gauge widget
	gaugeDelay    = 1 * time.Millisecond
	endGaugeWait  = 500 * time.Millisecond
	gaugeInterval = 1

	maxTransactions = 2100
	randFactor      = 297
)

var waitForGauge = make(chan bool)

// writeLogger logs messages into the SGX monitor widget.
func writeLogger(ctx context.Context, t *text.Text, delay_ time.Duration) {
	//TODO: Re-write write logger as a general purpose logger that receives
	// messages using buffered channels.
	_ = ctx
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
func writeConsensus(ctx context.Context, t *text.Text, delay time.Duration) {
	consensusCounter := 0
	leader := ""
	_ = delay

	for {
		t.Reset()
		consensusCounter++

		writeColorf(t, cell.ColorBlue, "\n CONSENSUS GROUP WAITING FOR BLOCK: ")
		writeColorf(t, cell.ColorRed, "%d\n\n", consensusCounter)

		select {
		default:
			nodes := consensus.NewGroup(numberOfNodes)
			for _, x := range *nodes {
				format := fmt.Sprintf(" %s\n", x.Node)
				if x.IsLeader {
					leader = x.Node
				}
				err2 := t.Write(format)
				if err2 != nil {
					panic(err2)
				}
			}
		case <-ctx.Done():
			return
		}

		writeColorf(t, cell.ColorBlue, "\n CONSENSUS GROUP LEADER: ")
		writeColorf(t, cell.ColorRed, "%s\n", leader)

		select {
		case <-waitForGauge:
			break
		}

		writeColorf(t, cell.ColorBlue, "\n WRITING BLOCK ")
		writeColorf(t, cell.ColorRed, "%d ", consensusCounter)
		writeColorf(t, cell.ColorRed, "--> ")

		for i := 0; i < numberOfMoneyBags; i++ {
			writeColorf(t, cell.ColorRed, "💰")
			time.Sleep(40 * time.Millisecond)
		}
	}
}

func maxTransactionsAdjust() int {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return r1.Intn(randFactor)
}

// playGauge continuously changes the displayed percent value on the gauge by the
// step once every delay. Exits when the context expires.
func playGauge(ctx context.Context, g *gauge.Gauge, step int, delay time.Duration, pt playType) {
	progress := 0
	var maxTrans = maxTransactions - maxTransactionsAdjust()

	ticker := time.NewTicker(delay)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C: // The delay.
			switch pt {
			case playTypePercent:
				if err := g.Percent(progress); err != nil {
					panic(err)
				}
			case playTypeAbsolute:
				if err := g.Absolute(progress, maxTrans); err != nil {
					panic(err)
				}
			}
			progress += step
			if progress > maxTrans {
				progress = 0
				maxTrans = maxTransactions - maxTransactionsAdjust()
				waitForGauge <- true
				time.Sleep(endGaugeWait)
			}
		case <-ctx.Done():
			return
		}
	}
}

func main() {
	// termbox.New returns a 'termbox' based on
	// the user's default terminal: Terminal or iTerm.
	t, err := termbox.New(termbox.ColorMode(terminalapi.ColorMode256))
	if err != nil {
		panic(err)
	}
	defer t.Close()

	// Returns a context and cancel function.
	ctx, cancel := context.WithCancel(context.Background())

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
	preConsensusWindow, err := text.New(text.WrapAtWords())
	if err != nil {
		panic(err)
	}

	// Pre Consensus Transaction Monitor
	blockWriteWindow, err := text.New(text.WrapAtWords())
	if err != nil {
		panic(err)
	}

	// SGX Monitor Window
	softwareMonitorWindow, err := text.New(text.RollContent(), text.WrapAtWords())
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
		container.SplitVertical(
			// Left Container.
			container.Left(
				container.SplitHorizontal(
					container.Top(
						container.PlaceWidget(transactionGauge),
					),
					container.Bottom(
						container.Border(linestyle.Light),
						container.BorderTitle(" IG17 Consensus Group Randomizer "),
						container.PlaceWidget(consensusWindow),
					),
					container.SplitPercent(splitPercent),
				),
			),

			// Right Container.
			container.Right(
				container.SplitHorizontal(
					container.Top(
						container.SplitHorizontal(
							container.Top(
								container.Border(linestyle.Light),
								container.BorderTitle(" Pre Consensus Transaction Monitor "),
								container.PlaceWidget(preConsensusWindow),
							),
							container.Bottom(container.Border(linestyle.Light),
								container.BorderTitle(" Block Creation Monitor "),
								container.PlaceWidget(blockWriteWindow),
							),
						),
					),
					container.Bottom(
						container.Border(linestyle.Light),
						container.BorderTitle(" SGX Security Monitor "),
						container.PlaceWidget(softwareMonitorWindow),
					), // Bottom
				),
			), // Right
		), // SplitVertical
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
	go playGauge(ctx, transactionGauge, gaugeInterval, gaugeDelay, playTypeAbsolute)
	// Logger
	go writeLogger(ctx, softwareMonitorWindow, loggerDelay)

	// Exit handler.
	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == 'q' || k.Key == 'Q' {
			cancel() // generated by contextWithCancel()
		}
	}

	if err2 := termdash.Run(ctx, t, c, termdash.KeyboardSubscriber(quitter)); err2 != nil {
		panic(err2)
	}
}

// writeColorf adds terminal Color and Sprintf parameters to the Write method.
//
// Params:
//  color: a cell.Color, such as cell.ColorRed, cell.ColorDefault, ... [termdash/cell/color.go]
//  format: a Printf/Sprintf-style format string
//  args: an optional list of comma-separated arguments (varags)
//
func writeColorf(t *text.Text, color cell.Color, format string, args ...interface{}) {
	_ = t.Write(fmt.Sprintf(format, args...), text.WriteCellOpts(cell.FgColor(color)))
}
