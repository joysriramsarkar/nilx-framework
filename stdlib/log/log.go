// Package log provides leveled logging utilities for NilLang.
package log

import (
	"fmt"
	"time"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

var CurrentLevel = LevelInfo

func Debug(msg string) {
	if CurrentLevel <= LevelDebug {
		printLog("DEBUG", msg)
	}
}

func Info(msg string) {
	if CurrentLevel <= LevelInfo {
		printLog("INFO", msg)
	}
}

func Warn(msg string) {
	if CurrentLevel <= LevelWarn {
		printLog("WARN", msg)
	}
}

func Error(msg string) {
	if CurrentLevel <= LevelError {
		printLog("ERROR", msg)
	}
}

func printLog(level, msg string) {
	t := time.Now().Format("15:04:05.000")
	fmt.Printf("[%s] [%s] %s\n", t, level, msg)
}
