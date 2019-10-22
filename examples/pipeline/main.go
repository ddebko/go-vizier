package main

import (
	"fmt"
	"math/rand"
	"time"

	vizier "github.com/SuperBadCode/go-vizier/pkg"
)

/*
	This example creates a pipeline, each state has a unidirectional edge to another state.
	Essentially creating a linked list. The starting state is called add & the output state is divide.

	Graph
	------
	(s:add) -> (s:subtract) -> (s:multiply) -> (s:divide)
*/

func add(payload interface{}) map[string]interface{} {
	sendTo := make(map[string]interface{})
	if value, ok := payload.(int); ok {
		sendTo["add_to_subtract_step_one"] = value + rand.Intn(100)
	}
	return sendTo
}

func subtract(payload interface{}) map[string]interface{} {
	sendTo := make(map[string]interface{})
	if value, ok := payload.(int); ok {
		sendTo["subtract_to_multiply_step_two"] = value - rand.Intn(100)
	}
	return sendTo
}

func multiply(payload interface{}) map[string]interface{} {
	sendTo := make(map[string]interface{})
	if value, ok := payload.(int); ok {
		sendTo["multiply_to_divide_step_three"] = value * rand.Intn(100)
	}
	return sendTo
}

func divide(payload interface{}) map[string]interface{} {
	sendTo := make(map[string]interface{})
	if value, ok := payload.(int); ok {
		sendTo["output"] = value / (rand.Intn(100) + 1)
	}
	return sendTo
}

func main() {
	manager, err := vizier.NewManager("pipeline", 160)
	if err != nil {
		panic(err)
	}

	manager.Node("add", add).
		Node("subtract", subtract).
		Edge("add", "subtract", "step_one").
		Node("multiply", multiply).
		Edge("subtract", "multiply", "step_two").
		Node("divide", divide).
		Edge("multiply", "divide", "step_three")

	output := manager.Output("divide", "output")

	err = manager.Start()
	if err != nil {
		panic(err)
	}

	n := 1000
	batch := make([]interface{}, n)
	start := time.Now()
	for i := 0; i < n; i++ {
		batch[i] = rand.Intn(100) + 1
	}
	wg, err := manager.BatchInvoke("add", batch)
	if err != nil {
		panic(err)
	}
	results := manager.GetResults(wg, n, output)
	for i := 0; i < n; i++ {
		fmt.Printf("DEBUG %+v\n", results[i])
	}

	fmt.Printf("completed %+v", time.Since(start))
}
