package rabbit

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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

func TestConfigFunctions(t *testing.T) {
	// 初始化数据库
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	db.AutoMigrate(&Config{})

	// Test SetValue and GetValue
	SetValue(db, "test_key", "test_value")
	value := GetValue(db, "test_key")
	assert.Equal(t, value, "test_value")

	// Test CheckValue
	CheckValue(db, "check_key", "default_value")
	value = GetValue(db, "check_key")
	assert.Equal(t, value, "default_value")

	// Test GetIntValue
	SetValue(db, "int_key", "42")
	intValue := GetIntValue(db, "int_key", -1)
	assert.Equal(t, intValue, 42)

	// Test GetBoolValue
	SetValue(db, "bool_key", "true")
	boolValue := GetBoolValue(db, "bool_key")
	assert.Equal(t, boolValue, true)
}
