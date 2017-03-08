package common

import (
	"bytes"
	"fmt"
	"runtime"
	"strconv"
	"time"
	"path/filepath"
)

func getGID() uint64 {
    b := make([]byte, 64)
    b = b[:runtime.Stack(b, false)]
    b = bytes.TrimPrefix(b, []byte("goroutine "))
    b = b[:bytes.IndexByte(b, ' ')]
    n, _ := strconv.ParseUint(string(b), 10, 64)
    return n
}


func Trace() {
	t := time.Now().Format("15:04:05.000000")
	id := getGID()
	pc := make([]uintptr, 10)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	file, line := f.FileLine(pc[0])
	fileName := filepath.Base(file)
	fmt.Printf("%s [TRACE] GID %3d, %s %s:%d\n", t, id, f.Name(), fileName, line)
}
