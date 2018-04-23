package gslib

import (
	"runtime"
	"time"
)

import (
	. "goslib/logger"
)

func SysRoutine() {
	gc_timer := make(chan int32, 10)
	gc_timer <- 1

	runtime.GC()
	INFO("Goroutine Count:", runtime.NumGoroutine())
	time.AfterFunc(10*time.Second, SysRoutine)
}
