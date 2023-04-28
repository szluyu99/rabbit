package rabbit

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnv(t *testing.T) {
	v := GetEnv("NOT_EXIST_ENV")
	assert.Empty(t, v)
	defer os.Remove(".env")

	os.WriteFile(".env", []byte(`
	#hello
	xx
	EXIST_ENV = 100	
	`), 0666)

	v = GetEnv("EXIST_ENV")
	assert.Equal(t, v, "100")
}
