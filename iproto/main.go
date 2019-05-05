// Copyright 2018 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	"github.com/mum4k/termdash/widgets/text"
)

const numberOfNodes = 21
const consensusDelay = 2 * time.Second

// writeConsensus generates a randomized consensus group every 3 seconds.
func writeConsensus(ctx context.Context, t *text.Text, delay time.Duration) {
	consensusCounter := 0
	leader := ""
	for {
		t.Reset()
		consensusCounter++
		err := t.Write(fmt.Sprintf("\n CONSENSUS GROUP NO: %d\n\n", consensusCounter),
			text.WriteCellOpts(cell.FgColor(cell.ColorBlue)))
		if err != nil {
			panic(err)
		}
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

		err = t.Write(fmt.Sprintf("\n CONSENSUS GROUP LEADER CHOSEN: %s\n", leader),
			text.WriteCellOpts(cell.FgColor(cell.ColorBlue)))
		if err != nil {
			panic(err)
		}
		time.Sleep(delay)

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

	// Transaction Generator Window
	// TODO: Replace this widget with a guage.
	transactionWindow, err := text.New(text.RollContent(), text.WrapAtWords())
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

	// Container Layout.
	c, err := container.New(
		t,
		container.Border(linestyle.Light),
		// container.BorderColor(cell.ColorBlue),
		container.BorderColor(cell.ColorDefault),
		container.BorderTitle(" IG17 DEMO v0.1.0 - PRESS Q TO QUIT "),
		container.SplitVertical(

			container.Left(
				container.SplitHorizontal(
					container.Top(
						container.Border(linestyle.Light),
						// container.BorderColor(cell.ColorYellow),
						container.BorderTitle(" Random Consensus Group Generator "),
						container.PlaceWidget(consensusWindow),
					),
					container.Bottom(
						// TODO: This widget should be a gauge.
						container.Border(linestyle.Light),
						// container.BorderColor(cell.ColorYellow),
						container.BorderTitle(" Gathering Transactions "),
						container.PlaceWidget(transactionWindow),
					),
					container.SplitPercent(80),
				),
			),
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
							// TODO: Add a bottom-right widget.
						),
					),
					container.Bottom(
						container.Border(linestyle.Light),
						// container.BorderColor(cell.ColorYellow),
						container.BorderTitle(" SGX Software Monitor "),
						container.PlaceWidget(softwareMonitorWindow),
					),
				),
			),
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
