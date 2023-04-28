package rabbit

import (
	"os"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

func GetEnv(key string) string {
	v, _ := LookupEnv(key)
	return v
}

/*
1. Check .env file
2. Check environment variables
*/
func LookupEnv(key string) (string, bool) {
	// 1
	data, err := os.ReadFile(".env")
	if err != nil {
		// 2
		return os.LookupEnv(key)
	}
	lines := strings.Split(string(data), "\n")
	for i := 0; i < len(lines); i++ {
		v := strings.TrimSpace(lines[i])
		if v == "" || v[0] == '#' {
			continue
		}
		if !strings.Contains(v, "=") {
			continue
		}
		vs := strings.SplitN(v, "=", 2)
		if strings.EqualFold(strings.TrimSpace(vs[0]), key) {
			return strings.TrimSpace(vs[1]), true
		}
	}

	return "", false
}

func SetValue(db *gorm.DB, key, value string) {
	key = strings.ToUpper(key)

	var v Config
	result := db.Where("key", key).Take(&v)
	if result.Error != nil {
		newV := &Config{
			Key:   key,
			Value: value,
		}
		db.Create(&newV)
		return
	}
	db.Model(&Config{}).Where("key", key).UpdateColumn("value", value)
}

func GetValue(db *gorm.DB, key string) string {
	key = strings.ToUpper(key)

	var v Config
	result := db.Where("key", key).Take(&v)
	if result.Error != nil {
		return ""
	}

	return v.Value
}

func GetIntValue(db *gorm.DB, key string, defaultValue int) int {
	v := GetValue(db, key)

	if v == "" {
		return defaultValue
	}

	val, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return defaultValue
	}

	return int(val)
}

func GetBoolValue(db *gorm.DB, key string) bool {
	v := GetValue(db, key)

	if v == "" {
		return false
	}

	val, err := strconv.ParseBool(v)
	if err != nil {
		return false
	}

	return val
}

// CheckValue check if key exists, if not, set defaultValue
func CheckValue(db *gorm.DB, key, defaultValue string) {
	if GetValue(db, key) == "" {
		SetValue(db, key, defaultValue)
	}
}
