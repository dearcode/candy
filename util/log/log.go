package log

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"time"
)

type logger struct {
	out   *os.File
	level int
	color bool
}

const (
	LOG_FATAL = iota
	LOG_ERROR
	LOG_WARNING
	LOG_INFO
	LOG_DEBUG
)

var mlog *logger

func init() {
	mlog = &logger{out: os.Stdout, level: LOG_DEBUG, color: true}
}

func SetLevel(level int) {
	mlog.SetLevel(level)
}

func GetLogLevel() int {
	return mlog.level
}

func Info(v ...interface{}) {
	mlog.Log(LOG_INFO, fmt.Sprint(v...))
}

func Infof(format string, v ...interface{}) {
	mlog.Log(LOG_INFO, fmt.Sprintf(format, v...))
}

func Debug(v ...interface{}) {
	mlog.Log(LOG_DEBUG, fmt.Sprint(v...))
}

func Debugf(format string, v ...interface{}) {
	mlog.Log(LOG_DEBUG, fmt.Sprintf(format, v...))
}

func Warning(v ...interface{}) {
	mlog.Log(LOG_WARNING, fmt.Sprint(v...))
}

func Warningf(format string, v ...interface{}) {
	mlog.Log(LOG_WARNING, fmt.Sprintf(format, v...))
}

func Error(v ...interface{}) {
	mlog.Log(LOG_ERROR, fmt.Sprint(v...))
}

func Errorf(format string, v ...interface{}) {
	mlog.Log(LOG_ERROR, fmt.Sprintf(format, v...))
}

func Fatal(v ...interface{}) {
	mlog.Log(LOG_FATAL, fmt.Sprint(v...))
	os.Exit(-1)
}

func Fatalf(format string, v ...interface{}) {
	mlog.Log(LOG_FATAL, fmt.Sprintf(format, v...))
	os.Exit(-1)
}

func SetLevelByString(level string) {
	mlog.SetLevelByString(level)
}

func SetColor(color bool) {
	mlog.SetColor(color)
}

func (l *logger) SetColor(color bool) {
	l.color = color
}

func (l *logger) SetLevel(level int) {
	l.level = level
}

func (l *logger) SetLevelByString(level string) {
	l.level = StringToLogLevel(level)
}

func (l *logger) caller() (string, string) {
	pc, file, line, _ := runtime.Caller(3)
	name := runtime.FuncForPC(pc).Name()
	if i := bytes.LastIndexAny([]byte(name), "."); i != -1 {
		name = name[i+1:]
	}

	if i := bytes.LastIndexAny([]byte(file), "/"); i != -1 {
		file = file[i+1:]
	}

	date := time.Now().Format("2006/01/02 15:04:05")

	return fmt.Sprintf("%s %s:%d", date, file, line), name
}

func (l *logger) Log(t int, info string) {
	if t > l.level {
		return
	}

	header, name := l.caller()

	logStr, logColor := LogTypeToString(t)

	if l.color {
		fmt.Fprintln(l.out, header, fmt.Sprint("\033", logColor, "m[", logStr, "] ", name, " ", info, "\033[0m"))
	} else {
		fmt.Fprintln(l.out, header, "[", logStr, "]", name, " ", info)
	}
}

func StringToLogLevel(level string) int {
	switch level {
	case "fatal":
		return LOG_FATAL
	case "error":
		return LOG_ERROR
	case "warn":
		return LOG_WARNING
	case "warning":
		return LOG_WARNING
	case "debug":
		return LOG_DEBUG
	case "info":
		return LOG_INFO
	}
	return LOG_DEBUG
}

func LogTypeToString(t int) (string, string) {
	switch t {
	case LOG_FATAL:
		return "fatal", "[0;31"
	case LOG_ERROR:
		return "error", "[0;31"
	case LOG_WARNING:
		return "warning", "[0;33"
	case LOG_DEBUG:
		return "debug", "[0;36"
	case LOG_INFO:
		return "info", "[0;37"
	}
	return "unknown", "[0;37"
}
