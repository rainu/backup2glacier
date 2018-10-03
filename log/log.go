package log

import (
	"fmt"
	logger "log"
)

func LogFatal(format string, args ...interface{}) {
	logger.Fatal(formatMessage("FATAL", fmt.Sprintf(format, args...)))
}

func LogError(format string, args ...interface{}) {
	log("ERROR", fmt.Sprintf(format, args...))
}

func LogInfo(format string, args ...interface{}) {
	log("INFO", fmt.Sprintf(format, args...))
}

func LogDebug(format string, args ...interface{}) {
	log("DEBUG", fmt.Sprintf(format, args...))
}

func formatMessage(level, message string) string {
	return fmt.Sprintf("[%s] %s\n", level, message)
}

func log(level, message string) {
	logger.Print(formatMessage(level, message))
}
