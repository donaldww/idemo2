// Copyright 2019 by Donald Wilson. All rights reserved.
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package main

import (
	"time"

	"github.com/mum4k/termdash/cell"
)

// preconScan is a placeholder function that puts a yellow message in the SGX monitor.
func preconScan(loggerCH chan loggerMSG) {
	for {
		loggerCH <- loggerMSG{"Precon Check Failed! Not enough Infinicoin.", cell.ColorYellow}
		time.Sleep(3 * time.Second)
	}
}
