// Copyright 2019 by Donald Wilson. All rights reserved.
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

// Binary textdemo displays a couple of Text widgets.
// Exist when 'q' or 'Q' is pressed.

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/donaldww/idemo/internal/consensus"
	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/gauge"
	"github.com/mum4k/termdash/widgets/text"
)

// playType indicates how to play a gauge.
type playType int

const (
	playTypePercent playType = iota
	playTypeAbsolute
)

const numberOfNodes = 23
const consensusDelay = 1500 * time.Millisecond

const splitPercent = 15

const gaugeDelay = 1 * time.Millisecond
const endGaugeWait = 500 * time.Millisecond
const gaugeInterval = 1
const maxTransactions = 2000
const version = "v0.2.0"

var waitForGauge = make(chan bool)

// writeConsensus generates a randomized consensus group every 3 seconds.
func writeConsensus(ctx context.Context, t *text.Text, delay time.Duration) {
	consensusCounter := 0
	leader := ""
	_ = delay

	for {
		t.Reset()
		consensusCounter++

		_ = t.Write(fmt.Sprintf("\n CONSENSUS GROUP WAITING FOR BLOCK: "),
			text.WriteCellOpts(cell.FgColor(cell.ColorBlue)))
		_ = t.Write(fmt.Sprintf("%d\n\n", consensusCounter),
			text.WriteCellOpts(cell.FgColor(cell.ColorRed)))

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

		_ = t.Write(fmt.Sprintf("\n CONSENSUS GROUP LEADER: "),
			text.WriteCellOpts(cell.FgColor(cell.ColorBlue)))
		_ = t.Write(fmt.Sprintf("%s\n", leader),
			text.WriteCellOpts(cell.FgColor(cell.ColorRed)))

		select {
		case <-waitForGauge:
			break
		}

		_ = t.Write(fmt.Sprintf("\n WRITING BLOCK "),
			text.WriteCellOpts(cell.FgColor(cell.ColorBlue)))
		_ = t.Write(fmt.Sprintf("%d ", consensusCounter),
			text.WriteCellOpts(cell.FgColor(cell.ColorRed)))
		_ = t.Write(fmt.Sprintf("--> "),
			text.WriteCellOpts(cell.FgColor(cell.ColorRed)))

		for i := 0; i < 21; i++ {
			_ = t.Write(fmt.Sprintf("💰"), // 💰
				text.WriteCellOpts(cell.FgColor(cell.ColorRed)))
			time.Sleep(40 * time.Millisecond)
		}
	}
}

// playGauge continuously changes the displayed percent value on the gauge by the
// step once every delay. Exits when the context expires.
func playGauge(ctx context.Context, g *gauge.Gauge, step int, delay time.Duration, pt playType) {
	progress := 0

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
				if err := g.Absolute(progress, maxTransactions); err != nil {
					panic(err)
				}
			}
			progress += step
			if progress > maxTransactions {
				progress = 0
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
	preConsensusWindow, err := text.New(text.WrapAtRunes())
	if err != nil {
		panic(err)
	}

	// Pre Consensus Transaction Monitor
	blockWriteWindow, err := text.New(text.WrapAtRunes())
	if err != nil {
		panic(err)
	}

	// SGX Monitor Window
	softwareMonitorWindow, err := text.New(text.WrapAtRunes())
	if err != nil {
		panic(err)
	}

	title := fmt.Sprintf(" IG17 DEMO %s - PRESS Q TO QUIT ", version)
	// Container Layout.
	c, err := container.New(
		t,
		container.Border(linestyle.Light),
		// container.BorderColor(cell.ColorBlue),
		container.BorderColor(cell.ColorDefault),
		container.BorderTitle(title),
		container.SplitVertical(
			container.Left(
				container.SplitHorizontal(
					container.Top(
						container.PlaceWidget(transactionGauge),
					),
					container.Bottom(
						container.Border(linestyle.Light),
						// container.BorderColor(cell.ColorYellow),
						container.BorderTitle(" IG17 Consensus Group Randomizer "),
						container.PlaceWidget(consensusWindow),
					),
					container.SplitPercent(splitPercent),
				),
			), // Left

			container.Right(
				container.SplitHorizontal(
					container.Top(
						container.SplitHorizontal(
							container.Top(
								container.Border(linestyle.Light),
								// container.BorderColor(cell.ColorYellow),
								container.BorderTitle(" Pre Consensus Transaction Monitor "),
								container.PlaceWidget(preConsensusWindow),
							),
							container.Bottom(container.Border(linestyle.Light),
								// container.BorderColor(cell.ColorYellow),
								container.BorderTitle(" Block Creation Monitor "),
								container.PlaceWidget(blockWriteWindow),
							),
						),
					),
					container.Bottom(
						container.Border(linestyle.Light),
						// container.BorderColor(cell.ColorYellow),
						container.BorderTitle(" SGX Software Monitor "),
						container.PlaceWidget(softwareMonitorWindow),
					), // Bottom
				),
			), // Right
		), // SplitVertical
	)
	if err != nil {
		panic(err)
	}

	// ******************
	// ACTION GO ROUTINES
	// ******************

	// Write generated nodes into the 'consensusWindow' window.
	go writeConsensus(ctx, consensusWindow, consensusDelay)
	go playGauge(ctx, transactionGauge, gaugeInterval, gaugeDelay, playTypeAbsolute)

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