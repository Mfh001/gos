package gslib

import (
	"runtime"
	"time"
)

import (
	. "goslib/logger"
)

func SysRoutine() {
	gcTimer := make(chan int32, 10)
	gcTimer <- 1

	INFO("Goroutine Count:", runtime.NumGoroutine())
	time.AfterFunc(10*time.Second, SysRoutine)
}
