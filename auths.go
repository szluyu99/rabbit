package rabbit

import (
	"errors"

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
	result := db.Where("id", userID).Preload("Groups").Take(&user)
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
	if err := db.Model(&GroupMember{}).
		Where("group_id", groupID).
		Count(&count).Error; err != nil {
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

func GetRolesByUser(db *gorm.DB, userID uint) ([]*Role, error) {
	var user User
	result := db.Model(&User{}).Preload("Roles").Take(&user, userID)
	if result.Error != nil {
		return nil, result.Error
	}
	return user.Roles, nil
}

func CheckRoleInUse(db *gorm.DB, roleID uint) (bool, error) {
	var count int64
	result := db.Model(&UserRole{}).Where("role_id", roleID).Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

func CheckRoleNameExist(db *gorm.DB, name string) (bool, error) {
	var count int64
	result := db.Model(&Role{}).Where("name", name).Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

func CreateRole(db *gorm.DB, name, label string) (*Role, error) {
	return AddRoleWithPermissions(db, name, label, nil)
}

func AddRoleWithPermissions(db *gorm.DB, name, label string, ps []uint) (*Role, error) {
	role := Role{
		Name:  name,
		Label: label,
	}
	result := db.Create(&role)
	if result.Error != nil {
		return nil, result.Error
	}

	// add new permissions related to this role
	for _, pid := range ps {
		rolePermission := RolePermission{
			RoleID:       role.ID,
			PermissionID: pid,
		}
		result := db.Create(&rolePermission)
		if result.Error != nil {
			return nil, result.Error
		}
	}

	return &role, nil
}

func UpdateRoleWithPermissions(db *gorm.DB, roleID uint, name, label string, ps []uint) (*Role, error) {
	role := Role{
		ID:    roleID,
		Name:  name,
		Label: label,
	}

	// update role, need to clear old permissions related to this role
	result := db.Model(&role).Select("name", "label").Updates(role)
	if result.Error != nil {
		return nil, result.Error
	}
	result = db.Delete(&RolePermission{}, "role_id", role.ID)
	if result.Error != nil {
		return nil, result.Error
	}

	// add new permissions related to this role
	for _, pid := range ps {
		rolePermission := RolePermission{
			RoleID:       role.ID,
			PermissionID: pid,
		}
		result := db.Create(&rolePermission)
		if result.Error != nil {
			return nil, result.Error
		}
	}

	return &role, nil
}

func DeleteRole(db *gorm.DB, roleID uint) error {
	result := db.Delete(&RolePermission{}, "role_id", roleID)
	if result.Error != nil {
		return result.Error
	}
	result = db.Delete(&Role{}, "id", roleID)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// permission
func SavePermission(db *gorm.DB, id, pID uint, name string, anonymous bool, policies ...string) (*Permission, error) {
	permission := Permission{
		ID:        id,
		Name:      name,
		ParentID:  pID,
		Anonymous: anonymous,
	}

	switch len(policies) {
	case 0:
	case 1:
		permission.P1 = policies[0]
	case 2:
		permission.P1 = policies[0]
		permission.P2 = policies[1]
	case 3:
		permission.P1 = policies[0]
		permission.P2 = policies[1]
		permission.P3 = policies[2]
	default:
		return nil, errors.New("invalid policies")
	}

	result := db.Save(&permission)
	if result.Error != nil {
		return nil, result.Error
	}
	return &permission, nil
}

func GetPermissionByID(db *gorm.DB, permissionID uint) (*Permission, error) {
	return GetByID[Permission](db, permissionID)
}

func GetPermissionByName(db *gorm.DB, code string) (*Permission, error) {
	return Get(db, &Permission{Name: code})
}

func GetPermissionsByRole(db *gorm.DB, roleID uint) ([]*Permission, error) {
	var role Role
	result := db.Model(&Role{}).Preload("Permissions").Take(&role, roleID)
	if result.Error != nil {
		return nil, result.Error
	}
	return role.Permissions, nil
}

func GetPermissionChildren(db *gorm.DB, permissionID uint) ([]*Permission, error) {
	var permissions []*Permission
	result := db.Model(&Permission{}).Where("parent_id", permissionID).Find(&permissions)
	if result.Error != nil {
		return nil, result.Error
	}
	return permissions, nil
}

func CheckPermissionInUse(db *gorm.DB, permissionID uint) (bool, error) {
	var count int64
	result := db.Model(&RolePermission{}).Where("permission_id", permissionID).Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

func CheckPermissionNameExist(db *gorm.DB, name string) (bool, error) {
	var count int64
	result := db.Model(&Permission{}).Where("name", name).Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

// if the permission is a parent permission, delete all its children
func DeletePermission(db *gorm.DB, permissionID uint) error {
	p, err := GetPermissionByID(db, permissionID)
	if err != nil {
		return err
	}

	if p.ParentID == 0 {
		children, err := GetPermissionChildren(db, permissionID)
		if err != nil {
			return err
		}
		ids := []uint{permissionID}
		for _, child := range children {
			ids = append(ids, child.ID)
		}
		result := db.Delete(&Permission{}, "id", ids)
		if result.Error != nil {
			return result.Error
		}
	} else {
		result := db.Delete(&Permission{}, "id", permissionID)
		if result.Error != nil {
			return result.Error
		}
	}

	return nil
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

func AddRoleForUser(db *gorm.DB, userId uint, roleId uint) error {
	userRole := UserRole{
		UserID: userId,
		RoleID: roleId,
	}
	return db.Model(&userRole).Create(userRole).Error
}

func UpdateRolesForUser(db *gorm.DB, userId uint, roleIDs []uint) (*User, error) {
	user := User{
		ID: userId,
	}

	result := db.Delete(&UserRole{}, "user_id", user.ID)
	if result.Error != nil {
		return nil, result.Error
	}

	for _, roleID := range roleIDs {
		if err := AddRoleForUser(db, user.ID, roleID); err != nil {
			return nil, err
		}
	}

	return &user, nil
}

// check
func CheckRolePermission(db *gorm.DB, roleID uint, policies ...string) (bool, error) {
	ps, err := GetPermissionsByRole(db, roleID)
	if err != nil {
		return false, err
	}

	for _, p := range ps {
		if p.Anonymous {
			return true, nil
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

func CheckUserPermission(db *gorm.DB, userID uint, policies ...string) (bool, error) {
	rs, err := GetRolesByUser(db, userID)
	if err != nil {
		return false, err
	}

	for _, r := range rs {
		if pass, _ := CheckRolePermission(db, r.ID, policies...); pass {
			return true, nil
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

// for test
func CreateRoleWithPermissions(db *gorm.DB, name, label string, permissions []*Permission) (*Role, error) {
	role := Role{
		Name:        name,
		Label:       label,
		Permissions: permissions,
	}
	result := db.Create(&role)
	if result.Error != nil {
		return nil, result.Error
	}
	return &role, nil
}
