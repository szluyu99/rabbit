package rabbit

import (
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	UserField  = "_rabbit_uid"
	GroupField = "_rabbit_gid"
	DbField    = "_rabbit_db"
	TzField    = "_rabbit_tz"
)

type RegisterUserForm struct {
	Email       string `json:"email" binding:"required"`
	Password    string `json:"password" binding:"required"`
	DisplayName string `json:"displayName"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Locale      string `json:"locale"`
	Timezone    string `json:"timezone"`
	Source      string `json:"source"`
}

type LoginForm struct {
	Email     string `json:"email"`
	Password  string `json:"password,omitempty"`
	Timezone  string `json:"timezone,omitempty"`
	Remember  bool   `json:"remember,omitempty"`
	AuthToken string `json:"token,omitempty"`
}

type ChangePasswordForm struct {
	Password string `json:"password" binding:"required"`
}

func InitAuthHandler(prefix string, db *gorm.DB, r *gin.Engine) {
	if prefix == "" {
		prefix = GetEnv(ENV_AUTH_PREFIX)
	}

	r.GET(filepath.Join(prefix, "info"), handleUserInfo)
	r.POST(filepath.Join(prefix, "login"), handleUserSignin)
	r.POST(filepath.Join(prefix, "register"), handleUserSignup)
	r.GET(filepath.Join(prefix, "logout"), handleUserLogout)
	r.POST(filepath.Join(prefix, "change_password"), handleUserChangePassword)
}

func handleUserInfo(c *gin.Context) {
	user := CurrentUser(c)
	if user == nil {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	c.JSON(http.StatusOK, user)
}

func handleUserSignin(c *gin.Context) {
	var form LoginForm
	if err := c.BindJSON(&form); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if form.AuthToken == "" && form.Email == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "email is required"})
		return
	}

	if form.Password == "" && form.AuthToken == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "empty password"})
		return
	}

	db := c.MustGet(DbField).(*gorm.DB)

	var user *User
	var err error

	// use password login or token login
	if form.Password != "" {
		user, err = GetUserByEmail(db, form.Email)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "user not exists"})
			return
		}
		if !CheckPassword(user, form.Password) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
	} else {
		user, err = DecodeHashToken(db, form.AuthToken, false)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
	}

	if !user.Enabled {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "user not allow login"})
		return
	}

	if GetBoolValue(db, KEY_USER_ACTIVATED) && !user.Activated {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "waiting for activation"})
		return
	}

	if form.Timezone != "" {
		InTimezone(c, form.Timezone)
	}

	Login(c, user)

	if form.Remember {
		// 7 days
		n := time.Now().Add(7 * 24 * time.Hour)
		user.AuthToken = EncodeHashToken(user, n.Unix(), false)
	}

	c.JSON(http.StatusOK, user)
}

func handleUserSignup(c *gin.Context) {
	var form RegisterUserForm
	if err := c.BindJSON(&form); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := c.MustGet(DbField).(*gorm.DB)
	if IsExistsByEmail(db, form.Email) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "email has exists"})
		return
	}

	user, err := CreateUser(db, form.Email, form.Password)
	if err != nil {
		log.Println("create user fail", form, err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	vals := StructAsMap(form, []string{
		"DisplayName",
		"FirstName",
		"LastName",
		"Locale",
		"Timezone",
		"Source",
	})

	n := time.Now().Truncate(1 * time.Second)
	vals["LastLogin"] = &n
	vals["LastLoginIP"] = c.ClientIP()

	err = UpdateUserFields(db, user, vals)
	if err != nil {
		log.Println("update user fields fail id:", user.ID, vals, err)
	}

	Sig().Emit(SigUserCreate, user, c)

	r := gin.H{
		"email":      user.Email,
		"activation": user.Activated,
	}

	if !user.Activated && GetBoolValue(db, KEY_USER_ACTIVATED) {
		// sendHashMail(db, user, SigUserVerifyEmail, KEY_VERIFY_EMAIL_EXPIRED, "180d", c.ClientIP(), c.Request.UserAgent())
		r["expired"] = "180d"
	} else {
		Login(c, user) // Login now
	}

	c.JSON(http.StatusOK, r)
}

func handleUserLogout(c *gin.Context) {
	user := CurrentUser(c)
	if user != nil {
		Logout(c, user)
	}
	c.JSON(http.StatusOK, gin.H{})
}

func handleUserChangePassword(c *gin.Context) {
	var form ChangePasswordForm
	if err := c.BindJSON(&form); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := CurrentUser(c)
	if user == nil {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	if !user.Enabled {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "user not allow login"})
		return
	}

	db := c.MustGet(DbField).(*gorm.DB)

	if GetBoolValue(db, KEY_USER_ACTIVATED) && !user.Activated {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "waiting for activation",
		})
		return
	}

	err := SetPassword(db, user, form.Password)
	if err != nil {
		log.Println("changed user password fail user:", user.ID, err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "changed fail"})
		return
	}

	c.JSON(http.StatusOK, true)
}
