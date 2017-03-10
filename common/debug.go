package common

import (
	"bytes"
	"fmt"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

func GetGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

func Trace() {
	t := time.Now().Format("15:04:05.000000")
	id := GetGID()
	pc := make([]uintptr, 10)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	file, line := f.FileLine(pc[0])
	fileName := filepath.Base(file)
	fmt.Printf("%s [TRACE] GID %d, %s %s:%d\n", t, id, f.Name(), fileName, line)
}
