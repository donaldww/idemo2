// Copyright 2019 by Donald Wilson. All rights reserved.
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"time"

	"github.com/donaldww/idemo/internal/sgx"
)

const scanInterval = 2

func main() {
	stop := make(chan bool)
	start := make(chan bool)

	go func() {
		for {
			select {
			case <-start:
				fmt.Println("\n  Received start signal.")
			case <-stop:
				fmt.Println("\n  Received stop signal.")
				return
			default:
				sgx.Scan()
				err := sgx.IsValid()
				if err != nil {
					fmt.Println(err)
				} else {
					fmt.Println("IG17 Enclave Scan: PASSED")
				}
				sgx.Reset()
				time.Sleep(scanInterval * time.Second)
			}
		}
	}()

	start <- true
	time.Sleep(100 * time.Second)
	stop <- true
}
