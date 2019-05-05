// Copyright 2019 by Donald Wilson. All rights reserved.
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminal/terminalapi"
)

// redrawInterval determines how often termdash redraws the screen.
const redrawInterval = 250 * time.Millisecond

// rootID is used as the container ID.
const rootID = "root"

func main() {
	// termbox.New returns a 'termbox' based on
	// the user's default terminal: Terminal or iTerm.
	t, err := termbox.New(termbox.ColorMode(terminalapi.ColorMode256))
	if err != nil {
		panic(err)
	}
	defer t.Close()

	// container.New opens a container.
	//
	// Container options are found in termdash/container/options.go.
	c, err := container.New(t, container.ID(rootID))
	if err != nil {
		panic(err)
	}

	// context.WithCancel returns a context and associated cancel function.
	ctx, cancel := context.WithCancel(context.Background())

	// newWidgets sets up the widgets using the context (ctx) and container (c).
	w, err := prototypes.newWidgets(ctx, c)
	if err != nil {
		panic(err)
	}

	lb, err := prototypes.newLayoutButtons(c, w)
	if err != nil {
		panic(err)
	}
	w.buttons = lb

	gridOpts, err := gridLayout(w, layoutAll) // equivalent to contLayout(w)
	if err != nil {
		panic(err)
	}

	if err1 := c.Update(rootID, gridOpts...); err1 != nil {
		panic(err1)
	}

	// quitter processes keyboard input.
	//
	// Key definitions are found in termdash/keyboard/keyboard.go
	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == keyboard.KeyEsc || k.Key == keyboard.KeyCtrlC {
			cancel()
		}
	}

	// Runs the terminal dashboard.
	if err2 := termdash.Run(ctx, t, c, termdash.KeyboardSubscriber(quitter),
		termdash.RedrawInterval(redrawInterval)); err2 != nil {
		panic(err2)
	}
}
