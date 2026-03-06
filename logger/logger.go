package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

const (
	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelError = "error"
	LevelFatal = "fatal"
)

type LogEntry struct {
	Level      string                 `json:"level"`
	Timestamp  string                 `json:"timestamp"`
	Message    string                 `json:"message"`
	Service    string                 `json:"service,omitempty"`
	Fields     map[string]interface{} `json:"fields,omitempty"`
}

type Logger struct {
	service string
	fields  map[string]interface{}
}

var (
	defaultLogger *Logger
	logLevel      string
)

func init() {
	logLevel = os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = LevelInfo
	}
	defaultLogger = NewLogger("sfsEdgeStore")
}

func NewLogger(service string) *Logger {
	return &Logger{
		service: service,
		fields:  make(map[string]interface{}),
	}
}

func SetLogLevel(level string) {
	logLevel = level
}

func (l *Logger) WithField(key string, value interface{}) *Logger {
	newLogger := NewLogger(l.service)
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}
	newLogger.fields[key] = value
	return newLogger
}

func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	newLogger := NewLogger(l.service)
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}
	for k, v := range fields {
		newLogger.fields[k] = v
	}
	return newLogger
}

func (l *Logger) log(level string, msg string, fields ...map[string]interface{}) {
	if !shouldLog(level) {
		return
	}

	entry := LogEntry{
		Level:     level,
		Timestamp: time.Now().Format(time.RFC3339Nano),
		Message:   msg,
		Service:   l.service,
		Fields:    make(map[string]interface{}),
	}

	for k, v := range l.fields {
		entry.Fields[k] = v
	}

	for _, f := range fields {
		for k, v := range f {
			entry.Fields[k] = v
		}
	}

	data, err := json.Marshal(entry)
	if err != nil {
		log.Printf("Failed to marshal log entry: %v", err)
		return
	}

	fmt.Println(string(data))
}

func shouldLog(level string) bool {
	levels := []string{LevelDebug, LevelInfo, LevelWarn, LevelError, LevelFatal}
	currentIndex := 0
	levelIndex := 0

	for i, l := range levels {
		if l == logLevel {
			currentIndex = i
		}
		if l == level {
			levelIndex = i
		}
	}

	return levelIndex >= currentIndex
}

func (l *Logger) Debug(msg string, fields ...map[string]interface{}) {
	l.log(LevelDebug, msg, fields...)
}

func (l *Logger) Info(msg string, fields ...map[string]interface{}) {
	l.log(LevelInfo, msg, fields...)
}

func (l *Logger) Warn(msg string, fields ...map[string]interface{}) {
	l.log(LevelWarn, msg, fields...)
}

func (l *Logger) Error(msg string, fields ...map[string]interface{}) {
	l.log(LevelError, msg, fields...)
}

func (l *Logger) Fatal(msg string, fields ...map[string]interface{}) {
	l.log(LevelFatal, msg, fields...)
	os.Exit(1)
}

func Debug(msg string, fields ...map[string]interface{}) {
	defaultLogger.Debug(msg, fields...)
}

func Info(msg string, fields ...map[string]interface{}) {
	defaultLogger.Info(msg, fields...)
}

func Warn(msg string, fields ...map[string]interface{}) {
	defaultLogger.Warn(msg, fields...)
}

func Error(msg string, fields ...map[string]interface{}) {
	defaultLogger.Error(msg, fields...)
}

func Fatal(msg string, fields ...map[string]interface{}) {
	defaultLogger.Fatal(msg, fields...)
}
