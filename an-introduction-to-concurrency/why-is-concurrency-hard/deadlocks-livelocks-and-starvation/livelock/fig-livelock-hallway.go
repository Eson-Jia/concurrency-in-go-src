package main

import (
	"bytes"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	cadence := sync.NewCond(&sync.Mutex{})
	go func() {
		for range time.Tick(time.Second * 2) {
			cadence.L.Lock()
			cadence.Broadcast()
			cadence.L.Unlock()
		}
	}()
	takeStep := func() {
		cadence.L.Lock()
		cadence.Wait()
		cadence.L.Unlock()
	}

	tryDir := func(buffer *bytes.Buffer, dir *int32, dirName string) bool {
		atomic.AddInt32(dir, 1)
		fmt.Fprintf(buffer, "%s,", dirName)
		takeStep()
		if atomic.LoadInt32(dir) == 1 {
			fmt.Fprint(buffer, "success")
			return true
		}
		atomic.AddInt32(dir, -1)
		takeStep()
		return false
	}
	var right, left int32
	tryLeft := func(buffer *bytes.Buffer) bool {
		return tryDir(buffer, &left, "left")
	}
	tryRight := func(buffer *bytes.Buffer) bool {
		return tryDir(buffer, &right, "right")
	}
	finished := sync.WaitGroup{}
	walk := func(name string) {
		defer finished.Done()
		out := &bytes.Buffer{}
		defer fmt.Println("test:", out)
		for i := 0; i < 5; i++ {
			if tryLeft(out) || tryRight(out) {
				return
			}
		}
	}
	finished.Add(2)
	go walk("alice")
	go walk("mike")
	finished.Wait()
}
