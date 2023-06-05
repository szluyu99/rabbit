package rabbit

import (
	"errors"
	"net/http"
	"strings"

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

// 1. auth from session
// 2. auth from token
func WithAuthentication() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// session
		uid := CurrentUserID(ctx)
		if uid != 0 {
			ctx.Next()
			return
		}

		// token
		authValue := ctx.Request.Header.Get("Authorization")
		if authValue == "" {
			HandleErrorMsg(ctx, http.StatusUnauthorized, "authorization header not found")
			return
		}

		vals := strings.Split(authValue, " ")
		if len(vals) != 2 || vals[0] != "Bearer" {
			HandleErrorMsg(ctx, http.StatusUnauthorized, "invalid authorization header")
			return
		}

		db := ctx.MustGet(DbField).(*gorm.DB)
		token := vals[1]

		user, err := DecodeHashToken(db, token, false)
		if err != nil {
			HandleError(ctx, http.StatusUnauthorized, err)
			return
		}

		ctx.Set(UserField, user.ID)
		ctx.Next()
	}
}

// check if the user has permission to access the url
// superuser no need to check
func WithAuthorization(prefix string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		db := ctx.MustGet(DbField).(*gorm.DB)

		if !GetBoolValue(db, KEY_API_NEED_AUTH) {
			ctx.Next()
		}

		url := ctx.FullPath()[len(prefix):]
		method := ctx.Request.Method

		user := CurrentUser(ctx)
		if user == nil {
			HandleError(ctx, http.StatusUnauthorized, errors.New("user need login"))
			return
		}

		if !user.IsSuperUser {
			pass, err := CheckUserPermission(db, user.ID, url, method)
			if err != nil || !pass {
				HandleTheError(ctx, ErrPermissionDenied)
				return
			}
		}

		ctx.Next()
	}
}
