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

func InTimezone(c *gin.Context, timezone string) {
	tz, err := time.LoadLocation(timezone)
	if err != nil {
		return
	}
	c.Set(TzField, tz)

	session := sessions.Default(c)
	session.Set(TzField, timezone)
	session.Save()
}

// gin context
func CurrentTimezone(c *gin.Context) *time.Location {
	if cachedObj, exists := c.Get(TzField); exists && cachedObj != nil {
		return cachedObj.(*time.Location)
	}

	session := sessions.Default(c)
	tzkey := session.Get(TzField)

	if tzkey == nil {
		if user := CurrentUser(c); user != nil {
			tzkey = user.Timezone
		}
	}

	var tz *time.Location
	defer func() {
		if tz == nil {
			tz = time.UTC
		}
		c.Set(TzField, tz)
	}()

	if tzkey == nil {
		return time.UTC
	}

	tz, _ = time.LoadLocation(tzkey.(string))
	if tz == nil {
		return time.UTC
	}
	return tz
}

/*
1. try get from token
2. try get from session
*/
func CurrentUserID(c *gin.Context) uint {
	userID := c.GetUint(UserField)

	if userID == 0 {
		session := sessions.Default(c)
		val := session.Get(UserField)
		if val == nil {
			return 0
		}
		userID = val.(uint)
	}

	return userID
}

/*
1. try get user from context cache
2. try get user from token/session
3. set context cache
*/
func CurrentUser(c *gin.Context) *User {
	// 1
	if cached, exist := c.Get(CacheUserField); exist && cached != nil {
		return cached.(*User)
	}

	// 2
	userID := CurrentUserID(c)

	// 3
	db := c.MustGet(DbField).(*gorm.DB)
	user, err := GetUserByID(db, userID)
	if err != nil {
		return nil
	}

	c.Set(CacheUserField, user)
	return user
}

func Login(c *gin.Context, user *User) {
	db := c.MustGet(DbField).(*gorm.DB)
	SetLastLogin(db, user, c.ClientIP())

	session := sessions.Default(c)
	session.Set(UserField, user.ID)
	session.Save()

	Sig().Emit(SigUserLogin, user, c)
}

func Logout(c *gin.Context, user *User) {
	c.Set(UserField, nil)

	session := sessions.Default(c)
	session.Delete(UserField)
	session.Save()

	Sig().Emit(SigUserLogout, user, c)
}

// password
func CheckPassword(user *User, password string) bool {
	return user.Password == HashPassword(password)
}

func SetPassword(db *gorm.DB, user *User, password string) (err error) {
	p := HashPassword(password)
	err = UpdateUserFields(db, user, map[string]any{
		"Password": p,
	})
	if err != nil {
		return
	}
	user.Password = p
	return
}

func HashPassword(password string) string {
	salt := GetEnv(ENV_SALT)
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

func UpdateUserFields(db *gorm.DB, user *User, vals map[string]any) error {
	return db.Model(user).Updates(vals).Error
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
	salt := GetEnv(ENV_SALT)
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
