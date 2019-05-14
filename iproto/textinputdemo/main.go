// Copyright 2019 Google Inc.
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

// textinputdemo shows the functionality of a text input field.
package main

import (
	"context"
	"time"
	"unicode"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/container/grid"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/widgets/button"
	"github.com/mum4k/termdash/widgets/textinput"
)

const (
	buttonHeight = 1
)

// func isValidInput(r rune) bool {
// 	return unicode.IsDigit(r)
// }

func main() {
	t, err := termbox.New()
	if err != nil {
		panic(err)
	}
	defer t.Close()

	ctx, cancel := context.WithCancel(context.Background())

	// The input field.
	input, err := textinput.New(
		textinput.Label("Amount: ", cell.FgColor(cell.ColorBlue)),
		textinput.MaxWidthCells(20),
		textinput.Filter(unicode.IsDigit),
	)
	if err != nil {
		panic(err)
	}

	// The Buttons.
	submitB, err := button.New("Submit", func() error {
		//TODO: add submit action here
		// updateText <- input.ReadAndClear()
		return nil
	},
		button.Height(buttonHeight),
		button.GlobalKey(keyboard.KeyEnter),
		button.FillColor(cell.ColorNumber(220)),
	)

	clearB, err := button.New("Clear", func() error {
		input.ReadAndClear()
		//TODO: what does the clear button do?
		// updateText <- ""
		return nil
	},
		button.Height(buttonHeight),
		button.WidthFor("Submit"),
		button.FillColor(cell.ColorNumber(220)),
	)

	quitB, err := button.New("Quit", func() error {
		cancel()
		return nil
	},
		button.Height(buttonHeight),
		button.WidthFor("Submit"),
		button.FillColor(cell.ColorNumber(196)),
	)

	// Make the grid.
	builder := grid.New()

	builder.Add(
		grid.RowHeightPerc(20,
			grid.Widget(
				input,
				container.AlignHorizontal(align.HorizontalCenter),
				container.AlignVertical(align.VerticalBottom),
				container.MarginBottom(1),
			),
		),
	)

	builder.Add(
		grid.RowHeightPerc(40,
			grid.ColWidthPerc(20),
			grid.ColWidthPerc(20,
				grid.Widget(
					submitB,
					container.AlignVertical(align.VerticalTop),
					container.AlignHorizontal(align.HorizontalRight),
				),
			),
			grid.ColWidthPerc(20,
				grid.Widget(
					clearB,
					container.AlignVertical(align.VerticalTop),
					container.AlignHorizontal(align.HorizontalCenter),
				),
			),
			grid.ColWidthPerc(20,
				grid.Widget(
					quitB,
					container.AlignVertical(align.VerticalTop),
					container.AlignHorizontal(align.HorizontalLeft),
				),
			),
			grid.ColWidthPerc(20),
		),
	)

	// Grid is added to the container here.
	gridOpts, err := builder.Build()
	if err != nil {
		panic(err)
	}

	c, err := container.New(t, gridOpts...)
	if err != nil {
		panic(err)
	}

	// Run.
	i := termdash.RedrawInterval(500 * time.Millisecond)
	if thisErr := termdash.Run(ctx, t, c, i); thisErr != nil {
		panic(thisErr)
	}
}
