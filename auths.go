package rabbit

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// group
func GetGroupByID(db *gorm.DB, groupID uint) (*Group, error) {
	return GetByID[Group](db, groupID)
}

func GetGroupByName(db *gorm.DB, name string) (*Group, error) {
	return Get(db, &Group{Name: name})
}

func GetGroupsByUser(db *gorm.DB, userID uint) ([]*Group, error) {
	var user User
	result := db.Model(&User{}).Preload("Groups").Take(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	return user.Groups, nil
}

func GetFirstGroupByUser(db *gorm.DB, userID uint) (*Group, error) {
	var member GroupMember
	result := db.Where("user_id", userID).Preload("Group").Take(&member)
	if result.Error != nil {
		return nil, result.Error
	}
	return &member.Group, nil
}

func CreateGroupByUser(db *gorm.DB, userID uint, name string) (*Group, error) {
	group := Group{
		Name: name,
	}
	result := db.Create(&group)
	if result.Error != nil {
		return nil, result.Error
	}

	member := GroupMember{
		UserID:  userID,
		GroupID: group.ID,
	}
	result = db.Create(&member)
	if result.Error != nil {
		return nil, result.Error
	}
	return &group, nil
}

func CheckGroupInUse(db *gorm.DB, groupID uint) (bool, error) {
	var count int64
	err := db.Model(&GroupMember{}).
		Where("group_id", groupID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// role
func GetRoleByID(db *gorm.DB, roleID uint) (*Role, error) {
	return GetByID[Role](db, roleID)
}

func GetRoleByName(db *gorm.DB, name string) (*Role, error) {
	return Get(db, &Role{Name: name})
}

func CreateRoleWithPermissions(db *gorm.DB, name string, permissions []*Permission) (*Role, error) {
	role := Role{
		Name:        name,
		Permissions: permissions,
	}
	result := db.Create(&role)
	if result.Error != nil {
		return nil, result.Error
	}
	return &role, nil
}

func GetRolesByUser(db *gorm.DB, userID uint) ([]*Role, error) {
	var user User
	result := db.Model(&User{}).Preload("Roles").Take(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return user.Roles, nil
}

func AddRoleForUser(db *gorm.DB, userId uint, roleId uint) error {
	userRole := UserRole{
		UserID: userId,
		RoleID: roleId,
	}
	return db.Model(&userRole).Create(userRole).Error
}

func CheckRoleInUse(db *gorm.DB, roleID uint) (bool, error) {
	var count int64
	result := db.Model(&UserRole{}).Where("role_id", roleID).Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	if count > 0 {
		return true, nil
	}

	result = db.Model(&RolePermission{}).Where("role_id", roleID).Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

// permission
func GetPermissionByID(db *gorm.DB, permissionID uint) (*Permission, error) {
	return GetByID[Permission](db, permissionID)
}

func GetPermissionByCode(db *gorm.DB, code string) (*Permission, error) {
	return Get(db, &Permission{Code: code})
}

func GetPermissionsByRole(db *gorm.DB, roleID uint) ([]*Permission, error) {
	var role Role
	result := db.Model(&Role{}).Preload("Permissions").Take(&role)
	if result.Error != nil {
		return nil, result.Error
	}
	return role.Permissions, nil
}

func AddPermissionForRole(db *gorm.DB, roleID uint, name, code string, policies ...string) (*Permission, error) {
	p := Permission{
		Name: name,
		Code: code,
	}

	switch len(policies) {
	case 1:
		p.P1 = policies[0]
	case 2:
		p.P1 = policies[0]
		p.P2 = policies[1]
	case 3:
		p.P1 = policies[0]
		p.P2 = policies[1]
		p.P3 = policies[2]
	default:
	}

	result := db.Create(&p)
	if result.Error != nil {
		return nil, result.Error
	}

	rolePermission := RolePermission{
		RoleID:       roleID,
		PermissionID: p.ID,
	}
	result = db.Create(&rolePermission)
	if result.Error != nil {
		return nil, result.Error
	}
	return &p, nil
}

func DeletePermissionForRole(db *gorm.DB, roleID uint, permissionCode string) error {
	p, err := GetPermissionByCode(db, permissionCode)
	if err != nil {
		return err
	}
	result := db.Model(&RolePermission{}).
		Where("role_id", roleID).
		Where("permission_id", p.ID).
		Delete(&RolePermission{})

	return result.Error
}

func CheckPermissionInUse(db *gorm.DB, permissionCode string) (bool, error) {
	var count int64
	err := db.Model(&RolePermission{}).
		Where("permission_id", permissionCode).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// user
func GetUsersByGroup(db *gorm.DB, groupID uint) ([]*User, error) {
	var group Group
	result := db.Model(&Group{}).Preload("Users").Take(&group)
	if result.Error != nil {
		return nil, result.Error
	}
	return group.Users, nil
}

func GetUsersByRole(db *gorm.DB, roleID uint) ([]*User, error) {
	var role Role
	result := db.Model(&Role{}).Preload("Users").Take(&role)
	if result.Error != nil {
		return nil, result.Error
	}
	return role.Users, nil
}

// check
func CheckRolePermission(db *gorm.DB, roleID uint, code string, policies ...string) (bool, error) {
	ps, err := GetPermissionsByRole(db, roleID)
	if err != nil {
		return false, err
	}
	for _, p := range ps {
		if p.Code != code {
			continue
		}
		switch len(policies) {
		case 1:
			if p.P1 == policies[0] {
				return true, nil
			}
		case 2:
			if p.P1 == policies[0] && p.P2 == policies[1] {
				return true, nil
			}
		case 3:
			if p.P1 == policies[0] && p.P2 == policies[1] && p.P3 == policies[2] {
				return true, nil
			}
		default:
			if p.P1 == "" && p.P2 == "" && p.P3 == "" {
				return true, nil
			}
		}
	}
	return false, nil
}

// TODO: optimize
func CheckUserPermission(db *gorm.DB, userID uint, code string, policies ...string) (bool, error) {
	rs, err := GetRolesByUser(db, userID)
	if err != nil {
		return false, nil
	}

	for _, r := range rs {
		ps, err := GetPermissionsByRole(db, r.ID)
		if err != nil {
			return false, err
		}
		for _, p := range ps {
			if p.Code != code {
				continue
			}
			switch len(policies) {
			case 1:
				if p.P1 == policies[0] {
					return true, nil
				}
			case 2:
				if p.P1 == policies[0] && p.P2 == policies[1] {
					return true, nil
				}
			case 3:
				if p.P1 == policies[0] && p.P2 == policies[1] && p.P3 == policies[2] {
					return true, nil
				}
			default:
				if p.P1 == "" && p.P2 == "" && p.P3 == "" {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

// gin handler
func CurrentGroup(c *gin.Context) *Group {
	if cachedObj, exists := c.Get(GroupField); exists && cachedObj != nil {
		return cachedObj.(*Group)
	}

	session := sessions.Default(c)
	groupId := session.Get(GroupField)
	if groupId == nil {
		return nil
	}

	db := c.MustGet(DbField).(*gorm.DB)
	group, err := GetGroupByID(db, groupId.(uint))
	if err != nil {
		return nil
	}
	c.Set(GroupField, group)
	return group
}

func SwitchGroup(c *gin.Context, group *Group) {
	session := sessions.Default(c)
	session.Set(GroupField, group.ID)
	session.Save()
}
