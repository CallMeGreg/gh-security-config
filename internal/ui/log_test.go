package ui

import "testing"

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    LogLevel
		wantErr bool
	}{
		{name: "empty defaults to warning", input: "", want: LogLevelWarning},
		{name: "info", input: "info", want: LogLevelInfo},
		{name: "warning", input: "warning", want: LogLevelWarning},
		{name: "error", input: "error", want: LogLevelError},
		{name: "mixed case", input: "Info", want: LogLevelInfo},
		{name: "padded", input: "  warning  ", want: LogLevelWarning},
		{name: "invalid", input: "verbose", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseLogLevel(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseLogLevel(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseLogLevel(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestInfoEnabled(t *testing.T) {
	t.Cleanup(func() { SetLogLevel(LogLevelWarning) })

	SetLogLevel(LogLevelInfo)
	if !InfoEnabled() {
		t.Errorf("InfoEnabled() = false, want true for LogLevelInfo")
	}
	SetLogLevel(LogLevelWarning)
	if InfoEnabled() {
		t.Errorf("InfoEnabled() = true, want false for LogLevelWarning")
	}
	SetLogLevel(LogLevelError)
	if InfoEnabled() {
		t.Errorf("InfoEnabled() = true, want false for LogLevelError")
	}
}
