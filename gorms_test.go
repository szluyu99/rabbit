package rabbit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

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
		db.Create(&product{UUID: "aaaa", Name: "demo_product"})

		// SELECT * FROM `products` WHERE uuid = "aaaa" LIMIT 1
		val, err := GetByID[product](db, "aaaa")
		assert.Nil(t, err)
		assert.NotNil(t, val)

		// SELECT * FROM `products` WHERE `name` = "demo_product" AND uuid = "aaaa" LIMIT 1
		val, err = GetByID[product](db, "aaaa", "name = ? AND can_buy = ?", "demo_product", false)
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

func TestUpdateFields(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open("file::memory:"), nil)

	type User struct {
		UUID uint `gorm:"primarykey"`
		Name string
		Age  int
	}

	db.AutoMigrate(User{})

	var u = User{Name: "demo", Age: 10}
	db.Create(&u)

	err := UpdateFields(db, &u, map[string]any{
		"name": "updated",
		"age":  20,
	})
	assert.Equal(t, err, nil)
}

func TestUniqueKey(t *testing.T) {
	db := InitDatabase("", "", nil)
	db.AutoMigrate(&User{}, &Config{})

	v := GenUniqueKey(db.Model(User{}), "email", 10)
	assert.Equal(t, len(v), 10)

	v = GenUniqueKey(db.Model(User{}), "xx", 10)
	assert.Equal(t, len(v), 0)
}

func TestCount(t *testing.T) {
	db := InitDatabase("", "", nil)
	db.AutoMigrate(&user{})

	db.Create(&user{Name: "user1", Email: "user1@example.com", Age: 10})
	count, err := Count[user](db)
	assert.Nil(t, err)
	assert.Equal(t, 1, count)

	db.Create(&user{Name: "user2", Email: "user2@example.com", Age: 20})
	count, err = Count[user](db, "name", "user2")
	assert.Nil(t, err)
	assert.Equal(t, 1, count)

	db.Create(&user{Name: "user3", Email: "user3@example.com", Age: 30})
	count, err = Count[user](db, "name like ?", "user%")
	assert.Nil(t, err)
	assert.Equal(t, 3, count)
}
