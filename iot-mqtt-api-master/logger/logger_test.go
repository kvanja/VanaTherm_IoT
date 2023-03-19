package logger

import (
	"log"
	"os"
	"testing"
)

func TestLogger(t *testing.T) {
	logFile := "logger_test.log"

	l, err := New(logFile)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	l.Debug("Debug test")
	l.Error("Error test")
	l.Info("Info test")
	l.Warning("Warning test")

	os.Remove(logFile)
}
