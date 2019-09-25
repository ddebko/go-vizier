package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/SuperBadCode/Vizier/vizier"
)

/*
	This example creates a pipeline, each state has a unidirectional edge to another state.
	Essentially creating a linked list. The starting state is called add & the output state is divide.

	Graph
	------
	(s:add) -> (s:subtract) -> (s:multiply) -> (s:divide)
*/

func main() {
	manager, err := vizier.NewManager("pipeline", 80)
	if err != nil {
		panic(err)
	}

	edgeStart, err := manager.CreateEdge("start")
	if err != nil {
		panic(err)
	}

	edgeSubtract, err := manager.CreateEdge("subtract")
	if err != nil {
		panic(err)
	}

	edgeMultiply, err := manager.CreateEdge("multiply")
	if err != nil {
		panic(err)
	}

	edgeDivide, err := manager.CreateEdge("divide")
	if err != nil {
		panic(err)
	}

	edgeOutput, err := manager.CreateEdge("output")
	if err != nil {
		panic(err)
	}

	stateAdd := vizier.NewState("add", func(payload interface{}) interface{} {
		if value, ok := payload.(int); ok {
			return value + rand.Intn(100)
		}
		return 0
	})
	stateAdd.AttachEdge("from_add_to_subtract", edgeStart, edgeSubtract)

	stateSubtract := vizier.NewState("subtract", func(payload interface{}) interface{} {
		if value, ok := payload.(int); ok {
			return value - rand.Intn(100)
		}
		return 1
	})
	stateSubtract.AttachEdge("from_subtract_to_multiply", edgeSubtract, edgeMultiply)

	stateMultiply := vizier.NewState("multiply", func(payload interface{}) interface{} {
		if value, ok := payload.(int); ok {
			return value * rand.Intn(100)
		}
		return 2
	})
	stateMultiply.AttachEdge("from_multiply_to_divide", edgeMultiply, edgeDivide)

	stateDivide := vizier.NewState("divide", func(payload interface{}) interface{} {
		if value, ok := payload.(int); ok {
			return value / (rand.Intn(100) + 1)
		}
		return 3
	})
	stateDivide.AttachEdge("from_divide_to_output", edgeDivide, edgeOutput)

	manager.CreateState("add", stateAdd)
	manager.CreateState("subtract", stateSubtract)
	manager.CreateState("multiply", stateMultiply)
	manager.CreateState("divide", stateDivide)

	err = manager.Pool.Create()
	if err != nil {
		panic(err)
	}

	n := 1000
	start := time.Now()
	for i := 0; i < n; i++ {
		stateAdd.Send("from_add_to_subtract", rand.Intn(100)+1)
	}

	for i := 0; i < n; i++ {
		out := <-edgeOutput
		if value, ok := out.Payload.(int); ok {
			fmt.Printf("OUTPUT %+v\n", value)
		}
	}
	fmt.Printf("completed %+v", time.Since(start))
}
