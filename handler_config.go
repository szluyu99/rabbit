package rabbit

// func RegisterConfigHandlers(prefix string, db *gorm.DB, r gin.IRoutes) {
// 	// only super user can access
// 	cr := r.Group(prefix).Use(func(ctx *gin.Context) {
// 		user := CurrentUser(ctx)
// 		if user == nil {
// 			HandleErrorMsg(ctx, http.StatusUnauthorized, "user not login")
// 			return
// 		}

// 		method := ctx.Request.Method

// 		if !user.IsSuperUser && method != http.MethodPost {
// 			HandleTheError(ctx, ErrPermissionDenied)
// 			return
// 		}

// 		ctx.Next()
// 	})

// 	RegisterObject(cr, &WebObject{
// 		Name:         "config",
// 		Model:        Config{},
// 		Editables:    []string{"Key", "Value", "Desc"},
// 		Filterables:  []string{"Key", "Value"},
// 		Searchables:  []string{"Key", "Desc"},
// 		GetDB:        func(c *gin.Context, isCreate bool) *gorm.DB { return db },
// 		AllowMethods: QUERY | EDIT | DELETE,
// 	})
// }
