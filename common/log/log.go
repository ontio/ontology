package log

import (
	"GoOnchain/common"
	"GoOnchain/config"
	"bytes"
	"fmt"
	"path/filepath"
	"io"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	PRINTLEVEL = 0
)

const (
	debugLog = iota
	infoLog
	warnLog
	errorLog
	fatalLog
	numSeverity = 5
)

var (
	levels = map[int]string{
		debugLog: "DEBUG",
		infoLog:  "INFO ",
		warnLog:  "WARN ",
		errorLog: "ERROR",
		fatalLog: "FATAL",
	}
)

const (
	namePrefix = "LEVEL"
	callDepth  = 2
)

var Log *Logger
var lock = sync.Mutex{}

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

func AddBracket(s string) string {
	b := bytes.Buffer{}
	b.WriteString("[")
	b.WriteString(s)
	b.WriteString("]")
	return b.String()
}

type Logger struct {
	level  int
	logger *log.Logger
}

func New(out io.Writer, prefix string, flag, level int) *Logger {
	return &Logger{
		level:  level,
		logger: log.New(out, prefix, flag),
	}
}

func (l *Logger) output(level int, s string) error {
	// FIXME enable print GID for all log, should be disable as it effect performance
	if (level == 0) || (level == 1) || (level == 2) || (level == 3) {
		gid := common.GetGID()
		gidStr := strconv.FormatUint(gid, 10)

		// Get file information only
		pc := make([]uintptr, 10)
		runtime.Callers(2, pc)
		f := runtime.FuncForPC(pc[0])
		file, line := f.FileLine(pc[0])
		fileName := filepath.Base(file)
		lineStr := strconv.FormatUint(uint64(line), 10)
		return l.logger.Output(callDepth, AddBracket(LevelName(level))+" "+"GID"+
			" "+gidStr+", "+s+" "+fileName+":"+lineStr)
	} else {
		return l.logger.Output(callDepth, AddBracket(LevelName(level))+" "+s)
	}
}

func (l *Logger) Output(level int, a ...interface{}) error {
	if level >= l.level {
		return l.output(level, fmt.Sprintln(a...))
	}
	return nil
}

func (l *Logger) Debug(a ...interface{}) {
	lock.Lock()
	defer lock.Unlock()
	l.Output(debugLog, a...)
}

func (l *Logger) Info(a ...interface{}) {
	lock.Lock()
	defer lock.Unlock()
	l.Output(infoLog, a...)
}

func (l *Logger) Warn(a ...interface{}) {
	lock.Lock()
	defer lock.Unlock()
	l.Output(warnLog, a...)
}

func (l *Logger) Error(a ...interface{}) {
	lock.Lock()
	defer lock.Unlock()
	l.Output(errorLog, a...)
}

func (l *Logger) Fatal(a ...interface{}) {
	lock.Lock()
	defer lock.Unlock()
	l.Output(fatalLog, a...)
}

func Debug(a ...interface{}) {
	Log.Debug(fmt.Sprint(a...))
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
