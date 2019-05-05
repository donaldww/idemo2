package main

import (
	"fmt"
	"time"

	"github.com/donaldww/idemo/internal/consensus"
)

// IG17 Consensus Group Randomizer
// -------------------------------
//
// - A new consensus group and leader is
// randomly generated for every group of
// transactions added to the IG17 blockchain.
//
// - The Randomizer is protected by the SGX
// trusted execution environment (TEE).
//
// - IG17 uses RAFT consensus algorithm because
// it supports fault-tolerance and high performance.`)

const numberOfNodes = 15

func main() {
	for {
		go func() {
			nodes := consensus.NewGroup(numberOfNodes)
			for _, x := range *nodes {
				if x.IsLeader {
					fmt.Println("--->", "LEADER:", x.Node)
				} else {
					fmt.Println(x.Node)
				}
			}
			fmt.Println()
		}()
		time.Sleep(3 * time.Second)
	}
}
