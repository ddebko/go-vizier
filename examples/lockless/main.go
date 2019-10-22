package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync/atomic"
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

	var ops int32
	stateWrite := vizier.NewState("start", func(payload interface{}) interface{} {
		if batch, ok := payload.(BatchWrite); ok {
			atomic.AddInt32(&ops, 1)

			startIndex := batch.index * batch.size
			endIndex := startIndex + batch.size

			for i := startIndex; i < endIndex; i++ {
				dest[i] = rand.Intn(100) + 1
			}

			batch.index += batch.poolSize

			if (batch.index*batch.size)+batch.size >= TEST_SIZE {
				return vizier.STOP_STATE
			}
			return batch
		}
		return vizier.STOP_STATE
	})
	stateWrite.AttachEdge("from_start_to_write", edgeStart, edgeWrite)
	stateWrite.AttachEdge("from_write_to_start", edgeWrite, edgeStart)

	manager.CreateState("lockless_write", stateWrite)

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

	for {
		count := atomic.LoadInt32(&ops)
		if count >= 47 {
			break
		}
		fmt.Println(count)
	}

	fmt.Printf("VIZIER completed %+v\n", time.Since(start))
}
