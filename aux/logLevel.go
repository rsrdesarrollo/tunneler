package aux

import log "github.com/alecthomas/log4go"

func LogLevel(level string) log.Level {
	switch level {
	case "FINEST":
		return log.FINEST
	case "FINE":
		return log.FINE
	case "DEBUG":
		return log.DEBUG
	case "TRACE":
		return log.TRACE
	case "INFO":
		return log.INFO
	case "WARNING":
		return log.WARNING
	case "ERROR":
		return log.ERROR
	case "CRITICAL":
		return log.CRITICAL
	default:
		return log.INFO
	}
}
