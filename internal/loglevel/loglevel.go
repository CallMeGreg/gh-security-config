// Package loglevel manages the global log verbosity for the extension.
// It is intentionally free of internal dependencies so that any package
// (including api and utils) can check the current level without import cycles.
package loglevel

import (
	"fmt"
	"strings"
	"sync"
)

// LogLevel represents the verbosity of output emitted by the extension.
type LogLevel int

const (
	// LogLevelInfo emits informational messages in addition to warnings and errors.
	LogLevelInfo LogLevel = iota
	// LogLevelWarning (the default) emits warnings and errors but suppresses info messages.
	LogLevelWarning
	// LogLevelError emits only errors.
	LogLevelError
)

// LogLevelDefault is the default log level used when the user does not set one.
const LogLevelDefault = "warning"

// LogLevelValues lists the accepted values for the --log-level flag.
var LogLevelValues = []string{"info", "warning", "error"}

var (
	logLevelMu sync.RWMutex
	logLevel   = LogLevelWarning
)

// ParseLogLevel converts a user-supplied string to a LogLevel. The comparison is
// case-insensitive and whitespace is trimmed. An empty string resolves to the
// default level.
func ParseLogLevel(value string) (LogLevel, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		normalized = LogLevelDefault
	}
	switch normalized {
	case "info":
		return LogLevelInfo, nil
	case "warning":
		return LogLevelWarning, nil
	case "error":
		return LogLevelError, nil
	default:
		return LogLevelWarning, fmt.Errorf("invalid value for log-level flag: %q (must be one of: %s)", value, strings.Join(LogLevelValues, ", "))
	}
}

// SetLogLevel updates the package-level log level. Safe for concurrent use.
func SetLogLevel(level LogLevel) {
	logLevelMu.Lock()
	defer logLevelMu.Unlock()
	logLevel = level
}

// GetLogLevel returns the current log level. Safe for concurrent use.
func GetLogLevel() LogLevel {
	logLevelMu.RLock()
	defer logLevelMu.RUnlock()
	return logLevel
}

// WarningEnabled reports whether warning messages should be emitted.
func WarningEnabled() bool {
	return GetLogLevel() <= LogLevelWarning
}

// InfoEnabled reports whether informational messages should be emitted.
func InfoEnabled() bool {
	return GetLogLevel() <= LogLevelInfo
}
