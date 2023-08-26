package logging

import (
	"bytes"
	"os"
)

func MockLogger() (Logger, error) {
	logFile, err := os.CreateTemp("", "test-log.txt")
	if err != nil {
		return nil, err
	}

	defer os.Remove(logFile.Name())
	small := make([]byte, 20)
	consoleMock := bytes.NewBuffer(small)
	cfg := NewLogerConfig(true, "debug")
	logger := NewLogger(cfg)
	logger.InitLogger(consoleMock, logFile)
	return logger, nil
}
