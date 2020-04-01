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
		for range time.Tick(time.Millisecond * 1) {
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
		if atomic.LoadInt32(dir) == 1 { // <1.1>
			fmt.Fprint(buffer, "success")
			return true
		}
		takeStep()               // <1.2>
		atomic.AddInt32(dir, -1) // <1.3>
		// 如果 1.2 和 1.3 位置互换的话，死锁不容易出现
		// 这是因为在两个 goroutine 在 takeStep 之后同步开始往下执行
		// 但是如果一个 goroutine 很快执行完 dir -1之后,
		//另一个 goroutine 才执行到 if 语句,那么条件为 true 就会 return true
		// 进而退出活锁状态
		// 不过这是不正确的,因为为了制造活锁,明确要求了两个问必须同时移动
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
		out := bytes.NewBufferString(fmt.Sprintf("%v is trying to scoot:", name))
		defer fmt.Println(out)
		for i := 0; i < 5; i++ {
			if tryLeft(out) || tryRight(out) {
				return
			}
		}
		fmt.Fprintf(out, "\n%v tosses her hands up in exasperation!", name)
	}
	start := time.Now()
	finished.Add(2)
	go walk("alice")
	go walk("mike")
	finished.Wait()
	fmt.Println("duration:", time.Since(start))
}
