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

const ENV_PASSWORD_SALT = "PASSWORD_SALT" // User Password salt
const ENV_AUTH_PREFIX = "AUTH_PREFIX"

const KEY_USER_NEED_ACTIVATE = "USER_NEED_ACTIVATE"
const KEY_API_NEED_AUTH = "API_NEED_AUTH"

// InitRabbit start with default middleware and auth handler
// 1. migrate models
// 2. gin middleware
// 3. setup env
// 4. setup config
// 5. auth handler
func InitRabbit(db *gorm.DB, r *gin.Engine) {
	// 1
	if err := InitMigrate(db); err != nil {
		log.Fatal("migrate fail: ", err)
	}

	// 2
	r.Use(WithGormDB(db), CORSEnabled())

	// 3
	secret := GetEnv(ENV_SESSION_SECRET)
	if secret != "" {
		r.Use(WithCookieSession(secret))
	} else {
		r.Use(WithMemSession(""))
	}

	// 4
	CheckValue(db, KEY_USER_NEED_ACTIVATE, "false")
	CheckValue(db, KEY_API_NEED_AUTH, "false")

	// 5
	RegisterAuthenticationHandlers("/auth", db, r)
}
