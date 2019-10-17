package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/SuperBadCode/Vizier/internal"
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
	manager, err := internal.NewManager("pipeline", 160)
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
	start := time.Now()
	for i := 0; i < n; i++ {
		manager.Invoke("add", rand.Intn(100)+1)
	}

	for i := 0; i < n; {
		packet := <-output
		if value, ok := packet.Payload.(int); ok {
			fmt.Printf("OUTPUT %+v\n", value)
			i++
		}
	}

	fmt.Printf("completed %+v", time.Since(start))
}
