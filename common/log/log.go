package log

import (
	"DNA/common/config"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	Blue   = "0;34"
	Red    = "0;31"
	Green  = "0;32"
	Yellow = "0;33"
	Cyan   = "0;36"
	Pink   = "1;35"
)

func Color(code, msg string) string {
	return fmt.Sprintf("\033[%sm%s\033[m", code, msg)
}

const (
	debugLog = iota
	infoLog
	warnLog
	errorLog
	fatalLog
	traceLog
	maxLevelLog
)

var (
	levels = map[int]string{
		debugLog: Color(Green, "[DEBUG]"),
		infoLog:  Color(Green, "[INFO ]"),
		warnLog:  Color(Yellow, "[WARN ]"),
		errorLog: Color(Red, "[ERROR]"),
		fatalLog: Color(Red, "[FATAL]"),
		traceLog: Color(Pink, "[TRACE]"),
	}
)

const (
	namePrefix = "LEVEL"
	callDepth  = 2
)

func GetGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

var Log *Logger

func LevelName(level int) string {
	if name, ok := levels[level]; ok {
		return name
	}
	return namePrefix + strconv.Itoa(level)
}

func NameLevel(name string) int {
	for k, v := range levels {
		if v == name {
			return k
		}
	}
	var level int
	if strings.HasPrefix(name, namePrefix) {
		level, _ = strconv.Atoi(name[len(namePrefix):])
	}
	return level
}

type Logger struct {
	sync.Mutex
	level  int
	logger *log.Logger
}

func New(out io.Writer, prefix string, flag, level int) *Logger {
	return &Logger{
		level:  level,
		logger: log.New(out, prefix, flag),
	}
}

func (l *Logger) SetDebugLevel(level int) error {
	l.Lock()
	defer l.Unlock()
	if level > maxLevelLog || level < 0 {
		return errors.New("Invalid Debug Level")
	}
	l.level = level
	return nil
}

func (l *Logger) output(level int, s string) error {
	gid := GetGID()
	gidStr := strconv.FormatUint(gid, 10)

	return l.logger.Output(callDepth, LevelName(level)+" "+"GID"+
		" "+gidStr+", "+s)
}

func (l *Logger) Output(level int, a ...interface{}) error {
	if level >= l.level {
		return l.output(level, fmt.Sprintln(a...))
	}
	return nil
}

func (l *Logger) Trace(a ...interface{}) {
	l.Lock()
	defer l.Unlock()
	l.Output(traceLog, a...)
}

func (l *Logger) Debug(a ...interface{}) {
	l.Lock()
	defer l.Unlock()
	l.Output(debugLog, a...)
}

func (l *Logger) Info(a ...interface{}) {
	l.Lock()
	defer l.Unlock()
	l.Output(infoLog, a...)
}

func (l *Logger) Warn(a ...interface{}) {
	l.Lock()
	defer l.Unlock()
	l.Output(warnLog, a...)
}

func (l *Logger) Error(a ...interface{}) {
	l.Lock()
	defer l.Unlock()
	l.Output(errorLog, a...)
}

func (l *Logger) Fatal(a ...interface{}) {
	l.Lock()
	defer l.Unlock()
	l.Output(fatalLog, a...)
}

func Trace(a ...interface{}) {
	pc := make([]uintptr, 10)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	file, line := f.FileLine(pc[0])
	fileName := filepath.Base(file)

	nameFull := f.Name()
	nameEnd := filepath.Ext(nameFull)
	funcName := strings.TrimPrefix(nameEnd, ".")
	Log.Trace(fmt.Sprint(funcName, "() ", fileName, ":", line, " ", fmt.Sprint(a...)))
}

func Debug(a ...interface{}) {
	pc := make([]uintptr, 10)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	file, line := f.FileLine(pc[0])
	fileName := filepath.Base(file)
	Log.Debug(fmt.Sprint(f.Name(), " ", fileName, ":", line, " ", fmt.Sprint(a...)))
}

func Info(a ...interface{}) {
	Log.Info(fmt.Sprint(a...))
}

func Warn(a ...interface{}) {
	Log.Warn(fmt.Sprint(a...))
}

func Error(a ...interface{}) {
	Log.Error(fmt.Sprint(a...))
}

func Fatal(a ...interface{}) {
	Log.Fatal(fmt.Sprint(a...))
}

func FileOpen(path string) (*os.File, error) {
	if fi, err := os.Stat(path); err == nil {
		if !fi.IsDir() {
			return nil, fmt.Errorf("open %s: not a directory", path)
		}
	} else if os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0766); err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	var currenttime string = time.Now().Format("2006-01-02")

	logfile, err := os.OpenFile(path+currenttime+"_LOG.log", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		//os.Exit(-1)
	}

	//defer logfile.Close()

	return logfile, nil
}

func CreatePrintLog(path string) {
	logfile, err := FileOpen(path)
	if err != nil {
		fmt.Printf("%s\n", err.Error)
	}

	var printlevel int = config.Parameters.PrintLevel
	writers := []io.Writer{
		logfile,
		os.Stdout,
	}
	fileAndStdoutWrite := io.MultiWriter(writers...)

	Log = New(fileAndStdoutWrite, "", log.Lmicroseconds, printlevel)
}

func ClosePrintLog() {
	//TODO
}
