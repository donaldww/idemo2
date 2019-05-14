// Copyright 2019 by Donald Wilson. All rights reserved.
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/rivo/tview"
)

func main() {
	app := tview.NewApplication()

	flx := tview.NewFlex()

	form := newForm(app)

	flx.AddItem(form, 0, 1, true)

	if err := app.SetRoot(flx, true).Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}

func newForm(app *tview.Application) *tview.Form {
	f := tview.NewForm()
	f.SetTitle(" PRE-CONSENSUS CHECK")

	f.AddInputField("Amount:", "", 20, isValidAmount, nil)
	f.AddButton("roger → pontius", nil)
	f.AddButton("pontius → roger", nil)
	f.AddButton("Quit", func() {
		app.Stop()
	})
	return f
}

// isValidAmount only allows numbers into the field.
func isValidAmount(text string, _ rune) bool {
	_, err := strconv.Atoi(text)
	if err != nil {
		return false
	}
	return true
}
