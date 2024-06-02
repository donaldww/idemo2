// Copyright 2019 by Donald Wilson. All rights reserved.
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package logger

import (
	"context"
	"fmt"
	"github.com/donaldww/idemo2/internal/config"
	"github.com/donaldww/idemo2/internal/sgx"
	"github.com/donaldww/idemo2/internal/term"
	"time"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/widgets/text"
)

type MSG struct {
	Msg   string
	Color cell.Color
}

// WriteLogger logs messages into the SGX monitor widget.
func WriteLogger(_ context.Context, t *text.Text, loggerCH chan MSG, cf *config.Config) {
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
			term.WriteColorf(t, log.Color, " %s: %s\n",
				time.Date(
					tNow.Year(), tNow.Month(), tNow.Day(),
					tNow.Hour(), tNow.Minute(), tNow.Second(), tNow.Nanosecond(),
					tNow.Location(),
				),
				log.Msg,
			)
			counter++
		}
	}
}

func ScanEnclave(loggerCH chan MSG, cf *config.Config) {
	loggerDelay := cf.GetMilliseconds("loggerDelay")
	for {
		sgx.Scan()
		if err := sgx.IsValid(); err != nil {
			loggerCH <- MSG{Msg: fmt.Sprintf("%v", err), Color: cell.ColorRed}
		} else {
			loggerCH <- MSG{Msg: "SGX SIMULATOR ENCLAVE: Verified.", Color: cell.ColorGreen}
		}
		sgx.Reset()
		time.Sleep(loggerDelay)
	}
}
