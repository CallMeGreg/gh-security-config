package ui

import (
	"fmt"
	"strings"

	"github.com/pterm/pterm"

	"github.com/callmegreg/gh-security-config/internal/loglevel"
)

// LogLevel is an alias for loglevel.LogLevel so existing callers compile unchanged.
type LogLevel = loglevel.LogLevel

// Re-export log-level constants.
const (
	LogLevelInfo    = loglevel.LogLevelInfo
	LogLevelWarning = loglevel.LogLevelWarning
	LogLevelError   = loglevel.LogLevelError
)

// LogLevelDefault is the default log level used when the user does not set one.
const LogLevelDefault = loglevel.LogLevelDefault

// LogLevelValues lists the accepted values for the --log-level flag.
var LogLevelValues = loglevel.LogLevelValues

// ParseLogLevel delegates to loglevel.ParseLogLevel.
func ParseLogLevel(value string) (LogLevel, error) { return loglevel.ParseLogLevel(value) }

// SetLogLevel delegates to loglevel.SetLogLevel.
func SetLogLevel(level LogLevel) { loglevel.SetLogLevel(level) }

// GetLogLevel delegates to loglevel.GetLogLevel.
func GetLogLevel() LogLevel { return loglevel.GetLogLevel() }

// WarningEnabled reports whether warning messages should be emitted.
func WarningEnabled() bool { return loglevel.WarningEnabled() }

// InfoEnabled reports whether informational messages should be emitted.
func InfoEnabled() bool { return loglevel.InfoEnabled() }

// LogWarningf prints a warning message using pterm.Warning only when the
// current log level is `warning` or lower. The format string and arguments
// follow the usual fmt.Printf conventions and a trailing newline is appended
// when absent.
func LogWarningf(format string, args ...interface{}) {
	if !WarningEnabled() {
		return
	}
	msg := fmt.Sprintf(format, args...)
	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}
	pterm.Warning.Print(msg)
}

// LogInfof prints an informational message using pterm.Info only when the
// current log level is `info`. The format string and arguments follow the
// usual fmt.Printf conventions and a trailing newline is appended when absent.
func LogInfof(format string, args ...interface{}) {
	if !InfoEnabled() {
		return
	}
	msg := fmt.Sprintf(format, args...)
	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}
	pterm.Info.Print(msg)
}

// LogOrgSuccess prints a standard success message for a processed organization
// when informational logging is enabled.
func LogOrgSuccess(org string) {
	if !InfoEnabled() {
		return
	}
	pterm.Success.Printf("Successfully processed organization '%s'\n", org)
}
