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

	Debugln("debug")
	assert.Contains(t, mockData.String(), "[DEBUG] debug")

	Infoln("info")
	assert.Contains(t, mockData.String(), "[INFO] info")

	Warningln("warning")
	assert.Contains(t, mockData.String(), "[WARNING] warning")

	Errorln("error")
	assert.Contains(t, mockData.String(), "[ERROR] error")
}
