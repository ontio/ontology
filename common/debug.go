package common

import (
	"fmt"
	"runtime"
	"time"
	"path/filepath"
)

func Trace() {
	t := time.Now().Format("15:04:05.000000")
	pc := make([]uintptr, 10)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	file, line := f.FileLine(pc[0])
	fileName := filepath.Base(file)
	fmt.Printf("%s [TRACE] %s@%s:%d\n", t, f.Name(), fileName, line)
}
