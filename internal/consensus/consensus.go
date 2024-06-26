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

var consensusGroup []nodeID

// NewGroup returns a randomized list of consensusGroup nodes.
//
// IsLeader=true indicates the group leader, false a regular node.
func NewGroup(nuNodes int) *[]nodeID {
	consensusGroup = nil
	list := randList(nuNodes)
	leader := randLeader(len(list))
	for key, node := range list {
		if key == leader {
			consensusGroup = append(consensusGroup, nodeID{NodeIds[node], true})
		} else {
			consensusGroup = append(consensusGroup, nodeID{NodeIds[node], false})
		}
	}
	return &consensusGroup
}

// Returns a random 'int' that represents the consensus leader.
func randLeader(n int) int {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return r1.Intn(n)
}

// Returns an array of random numbers of size 'n'.
func randList(n int) []int {
	maxLength := len(NodeIds)
	result := make([]int, n)
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

	for i := 0; i < n; i++ {
		result[i] = r1.Intn(maxLength)
	}
	return result
}
