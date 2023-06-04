package rabbit

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUserProfile(t *testing.T) {
	db := InitDatabase("", "", nil)
	MakeMigrates(db, &User{})

	bob, _ := CreateUser(db, "bob@example.org", "123456")
	assert.Nil(t, bob.Profile)

	bob.Profile = &Profile{
		Avatar: "mock_img",
	}
	db.Save(bob)

	u, _ := GetUserByEmail(db, "bob@example.org")
	assert.Equal(t, u.Profile.Avatar, "mock_img")
}

func TestUserHashToken(t *testing.T) {
	db := InitDatabase("", "", nil)
	MakeMigrates(db, &User{}, &Config{})

	bob, _ := CreateUser(db, "bob@example.org", "123456")
	n := time.Now().Add(1 * time.Minute)
	hash := EncodeHashToken(bob, n.Unix(), true)
	log.Println(hash)
	u, err := DecodeHashToken(db, hash, true)
	assert.Nil(t, err)
	assert.NotNil(t, u)
	assert.Equal(t, u.ID, bob.ID)
}

func TestLogin(t *testing.T) {

}
