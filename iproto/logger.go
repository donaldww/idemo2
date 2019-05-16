// Copyright 2019 by Donald Wilson. All rights reserved.
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/widgets/text"

	"github.com/donaldww/idemo/internal/sgx"
)

type loggerMSG struct {
	msg   string
	color cell.Color
}

var loggerCH = make(chan loggerMSG, 10)

// writeLogger logs messages into the SGX monitor widget.
func writeLogger(_ context.Context, t *text.Text) {
	counter := 0
	for {
		select {
		case log := <-loggerCH:
			if counter >= loggerRefresh {
				t.Reset()
				counter = 0
			}
			tNow := time.Now()
			writeColorf(t, log.color, " %s: %s\n",
				time.Date(
					tNow.Year(), tNow.Month(), tNow.Day(),
					tNow.Hour(), tNow.Minute(), tNow.Second(), tNow.Nanosecond(),
					tNow.Location(),
				),
				log.msg,
			)
			counter++
		}
	}
}

func enclaveScan(delay time.Duration) {
	for {
		sgx.Scan()
		if err := sgx.IsValid(); err != nil {
			loggerCH <- loggerMSG{fmt.Sprintf("%v", err), cell.ColorRed}
		} else {
			loggerCH <- loggerMSG{msg: "IG17-SGX ENCLAVE: Verified.", color: cell.ColorGreen}
		}
		sgx.Reset()
		time.Sleep(delay)
	}
}
