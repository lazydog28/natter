package natter

import (
	"fmt"
	"time"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

type Logger struct {
	Level LogLevel
}

var levelMap = map[LogLevel]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
}

func NewLogger(level LogLevel) *Logger {
	return &Logger{Level: level}
}
func (l *Logger) SetLevel(level LogLevel) {
	l.Level = level
}
func GetLogger() *Logger {
	return logger
}

func (l *Logger) logf(level LogLevel, msg string) {
	if level < l.Level {
		return
	}
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	prefix := getPrefix(level)
	println(fmt.Sprintf("[%s] [%s] %s", timestamp, prefix, msg))
}

func getPrefix(level LogLevel) string {
	return levelMap[level]
}

func (l *Logger) Debug(msg string) {
	l.logf(DEBUG, msg)
}

func (l *Logger) Info(msg string) {
	l.logf(INFO, msg)
}

func (l *Logger) Warning(msg string) {
	l.logf(WARN, msg)
}

func (l *Logger) Error(msg string) {
	l.logf(ERROR, msg)
}

var logger = NewLogger(INFO)

// var logger = NewLogger(DEBUG)
