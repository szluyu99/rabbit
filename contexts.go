package rabbit

import (
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	DbField    = "_rabbit_db"
	TzField    = "_rabbit_tz"
	UserField  = "_rabbit_uid" // for session: uid, for context: *User
	GroupField = "_rabbit_gid" // for session: gid, for context: *Group
)

// 1. set [*time.Location] to gin context, for cache
// 2. set [timezone string] to session
func InTimezone(c *gin.Context, timezone string) {
	tz, err := time.LoadLocation(timezone)
	if err != nil {
		return
	}

	// 1
	c.Set(TzField, tz)

	// 2
	session := sessions.Default(c)
	session.Set(TzField, timezone)
	session.Save()
}

/*
1. try get cache from context
2. try get from session
3. set context cache
*/
func CurrentTimezone(c *gin.Context) *time.Location {
	// 1
	if cache, exist := c.Get(TzField); exist && cache != nil {
		return cache.(*time.Location)
	}

	// 2
	session := sessions.Default(c)
	tzKey := session.Get(TzField)

	if tzKey == nil {
		if user := CurrentUser(c); user != nil {
			tzKey = user.Timezone
		}
	}

	var tz *time.Location
	defer func() {
		// 3
		if tz == nil {
			tz = time.UTC
		}
		c.Set(TzField, tz)
	}()

	if tzKey == nil {
		return time.UTC
	}

	tz, _ = time.LoadLocation(tzKey.(string))
	if tz == nil {
		return time.UTC
	}
	return tz
}

/*
1. try get cache from context
2. try get user from token/session
3. set context cache
*/
func CurrentUser(c *gin.Context) *User {
	// 1
	if cache, exist := c.Get(UserField); exist && cache != nil {
		return cache.(*User)
	}

	// 2
	session := sessions.Default(c)
	uid := session.Get(UserField)
	if uid == nil {
		return nil
	}

	// 3
	db := c.MustGet(DbField).(*gorm.DB)
	user, err := GetUserByID(db, uid.(uint))
	if err != nil {
		return nil
	}

	c.Set(UserField, user)
	return user
}

func CurrentGroup(c *gin.Context) *Group {
	if cache, exists := c.Get(GroupField); exists && cache != nil {
		return cache.(*Group)
	}

	session := sessions.Default(c)
	gid := session.Get(GroupField)
	if gid == nil {
		return nil
	}

	db := c.MustGet(DbField).(*gorm.DB)
	group, err := GetGroupByID(db, gid.(uint))
	if err != nil {
		return nil
	}
	c.Set(GroupField, group)
	return group
}

func SwitchGroup(c *gin.Context, gid uint) {
	session := sessions.Default(c)
	session.Set(GroupField, gid)
	session.Save()
}
