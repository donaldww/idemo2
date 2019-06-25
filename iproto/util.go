// Copyright 2019 by Donald Wilson. All rights reserved.
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/widgets/text"
)

// WriteColorf adds terminal Color and Sprintf parameters to the Write method.
//
// Params:
//  color: a cell.Color, such as cell.ColorRed, cell.ColorDefault,
//  ... [termdash/cell/color.go]
//  format: a Printf/Sprintf-style format string
//  args: an optional list of comma-separated arguments (varags)
func writeColorf(t *text.Text, color cell.Color, format string, args ...interface{}) {
	_ = t.Write(fmt.Sprintf(format, args...), text.WriteCellOpts(cell.FgColor(color)))
}
