package common

import (
	"fmt"
	"runtime"
	"time"
)

func Trace() {
	tx := time.Now()
	pc := make([]uintptr, 10)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	file, line := f.FileLine(pc[0])
	fmt.Printf("[%s] %s:%d %s\n", tx,file, line, f.Name())
}
