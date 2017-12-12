package log

import (
	"github.com/Ontology/common/config"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
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
	Stdout = os.Stdout
)

const (
	namePrefix        = "LEVEL"
	callDepth         = 2
	defaultMaxLogSize = 20
	byteToMb          = 1024 * 1024
	byteToKb          = 1024
	Path              = "./Log/"
)

func GetGID() uint64 {
	var buf [64]byte
	b := buf[:runtime.Stack(buf[:], false)]
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
	level   int
	logger  *log.Logger
	logFile *os.File
}

func New(out io.Writer, prefix string, flag, level int, file *os.File) *Logger {
	return &Logger{
		level:   level,
		logger:  log.New(out, prefix, flag),
		logFile: file,
	}
}

func (l *Logger) SetDebugLevel(level int) error {
	if level > maxLevelLog || level < 0 {
		return errors.New("Invalid Debug Level")
	}

	l.level = level
	return nil
}

func (l *Logger) Output(level int, a ...interface{}) error {
	if level >= l.level {
		gid := GetGID()
		gidStr := strconv.FormatUint(gid, 10)

		a = append([]interface{}{LevelName(level), "GID",
			gidStr + ","}, a...)

		return l.logger.Output(callDepth, fmt.Sprintln(a...))
	}
	return nil
}

func (l *Logger) Outputf(level int, format string, v ...interface{}) error {
	if level >= l.level {
		gid := GetGID()
		v = append([]interface{}{LevelName(level), "GID",
			gid}, v...)

		return l.logger.Output(callDepth, fmt.Sprintf("%s %s %d, "+format+"\n", v...))
	}
	return nil
}

func (l *Logger) Trace(a ...interface{}) {
	l.Output(traceLog, a...)
}

func (l *Logger) Tracef(format string, a ...interface{}) {
	l.Outputf(traceLog, format, a...)
}

func (l *Logger) Debug(a ...interface{}) {
	l.Output(debugLog, a...)
}

func (l *Logger) Debugf(format string, a ...interface{}) {
	l.Outputf(debugLog, format, a...)
}

func (l *Logger) Info(a ...interface{}) {
	l.Output(infoLog, a...)
}

func (l *Logger) Infof(format string, a ...interface{}) {
	l.Outputf(infoLog, format, a...)
}

func (l *Logger) Warn(a ...interface{}) {
	l.Output(warnLog, a...)
}

func (l *Logger) Warnf(format string, a ...interface{}) {
	l.Outputf(warnLog, format, a...)
}

func (l *Logger) Error(a ...interface{}) {
	l.Output(errorLog, a...)
}

func (l *Logger) Errorf(format string, a ...interface{}) {
	l.Outputf(errorLog, format, a...)
}

func (l *Logger) Fatal(a ...interface{}) {
	l.Output(fatalLog, a...)
}

func (l *Logger) Fatalf(format string, a ...interface{}) {
	l.Outputf(fatalLog, format, a...)
}

func Trace(a ...interface{}) {
	if traceLog < Log.level {
		return
	}

	pc := make([]uintptr, 10)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	file, line := f.FileLine(pc[0])
	fileName := filepath.Base(file)

	nameFull := f.Name()
	nameEnd := filepath.Ext(nameFull)
	funcName := strings.TrimPrefix(nameEnd, ".")

	a = append([]interface{}{funcName + "()", fileName + ":" + strconv.Itoa(line)}, a...)

	Log.Trace(a...)
}

func Tracef(format string, a ...interface{}) {
	if traceLog < Log.level {
		return
	}

	pc := make([]uintptr, 10)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	file, line := f.FileLine(pc[0])
	fileName := filepath.Base(file)

	nameFull := f.Name()
	nameEnd := filepath.Ext(nameFull)
	funcName := strings.TrimPrefix(nameEnd, ".")

	a = append([]interface{}{funcName, fileName, line}, a...)

	Log.Tracef("%s() %s:%d "+format, a...)
}

func Debug(a ...interface{}) {
	if debugLog < Log.level {
		return
	}

	pc := make([]uintptr, 10)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	file, line := f.FileLine(pc[0])
	fileName := filepath.Base(file)

	a = append([]interface{}{f.Name(), fileName + ":" + strconv.Itoa(line)}, a...)

	Log.Debug(a...)
}

func Debugf(format string, a ...interface{}) {
	if debugLog < Log.level {
		return
	}

	pc := make([]uintptr, 10)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	file, line := f.FileLine(pc[0])
	fileName := filepath.Base(file)

	a = append([]interface{}{f.Name(), fileName, line}, a...)

	Log.Debugf("%s %s:%d "+format, a...)
}

func Info(a ...interface{}) {
	Log.Info(a...)
}

func Warn(a ...interface{}) {
	Log.Warn(a...)
}

func Error(a ...interface{}) {
	Log.Error(a...)
}

func Fatal(a ...interface{}) {
	Log.Fatal(a...)
}

func Infof(format string, a ...interface{}) {
	Log.Infof(format, a...)
}

func Warnf(format string, a ...interface{}) {
	Log.Warnf(format, a...)
}

func Errorf(format string, a ...interface{}) {
	Log.Errorf(format, a...)
}

func Fatalf(format string, a ...interface{}) {
	Log.Fatalf(format, a...)
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

	var currenttime string = time.Now().Format("2006-01-02_15.04.05")

	logfile, err := os.OpenFile(path+currenttime+"_LOG.log", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	return logfile, nil
}

func Init(a ...interface{}) {
	writers := []io.Writer{}
	var logFile *os.File
	var err error
	if len(a) == 0 {
		writers = append(writers, ioutil.Discard)
	} else {
		for _, o := range a {
			switch o.(type) {
			case string:
				logFile, err = FileOpen(o.(string))
				if err != nil {
					fmt.Println("error: open log file failed")
					os.Exit(1)
				}
				writers = append(writers, logFile)
			case *os.File:
				writers = append(writers, o.(*os.File))
			default:
				fmt.Println("error: invalid log location")
				os.Exit(1)
			}
		}
	}
	fileAndStdoutWrite := io.MultiWriter(writers...)
	var printlevel int = config.Parameters.PrintLevel
	Log = New(fileAndStdoutWrite, "", log.Ldate|log.Lmicroseconds, printlevel, logFile)
}

func GetLogFileSize() (int64, error) {
	f, e := Log.logFile.Stat()
	if e != nil {
		return 0, e
	}
	return f.Size(), nil
}

func GetMaxLogChangeInterval() int64 {
	if config.Parameters.MaxLogSize != 0 {
		return (config.Parameters.MaxLogSize * byteToMb)
	} else {
		return (defaultMaxLogSize * byteToMb)
	}
}

func CheckIfNeedNewFile() bool {
	logFileSize, err := GetLogFileSize()
	maxLogFileSize := GetMaxLogChangeInterval()
	if err != nil {
		return false
	}
	if logFileSize > maxLogFileSize {
		return true
	} else {
		return false
	}
}

func ClosePrintLog() error {
	var err error
	if Log.logFile != nil {
		err = Log.logFile.Close()
	}
	return err
}
