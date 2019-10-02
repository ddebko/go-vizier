package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"time"

	"github.com/SuperBadCode/Vizier/vizier"
)

/*
	This example shows how to write to a lock free slice.
*/

var (
	POOL_SIZE  int = runtime.NumCPU()
	TEST_SIZE  int = 100000 * POOL_SIZE
	ITERATIONS int = POOL_SIZE * 4
)

type BatchWrite struct {
	index    int
	size     int
	poolSize int
}

func main() {
	dest := make([]int, TEST_SIZE)

	manager, err := vizier.NewManager("lockless", POOL_SIZE)
	if err != nil {
		panic(err)
	}

	edgeStart, err := manager.CreateEdge("start")
	if err != nil {
		panic(err)
	}

	edgeWrite, err := manager.CreateEdge("write")
	if err != nil {
		panic(err)
	}

	edgeComplete, err := manager.CreateEdge("complete")
	if err != nil {
		panic(err)
	}

	edgeOutput, err := manager.CreateEdge("output")
	if err != nil {
		panic(err)
	}

	stateComplete := vizier.NewState("complete", func(payload interface{}) interface{} {
		if value, ok := payload.(bool); ok {
			return value
		}
		return vizier.STOP_STATE
	})
	stateComplete.AttachEdge("from_complete_to_output", edgeComplete, edgeOutput)

	stateWrite := vizier.NewState("start", func(payload interface{}) interface{} {
		if batch, ok := payload.(BatchWrite); ok {
			startIndex := batch.index * batch.size
			endIndex := startIndex + batch.size

			if endIndex >= TEST_SIZE {
				stateComplete.Invoke("from_complete_to_output", true)
				return vizier.STOP_STATE
			}

			for i := startIndex; i < endIndex; i++ {
				dest[i] = rand.Intn(100) + 1
			}

			batch.index += batch.poolSize
			return batch
		}
		return vizier.STOP_STATE
	})
	stateWrite.AttachEdge("from_start_to_write", edgeStart, edgeWrite)
	stateWrite.AttachEdge("from_write_to_start", edgeWrite, edgeStart)

	manager.CreateState("lockless_write", stateWrite)
	manager.CreateState("lockless_complete", stateComplete)

	err = manager.Pool.Create()
	if err != nil {
		panic(err)
	}

	start := time.Now()

	for i := 0; i < manager.Pool.GetSize(); i++ {
		stateWrite.Invoke("from_start_to_write", BatchWrite{
			index:    i,
			size:     TEST_SIZE / ITERATIONS,
			poolSize: manager.Pool.GetSize(),
		})
	}

	for i := 0; i < manager.Pool.GetSize(); {
		select {
		case <-edgeOutput:
			i++
		default:
			continue
		}
	}

	fmt.Printf("VIZIER completed %+v\n", time.Since(start))
}
