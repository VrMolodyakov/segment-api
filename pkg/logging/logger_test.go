package logging

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	tests := []struct {
		title         string
		input         string
		find          string
		level         string
		expected      bool
		loggingMethod func(logger *apiLogger, input string)
	}{
		{
			title: "Test Debug",
			input: "debug message",
			find:  "debug message",
			loggingMethod: func(logger *apiLogger, input string) {
				logger.Debug(input)
			},
			level:    "debug",
			expected: true,
		},
		{
			title: "Test Info",
			input: "info message",
			find:  "info message",
			loggingMethod: func(logger *apiLogger, input string) {
				logger.Info(input)
			},
			level:    "info",
			expected: true,
		},
		{
			title: "Test Warn",
			input: "warn message",
			find:  "warn message",
			loggingMethod: func(logger *apiLogger, input string) {
				logger.Warn(input)
			},
			level:    "warn",
			expected: true,
		},
		{
			title: "Test Error",
			input: "error message",
			find:  "error message",
			loggingMethod: func(logger *apiLogger, input string) {
				logger.Error(input)
			},
			level:    "error",
			expected: true,
		},
		{
			title: "Test Debug message with Info level",
			input: "debug message",
			find:  "debug message",
			loggingMethod: func(logger *apiLogger, input string) {
				logger.Debug(input)
			},
			level:    "info",
			expected: false,
		},
		{
			title: "Test Info message with Debug level",
			input: "info message",
			find:  "info message",
			loggingMethod: func(logger *apiLogger, input string) {
				logger.Info(input)
			},
			level:    "debug",
			expected: true,
		},
		{
			title: "Test Info message with Warn level",
			input: "info message",
			find:  "info message",
			loggingMethod: func(logger *apiLogger, input string) {
				logger.Info(input)
			},
			level:    "warn",
			expected: false,
		},
		{
			title: "Test Warn message with Info level",
			input: "error message",
			find:  "error message",
			loggingMethod: func(logger *apiLogger, input string) {
				logger.Warn(input)
			},
			level:    "info",
			expected: true,
		},
		{
			title: "Test Warn message with Error level",
			input: "warn message",
			find:  "warn message",
			loggingMethod: func(logger *apiLogger, input string) {
				logger.Warn(input)
			},
			level:    "error",
			expected: false,
		},
		{
			title: "Test Error message with Warn level",
			input: "error message",
			find:  "error message",
			loggingMethod: func(logger *apiLogger, input string) {
				logger.Error(input)
			},
			level:    "warn",
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			logFile, err := os.CreateTemp("", "test-log.txt")
			if err != nil {
				assert.NoError(t, err)
			}

			defer os.Remove(logFile.Name())
			small := make([]byte, 20)
			consoleMock := bytes.NewBuffer(small)
			cfg := NewLogerConfig(true, test.level)
			logger := NewLogger(cfg)
			logger.InitLogger(consoleMock, logFile)
			test.loggingMethod(logger, test.input)

			logBytes, err := os.ReadFile(logFile.Name())
			if err != nil {
				t.Fatalf("Error reading log file: %v", err)
			}
			content := string(logBytes)
			actual := strings.Contains(content, test.find)
			actual = actual && strings.Contains(consoleMock.String(), test.find)
			assert.Equal(t, test.expected, actual)
		})
	}
}
