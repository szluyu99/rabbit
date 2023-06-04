package rabbit

import (
	"math/rand"
	"reflect"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"gorm.io/gorm/utils"
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

func GetByID[T any, E ~uint | ~int | ~string](db *gorm.DB, id E, where ...any) (*T, error) {
	var val T

	if len(where) > 0 {
		db = db.Where(where[0], where[1:]...)
	}

	result := db.Take(&val, GetPkColumnName[T](), id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &val, nil
}

func Get[T any](db *gorm.DB, val *T, where ...any) (*T, error) {
	if len(where) > 0 {
		db = db.Where(where[0], where[1:]...)
	}

	result := db.Where(val).Take(val)
	if result.Error != nil {
		return nil, result.Error
	}
	return val, nil
}

func GetPkColumnName[T any]() string {
	rt := reflect.TypeOf(new(T)).Elem()

	var columnName string
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		tagSetting := schema.ParseTagSetting(field.Tag.Get("gorm"), ";")
		isPrimaryKey := utils.CheckTruth(tagSetting["PRIMARYKEY"], tagSetting["PRIMARY_KEY"])
		if isPrimaryKey {
			name, ok := tagSetting["COLUMN"]
			if !ok {
				namingStrategy := schema.NamingStrategy{}
				name = namingStrategy.ColumnName("", field.Name)
			}
			columnName = name
			break
		}
	}
	if columnName == "" {
		return "id"
	}
	return columnName
}

func GetPkJsonName[T any]() string {
	rt := reflect.TypeOf(new(T)).Elem()
	return getPkJsonName(rt)
}

func GetColumnNameByField[T any](name string) string {
	rt := reflect.TypeOf(new(T)).Elem()
	val, ok := getColumnNameByField(rt, name)
	if ok {
		return val
	}
	return ""
}

func GetFieldNameByJsonTag[T any](jsonTag string) string {
	rt := reflect.TypeOf(new(T)).Elem()
	f, ok := getFieldByJsonTag(rt, jsonTag)
	if ok {
		return f.Name
	}
	return ""
}

func GetTableName[T any](db *gorm.DB) string {
	name := reflect.TypeOf(new(T)).Elem().Name()
	return db.NamingStrategy.TableName(name)
}

func getPkJsonName(rt reflect.Type) string {
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		gormTag := field.Tag.Get("gorm")
		if gormTag != "primarykey" && gormTag != "primaryKey" {
			continue
		}
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			return field.Name
		}
		return jsonTag
	}
	return ""
}

func getColumnNameByField(rt reflect.Type, name string) (string, bool) {
	field, ok := rt.FieldByName(name)
	if !ok {
		return "", false
	}

	tagSetting := schema.ParseTagSetting(field.Tag.Get("gorm"), ";")
	val, ok := tagSetting["COLUMN"]
	if !ok {
		namingStrategy := schema.NamingStrategy{}
		val = namingStrategy.ColumnName("", field.Name)
	}
	return val, true
}

func getFieldByJsonTag(rt reflect.Type, jsonTag string) (reflect.StructField, bool) {
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		if field.Tag.Get("json") == jsonTag {
			return field, true
		}
	}
	return reflect.StructField{}, false
}
