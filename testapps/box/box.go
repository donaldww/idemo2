// Copyright 2019 by Donald Wilson. All rights reserved.
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package main

import "github.com/rivo/tview"

func main() {
	box := tview.NewBox().
		SetBorder(true).
		SetTitle("IG17 Blockchain Demo")
	if err := tview.NewApplication().SetRoot(box, true).Run(); err != nil {
		panic(err)
	}

	app := tview.NewApplication()
	modal := tview.NewModal().
		SetText("Do you want to quit the application?").
		AddButtons([]string{"Quit", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Quit" {
				app.Stop()
			}
		})
	if err := app.SetRoot(modal, false).SetFocus(modal).Run(); err != nil {
		panic(err)
	}
	if err := tview.NewApplication().SetRoot(box, true).Run(); err != nil {
		panic(err)
	}
}
