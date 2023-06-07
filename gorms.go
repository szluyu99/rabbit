package rabbit

import (
	"reflect"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"gorm.io/gorm/utils"
)

func UpdateFields[T any](db *gorm.DB, model *T, vals map[string]any) error {
	return db.Model(model).Updates(vals).Error
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

func Count[T any](db *gorm.DB, where ...any) (int, error) {
	var count int64
	if len(where) > 0 {
		db = db.Where(where[0], where[1:]...)
	}
	result := db.Model(new(T)).Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}
	return int(count), nil
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

// GenUniqueKey generate unique key for field
func GenUniqueKey(tx *gorm.DB, field string, size int) (key string) {
	for i := 0; i < 10; i++ {
		key = RandText(size)
		var count int64
		if err := tx.Where(field, key).Limit(1).Count(&count).Error; err != nil {
			return ""
		}
		if count > 0 {
			continue
		}
		return key
	}

	return ""
}
