package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

var (
	currentLevel = LevelInfo
	logger       = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
)

func Init(level string) {
	switch strings.ToLower(level) {
	case "debug":
		currentLevel = LevelDebug
	case "info":
		currentLevel = LevelInfo
	case "warn", "warning":
		currentLevel = LevelWarn
	case "error":
		currentLevel = LevelError
	case "fatal":
		currentLevel = LevelFatal
	}
}

func Debug(format string, v ...interface{}) {
	if currentLevel <= LevelDebug {
		logger.Output(2, fmt.Sprintf("[DEBUG] "+format, v...))
	}
}

func Info(format string, v ...interface{}) {
	if currentLevel <= LevelInfo {
		logger.Output(2, fmt.Sprintf("[INFO] "+format, v...))
	}
}

func Warn(format string, v ...interface{}) {
	if currentLevel <= LevelWarn {
		logger.Output(2, fmt.Sprintf("[WARN] "+format, v...))
	}
}

func Error(format string, v ...interface{}) {
	if currentLevel <= LevelError {
		logger.Output(2, fmt.Sprintf("[ERROR] "+format, v...))
	}
}

func Fatal(format string, v ...interface{}) {
	logger.Output(2, fmt.Sprintf("[FATAL] "+format, v...))
	os.Exit(1)
}
