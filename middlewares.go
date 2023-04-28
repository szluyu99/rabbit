package rabbit

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const CORS_ALLOW_ALL = "*"
const CORS_ALLOW_CREDENTIALS = "true"
const CORS_ALLOW_HEADERS = "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Auth-Token"
const CORS_ALLOW_METHODS = "POST, OPTIONS, GET, PUT, PATCH, DELETE"
const XAuthTokenHeader = "X-Auth-Token"

func CORSEnabled() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", CORS_ALLOW_ALL)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", CORS_ALLOW_CREDENTIALS)
		c.Writer.Header().Set("Access-Control-Allow-Headers", CORS_ALLOW_HEADERS)
		c.Writer.Header().Set("Access-Control-Allow-Methods", CORS_ALLOW_METHODS)

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent) // 204
			return
		}
		c.Next()
	}
}

func WithCookieSession(secret string) gin.HandlerFunc {
	store := cookie.NewStore([]byte(secret))
	store.Options(sessions.Options{Path: "/", MaxAge: 0})
	return sessions.Sessions(SessionField, store)
}

func WithMemSession(secret string) gin.HandlerFunc {
	store := memstore.NewStore([]byte(secret))
	store.Options(sessions.Options{Path: "/", MaxAge: 0})
	return sessions.Sessions(SessionField, store)
}

func WithGormDB(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set(DbField, db)
		ctx.Next()
	}
}
