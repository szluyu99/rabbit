package rabbit

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type testMapForm struct {
	ID     uint    `json:"id" binding:"required"`
	Title  *string `json:"title"`
	Source *string `json:"source"`
}

func TestFormAsMap(t *testing.T) {
	title := "title"
	form := testMapForm{
		ID:    100,
		Title: &title,
	}
	{
		vals := StructAsMap(form, []string{"Title", "Target"})
		assert.Equal(t, 1, len(vals))
		assert.Equal(t, title, vals["Title"])
	}
	{
		vals := StructAsMap(form, []string{"ID", "Source"})
		assert.Equal(t, 1, len(vals))
		assert.Equal(t, uint(100), vals["ID"])
	}
	{
		vals := StructAsMap(&form, []string{"ID", "Title"})
		assert.Equal(t, 2, len(vals))
		assert.Equal(t, uint(100), vals["ID"])
		assert.Equal(t, title, vals["Title"])
	}
}

func TestRandText(t *testing.T) {
	v := RandText(10)
	assert.Equal(t, len(v), 10)

	v2 := RandNumberText(5)
	assert.Equal(t, len(v2), 5)
	_, err := strconv.ParseInt(v2, 10, 64)
	assert.Nil(t, err)
}

type user struct {
	UUID      uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Name      string
	Email     string
	Age       int
	Enabled   bool
}

type product struct {
	UUID      string `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Name      string
	CanBuy    bool
}

func TestGetByID(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{SkipDefaultTransaction: true})
	db.AutoMigrate(user{}, product{})

	{
		db.Create(&user{Name: "demo", Enabled: true})

		val, err := GetByID[user](db, 1)
		assert.Nil(t, err)
		assert.NotEmpty(t, val.UUID)

		val, err = GetByID[user](db, 1, "name = ? AND enabled = ?", "demo", true)
		assert.Nil(t, err)
		assert.NotEmpty(t, val.UUID)
	}
	{
		db.Create(&product{UUID: "aaaa", Name: "demoproduct"})

		// SELECT * FROM `products` WHERE uuid = "aaaa" LIMIT 1
		val, err := GetByID[product](db, "aaaa")
		assert.Nil(t, err)
		assert.NotNil(t, val)

		// SELECT * FROM `products` WHERE `name` = "demoproduct" AND uuid = "aaaa" LIMIT 1
		val, err = GetByID[product](db, "aaaa", "name = ? AND can_buy = ?", "demoproduct", false)
		assert.Nil(t, err)
		assert.NotNil(t, val)
	}
}

func TestGet(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{SkipDefaultTransaction: true})
	db.AutoMigrate(user{}, product{})

	db.Create(&user{Name: "demo", Enabled: true})
	{
		val, err := Get(db, &user{})
		assert.Nil(t, err)
		assert.NotNil(t, val)
	}
	{
		val, err := Get(db, &user{Name: "demo", Enabled: true})
		assert.Nil(t, err)
		assert.NotNil(t, val)
	}
	{
		val, err := Get(db, &user{}, "enabled", true)
		assert.Nil(t, err)
		assert.NotNil(t, val)
	}
}

func TestGetColumnName(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{SkipDefaultTransaction: true})
	db.AutoMigrate(user{}, product{})

	type test struct {
		UUID      uint `gorm:"primarykey"`
		Name      string
		CreatedAt time.Time
		AName     string `gorm:"column:a_n"`
		BName     string
	}

	assert.Equal(t, "uuid", GetColumnNameByField[test]("UUID"))
	assert.Equal(t, "created_at", GetColumnNameByField[test]("CreatedAt"))
	assert.Equal(t, "a_n", GetColumnNameByField[test]("AName"))
	assert.Equal(t, "b_name", GetColumnNameByField[test]("BName"))

	assert.Equal(t, "uuid", GetPkColumnName[test]())
}

func TestGetFieldNameByJSONTag(t *testing.T) {
	type test struct {
		UUID      uint      `json:"id"`
		Name      string    `json:"name"`
		CreatedAt time.Time `json:"createdAt"`
	}

	assert.Equal(t, "UUID", GetFieldNameByJsonTag[test]("id"))
	assert.Equal(t, "Name", GetFieldNameByJsonTag[test]("name"))
	assert.Equal(t, "CreatedAt", GetFieldNameByJsonTag[test]("createdAt"))
}

func TestGetTableName(t *testing.T) {

	type test struct {
		UUID      uint      `json:"id"`
		Name      string    `json:"name"`
		CreatedAt time.Time `json:"createdAt"`
	}

	{
		db, _ := gorm.Open(sqlite.Open("file::memory:"), nil)
		tblName := GetTableName[test](db)
		assert.Equal(t, "tests", tblName)
	}
	{
		db, _ := gorm.Open(sqlite.Open("file::memory:"),
			&gorm.Config{NamingStrategy: schema.NamingStrategy{SingularTable: true}},
		)
		tblName := GetTableName[test](db)
		assert.Equal(t, "test", tblName)
	}
}

func TestGetPkJsonName(t *testing.T) {
	type test1 struct {
		UUID uint `json:"id" gorm:"primaryKey"`
	}
	assert.Equal(t, "id", GetPkJsonName[test1]())

	type test2 struct {
		UUID uint `gorm:"primaryKey"`
	}
	assert.Equal(t, "UUID", GetPkJsonName[test2]())
}
