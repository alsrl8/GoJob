package xlog

import (
	"GoJob/config"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

type logger interface {
	Info(msg ...interface{})
	Error(msg ...interface{})
	Close()
}

type XLogger struct {
	file *os.File
}

var Logger *XLogger
var once sync.Once

func getLogPath() string {
	runEnv := config.GetRunEnv()
	switch runEnv {
	case config.Local:
		return "./app.log"
	case config.Test:
		return "../test.log"
	default:
		return ""
	}
}

func NewXLogger() *XLogger {
	once.Do(func() {
		logPath := getLogPath()
		logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			fmt.Printf("failed to open log file: %v", err)
			return
		}

		Logger = &XLogger{
			file: logFile,
		}
	})

	return Logger
}

func (x *XLogger) Info(msg ...interface{}) {
	message := makeMsgFormat(msg...)
	x.logToFile("[INFO] " + message)
}

func (x *XLogger) Error(msg ...interface{}) {
	message := makeMsgFormat(msg...)
	trace := getTrace()
	x.logToFile(fmt.Sprintf("[ERROR] %s\n%s", message, trace))
}

func (x *XLogger) logToFile(msg string) {
	timestamp := time.Now().Format("2006/01/02 15:04:05")
	if _, err := x.file.WriteString(fmt.Sprintf("%s %s\n", timestamp, msg)); err != nil {
		fmt.Printf("failed to write to log file: %v", err)
	}
}

func (x *XLogger) Close() {
	if err := x.file.Close(); err != nil {
		fmt.Printf("failed to close log file: %v", err)
	}
}

func makeMsgFormat(msg ...interface{}) string {
	messages := make([]string, len(msg))
	for i, m := range msg {
		messages[i] = fmt.Sprint(m)
	}

	return strings.Join(messages, ", ")
}

func getTrace() string {
	stack := make([]uintptr, 10)
	n := runtime.Callers(2, stack)
	frames := runtime.CallersFrames(stack[:n])

	var trace strings.Builder
	for {
		frame, more := frames.Next()
		trace.WriteString(fmt.Sprintf("\t%s:%d\n", frame.File, frame.Line))
		if !more {
			break
		}
	}
	return trace.String()
}
