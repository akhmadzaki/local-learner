package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

// Logger provides structured logging with severity levels.
// In production, logs are output as JSON for ingestion by log aggregators.

var Level int

const (
	LevelDebug   = 0
	LevelInfo    = 1
	LevelWarning = 2
	LevelError   = 3
)

func Init(env string) {
	if env == "production" {
		Level = LevelInfo
	} else {
		Level = LevelDebug
	}
}

func Debug(component string, message string, args ...interface{}) {
	if Level <= LevelDebug {
		logEntry("DEBUG", component, message, args...)
	}
}

func Info(component string, message string, args ...interface{}) {
	if Level <= LevelInfo {
		logEntry("INFO", component, message, args...)
	}
}

func Warning(component string, message string, args ...interface{}) {
	if Level <= LevelWarning {
		logEntry("WARNING", component, message, args...)
	}
}

func Error(component string, message string, args ...interface{}) {
	if Level <= LevelError {
		logEntry("ERROR", component, message, args...)
	}
}

func logEntry(severity string, component string, message string, args ...interface{}) {
	now := time.Now().Format(time.RFC3339)
	prefix := fmt.Sprintf("[%s] [%s] [%s] %s", now, severity, component, message)
	if len(args) > 0 {
		prefix += fmt.Sprintf(" | %v", args)
	}
	log.Println(prefix)

	if severity == "ERROR" {
		fmt.Fprintf(os.Stderr, "%s\n", prefix)
	}
}