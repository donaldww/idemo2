// Copyright 2019 by Donald Wilson. All rights reserved.
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"time"

	"github.com/donaldww/idemo/internal/sgx"
)

func main() {
	quit := make(chan bool)
	start := make(chan bool)

	go func() {
		for {
			select {
			case <-start:
				fmt.Println("\n  Received start signal.")
			case <-quit:
				fmt.Println("\n  Received stop signal.")
				return
			default:
				sgx.Scan()
				// TODO: call function to compare enclave results here.
				// sgx.PrintScanned()
				sgx.Reset()

				time.Sleep(2 * time.Second)

			}
		}
	}()

	start <- false
	time.Sleep(3 * time.Second)
	quit <- true
}
