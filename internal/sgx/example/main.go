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
				sgx.PrintScanned()
				// root, err := sgx.InfiniBin()
				// if err != nil {
				// 	log.Panic(err)
				// }
				// result, err2 := sgx.Md5All(root)
				// if err2 != nil {
				// 	log.Panic(err2)
				// }
				//
				// var paths []string
				// for path := range result {
				// 	paths = append(paths, path)
				// }
				// sort.Strings(paths)
				// for _, path := range paths {
				// 	fmt.Printf("%x  %s\n", result[path], path)
				// }
				// fmt.Println()
				time.Sleep(2 * time.Second)

			}
		}
	}()

	start <- false
	time.Sleep(21 * time.Second)
	quit <- true
}
