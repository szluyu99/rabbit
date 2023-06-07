package rabbit

import (
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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

func RegisterAuthenticationHandlers(prefix string, db *gorm.DB, r *gin.Engine) {
	r.GET(filepath.Join(prefix, "info"), handleUserInfo)
	r.POST(filepath.Join(prefix, "login"), handleUserSignin)
	r.POST(filepath.Join(prefix, "register"), handleUserSignup)
	r.GET(filepath.Join(prefix, "logout"), handleUserLogout)
	r.POST(filepath.Join(prefix, "change_password"), handleUserChangePassword)
}

func handleUserInfo(c *gin.Context) {
	user := CurrentUser(c)
	if user == nil {
		HandleErrorMessage(c, http.StatusForbidden, "user not login")
		return
	}
	c.JSON(http.StatusOK, user)
}

func handleUserSignin(c *gin.Context) {
	var form LoginForm
	if err := c.BindJSON(&form); err != nil {
		HandleError(c, http.StatusBadRequest, err)
		return
	}

	if form.Email == "" && form.AuthToken == "" {
		HandleErrorMessage(c, http.StatusBadRequest, "email is required")
		return
	}

	if form.Password == "" && form.AuthToken == "" {
		HandleErrorMessage(c, http.StatusBadRequest, "empty password")
		return
	}

	db := c.MustGet(DbField).(*gorm.DB)

	var user *User
	var err error

	// use password login or token login
	if form.Password != "" {
		user, err = GetUserByEmail(db, form.Email)
		if err != nil {
			HandleErrorMessage(c, http.StatusBadRequest, "user not exists")
			return
		}
		if !CheckPassword(user.Password, form.Password) {
			HandleErrorMessage(c, http.StatusUnauthorized, "unauthorized")
			return
		}
	} else {
		user, err = DecodeHashToken(db, form.AuthToken, false)
		if err != nil {
			HandleError(c, http.StatusUnauthorized, err)
			return
		}
	}

	if !user.Enabled {
		HandleErrorMessage(c, http.StatusForbidden, "user not allow login")
		return
	}

	// if need activated
	if GetBoolValue(db, KEY_USER_NEED_ACTIVATE) && !user.Activated {
		HandleErrorMessage(c, http.StatusUnauthorized, "waiting for activation")
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
		HandleError(c, http.StatusBadRequest, err)
		return
	}

	db := c.MustGet(DbField).(*gorm.DB)
	if IsExistByEmail(db, form.Email) {
		HandleErrorMessage(c, http.StatusBadRequest, "email has exists")
		return
	}

	user, err := CreateUser(db, form.Email, form.Password)
	if err != nil {
		HandleError(c, http.StatusBadRequest, err)
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

	err = UpdateFields(db, user, vals)
	if err != nil {
		log.Println("update user fields fail id:", user.ID, vals, err)
	}

	Sig().Emit(SigUserCreate, user, c)

	r := gin.H{
		"email":      user.Email,
		"activation": user.Activated,
	}

	if GetBoolValue(db, KEY_USER_NEED_ACTIVATE) && !user.Activated {
		// sendHashMail(db, user, SigUserVerifyEmail, KEY_VERIFY_EMAIL_EXPIRED, "180d", c.ClientIP(), c.Request.UserAgent())
		r["expired"] = "180d"
	} else {
		Login(c, user) // Login now
	}

	c.JSON(http.StatusOK, r)
}

func handleUserLogout(c *gin.Context) {
	if user := CurrentUser(c); user != nil {
		Logout(c, user)
	}
	c.JSON(http.StatusOK, gin.H{})
}

func handleUserChangePassword(c *gin.Context) {
	var form ChangePasswordForm
	if err := c.BindJSON(&form); err != nil {
		HandleError(c, http.StatusBadRequest, err)
		return
	}

	user := CurrentUser(c)
	if user == nil {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	if !user.Enabled {
		HandleErrorMessage(c, http.StatusForbidden, "user not allow login")
		return
	}

	db := c.MustGet(DbField).(*gorm.DB)
	if GetBoolValue(db, KEY_USER_NEED_ACTIVATE) && !user.Activated {
		HandleErrorMessage(c, http.StatusUnauthorized, "waiting for activation")
		return
	}

	if err := SetPassword(db, user, form.Password); err != nil {
		HandleErrorMessage(c, http.StatusBadRequest, "password changed fail")
		return
	}

	c.JSON(http.StatusOK, true)
}
