package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/donaldww/idemo/testdata/names"
)

const numberOfNodes = 13

func main() {
	var result = randomList(numberOfNodes, len(names.IDs))
	// fmt.Println(result)
	printHeader()
	printGroup(result, names.IDs)
	fmt.Println()
}

/*
func printAll() {
	for i := 0; i < 500; i++ {
		fmt.Println(i+1, ids[i])
	}
}
*/

func printGroup(searchResult []int, names names.NodeID) {
	leader := randomLeader()
	for i, key := range searchResult {
		if i == leader {
			fmt.Print("    LEADER = ")
		}
		fmt.Println(names[key])
	}
}

func printHeader() {
	fmt.Println()
	
	fmt.Println(`
IG17 Consensus Group Randomizer
-------------------------------

- A new consensus group and leader is
randomly generated for every group of
transactions added to the IG17 blockchain.

- The Randomizer is protected by the SGX
trusted execution environment (TEE).

- IG17 uses RAFT consensus algorithm because
it supports fault-tolerance and high performance.`)
	
	fmt.Println()
}

func randomLeader() int {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return r1.Intn(numberOfNodes)
}

func randomList(n, max int) []int {
	result := make([]int, n)
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

	for i := 0; i < n; i++ {
		result[i] = r1.Intn(max)
	}
	return result
}
