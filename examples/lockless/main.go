package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"time"

	vizier "github.com/SuperBadCode/go-vizier/pkg"
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

	manager.Node("write", func(payload interface{}) map[string]interface{} {
		send_to := map[string]interface{}{}
		if batch, ok := payload.(BatchWrite); ok {
			startIndex := batch.index * batch.size
			endIndex := startIndex + batch.size

			for i := startIndex; i < endIndex; i++ {
				dest[i] = rand.Intn(100) + 1
			}

			batch.index += batch.poolSize

			if (batch.index*batch.size)+batch.size >= TEST_SIZE {
				send_to["batch_complete"] = true
				return send_to
			}

			send_to["write_to_write_next_batch"] = batch
			return send_to
		}
		return send_to
	}).Edge("write", "write", "next_batch")

	output := manager.Output("write", "batch_complete")

	err = manager.Start()
	if err != nil {
		panic(err)
	}

	start := time.Now()
	batch := make([]interface{}, POOL_SIZE)
	for i := 0; i < manager.GetSize(); i++ {
		batch[i] = BatchWrite{
			index:    i,
			size:     TEST_SIZE / ITERATIONS,
			poolSize: POOL_SIZE,
		}
	}

	wg, err := manager.BatchInvoke("write", batch)
	if err != nil {
		panic(err)
	}

	results := manager.GetResults(wg, POOL_SIZE, output)

	fmt.Printf("Complete %+v\n", results)
	fmt.Printf("time %+v\n", time.Since(start))
}
