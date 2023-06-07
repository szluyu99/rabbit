package rabbit

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	// SigUserLogin: user *User, c *gin.Context
	SigUserLogin = "user.login"
	// SigUserLogout: user *User, c *gin.Context
	SigUserLogout = "user.logout"
	//SigUserCreate: user *User, c *gin.Context
	SigUserCreate = "user.create"
)

// set session
func Login(c *gin.Context, user *User) {
	db := c.MustGet(DbField).(*gorm.DB)

	SetLastLogin(db, user, c.ClientIP())

	session := sessions.Default(c)
	session.Set(UserField, user.ID)
	session.Save()

	Sig().Emit(SigUserLogin, user, c)
}

// 1. remove context
// 2. remove session
func Logout(c *gin.Context, user *User) {
	// 1
	c.Set(UserField, nil)

	// 2
	session := sessions.Default(c)
	session.Delete(UserField)
	session.Save()

	Sig().Emit(SigUserLogout, user, c)
}

// password
func CheckPassword(dbPassword, password string) bool {
	return dbPassword == HashPassword(password)
}

func SetPassword(db *gorm.DB, user *User, password string) (err error) {
	p := HashPassword(password)
	if err = UpdateFields(db, user, map[string]any{
		"Password": p,
	}); err != nil {
		return err
	}
	user.Password = p
	return err
}

func HashPassword(password string) string {
	salt := GetEnv(ENV_PASSWORD_SALT)
	hashVal := sha256.Sum256([]byte(salt + password))
	return fmt.Sprintf("sha256$%s%x", salt, hashVal)
}

// user
func GetUserByID(db *gorm.DB, userID uint) (*User, error) {
	var val User
	result := db.Where("id", userID).Where("Enabled", true).Preload("Roles").Take(&val)
	if result.Error != nil {
		return nil, result.Error
	}
	return &val, nil
}

func GetUserByEmail(db *gorm.DB, email string) (user *User, err error) {
	var val User
	result := db.Where("email", strings.ToLower(email)).Take(&val)
	if result.Error != nil {
		return nil, result.Error
	}
	return &val, nil
}

func IsExistByEmail(db *gorm.DB, email string) bool {
	_, err := GetUserByEmail(db, email)
	return err == nil
}

func CreateUser(db *gorm.DB, email, password string) (*User, error) {
	user := User{
		Email:     email,
		Password:  HashPassword(password),
		Enabled:   true,
		Activated: false,
	}
	result := db.Create(&user)
	return &user, result.Error
}

func SetLastLogin(db *gorm.DB, user *User, lastIp string) error {
	now := time.Now().Truncate(1 * time.Second)
	vals := map[string]any{
		"LastLoginIP": lastIp,
		"LastLogin":   &now,
	}
	user.LastLogin = &now
	user.LastLoginIP = lastIp
	return db.Model(user).Updates(vals).Error
}

/*
timestamp-uid-token
base64(email$timestamp) + "-" + sha256(salt + logintimestamp + password + email$timestamp)
*/
func EncodeHashToken(user *User, timestamp int64, useLastLogin bool) (hash string) {
	logintimestamp := "0"
	if useLastLogin && user.LastLogin != nil {
		logintimestamp = fmt.Sprintf("%d", user.LastLogin.Unix())
	}
	t := fmt.Sprintf("%s$%d", user.Email, timestamp)
	salt := GetEnv(ENV_PASSWORD_SALT)
	hashVal := sha256.Sum256([]byte(salt + logintimestamp + user.Password + t))
	hash = base64.RawStdEncoding.EncodeToString([]byte(t)) + "-" + fmt.Sprintf("%x", hashVal)
	return hash
}

/*
base64(email$timestamp) + "-" + sha256(salt + logintimestamp + password + email$timestamp)
*/
func DecodeHashToken(db *gorm.DB, hash string, useLastLogin bool) (user *User, err error) {
	vals := strings.Split(hash, "-")
	if len(vals) != 2 {
		return nil, errors.New("bad token")
	}
	data, err := base64.RawStdEncoding.DecodeString(vals[0])
	if err != nil {
		return nil, errors.New("bad token")
	}

	vals = strings.Split(string(data), "$")
	if len(vals) != 2 {
		return nil, errors.New("bad token")
	}

	ts, err := strconv.ParseInt(vals[1], 10, 64)
	if err != nil {
		return nil, errors.New("bad token")
	}

	// check expired
	if time.Now().Unix() > ts {
		return nil, errors.New("token expired")
	}

	user, err = GetUserByEmail(db, vals[0])
	if err != nil {
		return nil, errors.New("bad token")
	}

	token := EncodeHashToken(user, ts, useLastLogin)
	if token != hash {
		return nil, errors.New("bad token")
	}

	return user, nil
}
