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
