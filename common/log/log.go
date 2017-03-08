package log

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	debugLog = iota
	infoLog
	warningLog
	errorLog
	fatalLog
	numSeverity = 5
)

var (
	levels = map[int]string{
		debugLog:   "DEBUG",
		infoLog:    "INFO",
		warningLog: "WARNING",
		errorLog:   "ERROR",
		fatalLog:   "FATAL",
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
	return l.logger.Output(callDepth, AddBracket(LevelName(level))+" "+s)
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
	l.Output(warningLog, a...)
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
	var printlevel int = 1
	writers := []io.Writer{
		logfile,
		os.Stdout,
	}
	fileAndStdoutWrite := io.MultiWriter(writers...)

	Log = New(fileAndStdoutWrite, "\r\n", log.Lmicroseconds, printlevel)
}

func ClosePrintLog() {
	//TODO
}
