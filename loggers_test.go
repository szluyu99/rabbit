package rabbit

import (
	"bytes"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	mockData := bytes.NewBufferString("")
	log.Default().SetOutput(mockData)
	SetLogLevel(LevelDebug)

	Debug("debug")
	assert.Contains(t, mockData.String(), "[DEBUG] debug")

	Info("info")
	assert.Contains(t, mockData.String(), "[INFO] info")

	Warning("warning")
	assert.Contains(t, mockData.String(), "[WARNING] warning")

	Error("error")
	assert.Contains(t, mockData.String(), "[ERROR] error")
}
