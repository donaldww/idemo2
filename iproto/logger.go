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
	
	"idemo/internal/sgx"
)

type loggerMSG struct {
	msg   string
	color cell.Color
}

// writeLogger logs messages into the SGX monitor widget.
func writeLogger(_ context.Context, t *text.Text, loggerCH chan loggerMSG) {
	counter := 0
	loggerRefresh := cf.GetInt("loggerRefresh")
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

func scanEnclave(loggerCH chan loggerMSG) {
	loggerDelay := cf.GetMilliseconds("loggerDelay")
	for {
		sgx.Scan()
		if err := sgx.IsValid(); err != nil {
			loggerCH <- loggerMSG{msg: fmt.Sprintf("%v", err), color: cell.ColorRed}
		} else {
			loggerCH <- loggerMSG{msg: "IG17-SGX ENCLAVE: Verified.", color: cell.ColorGreen}
		}
		sgx.Reset()
		time.Sleep(loggerDelay)
	}
}
