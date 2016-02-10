//
//  message.go
//
//  Created by Adrian Zubarev.
//  Copyright © 2016 Adrian Zubarev.
//  All rights reserved.
//

package graph

import . "fmt"
import "time"
import "math/rand"

type Graph []Edge
type Edge []Vertex
type Vertex int

var seed *rand.Rand

func init() {

	// initialize global time based seed
	seed = rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
}

func CreateRandomGraph(max int) Graph {

	var indexSlice Graph

	for i := 0; i < max; i++ {

		for j := i + 1; j < max; j++ {

			indexSlice = append(indexSlice, Edge{Vertex(i), Vertex(j)})
		}
	}

	var numberToRemove = seed.Intn(len(indexSlice)/2) + 1

	for i := 0; i < numberToRemove; i++ {

		var indexToRemove = seed.Intn(len(indexSlice))
		indexSlice = append(indexSlice[:indexToRemove], indexSlice[indexToRemove+1:]...)
	}

	return indexSlice
}

func LogGraph(graph *Graph) {

	// log all graph edges
	Println("[Log] [Comment]: Use the following code at https://develop.open.wolframcloud.com -> 'Create a New Notebook'")
	Printf("[Log] [Graph] [Code]: GraphPlot[{")

	var length = len(*graph)

	for index, edge := range *graph {

		var separator = "}, VertexLabeling -> True]\n"

		if index < (length - 1) {
			separator = ", "
		}
		Printf("%d -> %d%s", edge[0], edge[1], separator)
	}
}
