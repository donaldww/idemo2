// Copyright 2019 by Donald Wilson. All rights reserved.
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package term

import (
	"fmt"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/widgets/text"
)

// WriteColorf adds color and formatting parameters the Write function.
func WriteColorf(t *text.Text, color cell.Color, format string, args ...interface{}) {
	_ = t.Write(fmt.Sprintf(format, args...), text.WriteCellOpts(cell.FgColor(color)))
}
