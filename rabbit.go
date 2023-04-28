package rabbit

import (
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Gin session field
const SessionField = "rabbit"

// DB
const ENV_DB_DRIVER = "DB_DRIVER"
const ENV_DSN = "DSN"
const ENV_SESSION_SECRET = "SESSION_SECRET"

// User Password salt
const ENV_SALT = "PASSWORD_SALT"
const ENV_AUTH_PREFIX = "AUTH_PREFIX"

// User need to activate
const KEY_USER_ACTIVATED = "USER_ACTIVATED"

// InitRabbit start with default middleware and auth handler
// 1. migrate models
// 2. gin middleware
// 3. auth handler
func InitRabbit(db *gorm.DB, r *gin.Engine) {
	err := InitMigrate(db)
	if err != nil {
		log.Fatal("migrate fail: ", err)
	}

	r.Use(WithGormDB(db), CORSEnabled())

	secret := GetEnv(ENV_SESSION_SECRET)
	if secret != "" {
		r.Use(WithCookieSession(secret))
	} else {
		r.Use(WithMemSession(""))
	}

	InitAuthHandler("/auth", db, r)
}
