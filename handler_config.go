package rabbit

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterConfigHandlers(prefix string, db *gorm.DB, r *gin.Engine) {
	if prefix == "" {
		prefix = GetEnv(ENV_CONFIG_PREFIX)
	}

	// only super user can access
	cr := r.Group(prefix).Use(func(ctx *gin.Context) {
		user := CurrentUser(ctx)
		if user == nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not login"})
			return
		}

		method := ctx.Request.Method

		if !user.IsSuperUser && method != http.MethodPost {
			HandleTheError(ctx, ErrPermissionDenied)
			return
		}

		ctx.Next()
	})

	webobject := WebObject{
		Name:         "config",
		Model:        Config{},
		Editables:    []string{"Key", "Value", "Desc"},
		Filterables:  []string{"Key", "Value"},
		Searchables:  []string{"Key"},
		GetDB:        func(c *gin.Context, isCreate bool) *gorm.DB { return db },
		AllowMethods: QUERY | EDIT | DELETE,
	}

	webobject.RegisterObject(cr)
}
