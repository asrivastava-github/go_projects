package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

var levelNames = map[Level]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
}

var levelColors = map[Level]string{
	DEBUG: "\033[36m", // Cyan
	INFO:  "\033[32m", // Green
	WARN:  "\033[33m", // Yellow
	ERROR: "\033[31m", // Red
}

const resetColor = "\033[0m"

type Logger struct {
	level      Level
	output     io.Writer
	timeFormat string
	useColor   bool
}

var defaultLogger = &Logger{
	level:      INFO,
	output:     os.Stdout,
	timeFormat: "15:04:05",
	useColor:   true,
}

func SetLevel(level Level) {
	defaultLogger.level = level
}

func SetDebug(debug bool) {
	if debug {
		defaultLogger.level = DEBUG
	} else {
		defaultLogger.level = INFO
	}
}

func GetLevel() Level {
	return defaultLogger.level
}

func IsDebug() bool {
	return defaultLogger.level == DEBUG
}

func (l *Logger) log(level Level, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	msg := fmt.Sprintf(format, args...)
	timestamp := time.Now().Format(l.timeFormat)
	levelName := levelNames[level]

	var line string
	if l.useColor {
		color := levelColors[level]
		line = fmt.Sprintf("%s %s[%s]%s %s\n", timestamp, color, levelName, resetColor, msg)
	} else {
		line = fmt.Sprintf("%s [%s] %s\n", timestamp, levelName, msg)
	}

	fmt.Fprint(l.output, line)
}

func Debug(format string, args ...interface{}) {
	defaultLogger.log(DEBUG, format, args...)
}

func Info(format string, args ...interface{}) {
	defaultLogger.log(INFO, format, args...)
}

func Warn(format string, args ...interface{}) {
	defaultLogger.log(WARN, format, args...)
}

func Error(format string, args ...interface{}) {
	defaultLogger.log(ERROR, format, args...)
}

func ParseLevel(s string) Level {
	switch strings.ToUpper(s) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN", "WARNING":
		return WARN
	case "ERROR":
		return ERROR
	default:
		return INFO
	}
}
