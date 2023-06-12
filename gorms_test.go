package rabbit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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

func TestGetPkColumnName(t *testing.T) {
	{
		type User struct {
			ID int64
		}
		assert.Equal(t, "id", GetPkColumnName[User]())
	}
	{
		type User struct {
			ID int64 `gorm:"primary_key"`
		}
		assert.Equal(t, "id", GetPkColumnName[User]())
	}
	{
		type User struct {
			UUID int64 `gorm:"primary_key"`
		}
		assert.Equal(t, "uuid", GetPkColumnName[User]())
	}
	{
		type User struct {
			UUID int64 `gorm:"primaryKey"`
		}
		assert.Equal(t, "uuid", GetPkColumnName[User]())
	}
}
