package logging

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

func ParseLevel(lvl string) Level {
	switch lvl {
	case "debug":
		return DEBUG
	case "warn":
		return WARN
	case "error":
		return ERROR
	default:
		return INFO
	}
}

func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "INFO"
	}
}

type Logger struct {
	mu     sync.Mutex
	writer io.Writer
	level  Level
}

var (
	defaultLogger *Logger
	once          sync.Once
)

// Init initializes the global logger writing to a specific file.
func Init(filePath string, levelStr string) error {
	var writer io.Writer
	if filePath == "" {
		writer = io.Discard
	} else {
		file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("failed to open log file %s: %w", filePath, err)
		}
		writer = file
	}

	once.Do(func() {
		defaultLogger = &Logger{
			writer: writer,
			level:  ParseLevel(levelStr),
		}
	})
	return nil
}

func Get() *Logger {
	if defaultLogger == nil {
		defaultLogger = &Logger{
			writer: io.Discard,
			level:  INFO,
		}
	}
	return defaultLogger
}

func (l *Logger) log(lvl Level, format string, args ...interface{}) {
	if lvl < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(l.writer, "[%s] [%s] %s\n", timestamp, lvl.String(), msg)
}

func Debug(format string, args ...interface{}) { Get().log(DEBUG, format, args...) }
func Info(format string, args ...interface{})  { Get().log(INFO, format, args...) }
func Warn(format string, args ...interface{})  { Get().log(WARN, format, args...) }
func Error(format string, args ...interface{}) { Get().log(ERROR, format, args...) }
