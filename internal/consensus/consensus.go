// Copyright 2019 by Donald Wilson. All rights reserved.
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package consensus

import (
	"math/rand"
	"time"
)

// NodeID is a list of Node identifiers.
type nodeID struct {
	Node     string
	IsLeader bool
}

var consGrp []nodeID

// NewGroup returns a random list of consGrp used for consensus.
func NewGroup(nuNodes int) *[]nodeID {
	consGrp = nil
	list := randList(nuNodes)
	leader := randLeader(len(list))
	for key, node := range list {
		if key == leader {
			consGrp = append(consGrp, nodeID{NodeIds[node], true})
		} else {
			consGrp = append(consGrp, nodeID{NodeIds[node], false})
		}
	}
	return &consGrp
}

// CurrentGroup returns a pointer to the current consensus group.
func CurrentGroup() *[]nodeID {
	return &consGrp
}

// GenLeader returns an 'int' that represents the consensus leader.
func randLeader(n int) int {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return r1.Intn(n)
}

// Returns an array of random numbers of 'n' length.
func randList(n int) []int {
	max := len(NodeIds)
	result := make([]int, n)
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

	for i := 0; i < n; i++ {
		result[i] = r1.Intn(max)
	}
	return result
}
