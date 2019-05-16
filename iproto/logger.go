// Copyright 2019 by Donald Wilson. All rights reserved.
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"time"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/widgets/text"

	"github.com/donaldww/idemo/internal/sgx"
)

// writeLogger logs messages into the SGX monitor widget.
func writeLogger(_ context.Context, t *text.Text, delay_ time.Duration) {
	counter := 0
	for {
		sgx.Scan()
		tNow := time.Now()
		err := sgx.IsValid()
		if counter >= loggerRefresh {
			t.Reset()
			counter = 0
		}
		if err != nil {
			writeColorf(t, cell.ColorRed, " %v\n", err)
		} else {
			writeColorf(t, cell.ColorGreen, " %s: %s\n", time.Date(
				tNow.Year(), tNow.Month(), tNow.Day(), tNow.Hour(), tNow.Minute(),
				tNow.Second(), tNow.Nanosecond(),
				tNow.Location()),
				"IG17-SGX ENCLAVE: Verified.")
		}
		counter++
		sgx.Reset()
		time.Sleep(delay_)
	}
}
