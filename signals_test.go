package rabbit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSignals(t *testing.T) {
	var val string = ""
	Sig().Connect("mock_test", func(sender any, params ...any) {
		val = sender.(string)
	})
	Sig().Connect("mock_test", func(sender any, params ...any) {
		val += sender.(string)
	})
	Sig().Emit("mock_test", "abc")
	assert.Equal(t, "abcabc", val)
	Sig().DisConnect("mock_test")
}
