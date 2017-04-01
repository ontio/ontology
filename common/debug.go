package common

import (
	"bytes"
	"runtime"
	"strconv"
)

func GetGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

/*
func Trace() {
	t := time.Now().Format("15:04:05.000000")
	id := GetGID()
	pc := make([]uintptr, 10)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	file, line := f.FileLine(pc[0])
	fileName := filepath.Base(file)
	fmt.Printf("%s %s GID %d, %s %s:%d\n", t, Color(Pink, "[TRACE]"), id, f.Name(), fileName, line)
}
*/
