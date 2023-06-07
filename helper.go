package rabbit

import (
	"math/rand"
	"reflect"
	"time"
)

func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

var letterRunes = []rune("0123456789abcdefghijklmnopqrstuvwxyz")
var numberRunes = []rune("0123456789")

func randRunes(n int, source []rune) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = source[rand.Intn(len(source))]
	}
	return string(b)
}

func RandText(n int) string {
	return randRunes(n, letterRunes)
}

func RandNumberText(n int) string {
	return randRunes(n, numberRunes)
}

func StructAsMap(form any, fields []string) (vals map[string]any) {
	vals = make(map[string]any)
	v := reflect.ValueOf(form)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return vals
	}
	for i := 0; i < len(fields); i++ {
		k := v.FieldByName(fields[i])
		if !k.IsValid() || k.IsZero() {
			continue
		}
		if k.Kind() == reflect.Ptr {
			if !k.IsNil() {
				vals[fields[i]] = k.Elem().Interface()
			}
		} else {
			vals[fields[i]] = k.Interface()
		}
	}
	return vals
}
