package internal

import (
	"fmt"
)

const (
	// ANSI escape code for red color
	red        = "\033[31m"
	green      = "\033[32m"
	yellow     = "\033[33m"
	resetColor = "\033[0m"
)

func DebugPrintf(format string, a ...any) {
	consolePrint("Debug", yellow, fmt.Sprintf(format, a...))
}

func InfoPrintf(format string, a ...any) {
	consolePrint("Info", green, fmt.Sprintf(format, a...))
}

func ErrorPrintf(format string, a ...any) {
	consolePrint("Error", red, fmt.Sprintf(format, a...))
}

func DebugPrint(a ...any) {
	consolePrint("Debug", yellow, a...)
}

func InfoPrint(a ...any) {
	consolePrint("Info", green, a...)
}

func ErrorPrint(a ...any) {
	consolePrint("Error", red, a...)
}

func consolePrint(level, color string, a ...any) {
	var args []any
	args = append(args, color+level+resetColor)
	args = append(args, a...)
	fmt.Println(args...)
}
