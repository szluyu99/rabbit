package rabbit

import (
	"errors"

	"gorm.io/gorm"
)

// group
func GetGroupByID(db *gorm.DB, gid uint) (*Group, error) {
	return GetByID[Group](db, gid)
}

func GetGroupByName(db *gorm.DB, name string) (*Group, error) {
	return Get(db, &Group{Name: name})
}

func GetGroupsByUser(db *gorm.DB, uid uint) ([]*Group, error) {
	var user User
	result := db.Where("id", uid).Preload("Groups").Take(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	return user.Groups, nil
}

func GetFirstGroupByUser(db *gorm.DB, uid uint) (*Group, error) {
	var member GroupMember
	result := db.Where("user_id", uid).Preload("Group").Take(&member)
	if result.Error != nil {
		return nil, result.Error
	}
	return &member.Group, nil
}

func CreateGroupByUser(db *gorm.DB, uid uint, name string) (*Group, error) {
	group := Group{
		Name: name,
	}
	result := db.Create(&group)
	if result.Error != nil {
		return nil, result.Error
	}

	member := GroupMember{
		UserID:  uid,
		GroupID: group.ID,
	}
	result = db.Create(&member)
	if result.Error != nil {
		return nil, result.Error
	}
	return &group, nil
}

func CheckGroupInUse(db *gorm.DB, gid uint) (bool, error) {
	var count int64
	if err := db.Model(&GroupMember{}).
		Where("group_id", gid).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// role
func GetRoleByID(db *gorm.DB, rid uint) (*Role, error) {
	return GetByID[Role](db, rid)
}

func GetRoleByName(db *gorm.DB, name string) (*Role, error) {
	return Get(db, &Role{Name: name})
}

func GetRolesByUser(db *gorm.DB, uid uint) ([]*Role, error) {
	var user User
	result := db.Model(&User{}).Preload("Roles").Take(&user, uid)
	if result.Error != nil {
		return nil, result.Error
	}
	return user.Roles, nil
}

func CheckRoleInUse(db *gorm.DB, rid uint) (bool, error) {
	var count int64
	result := db.Model(&UserRole{}).Where("role_id", rid).Count(&count)
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

func UpdateRoleWithPermissions(db *gorm.DB, rid uint, name, label string, ps []uint) (*Role, error) {
	role := Role{
		ID:    rid,
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

func DeleteRole(db *gorm.DB, rid uint) error {
	result := db.Delete(&RolePermission{}, "role_id", rid)
	if result.Error != nil {
		return result.Error
	}
	result = db.Delete(&Role{}, "id", rid)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// permission
func SavePermission(db *gorm.DB, id, pid uint, name, uri, method string, anonymous bool) (*Permission, error) {
	permission := Permission{
		ID:        id,
		Name:      name,
		ParentID:  pid,
		Uri:       uri,
		Method:    method,
		Anonymous: anonymous,
	}

	result := db.Save(&permission)
	if result.Error != nil {
		return nil, result.Error
	}
	return &permission, nil
}

func GetPermissionByID(db *gorm.DB, pid uint) (*Permission, error) {
	return GetByID[Permission](db, pid)
}

func GetPermissionByName(db *gorm.DB, name string) (*Permission, error) {
	return Get(db, &Permission{Name: name})
}

func GetPermission(db *gorm.DB, uri, method string) (*Permission, error) {
	return Get(db.Debug(), &Permission{Uri: uri, Method: method})
}

func GetPermissionsByRole(db *gorm.DB, rid uint) ([]*Permission, error) {
	var role Role
	result := db.Model(&Role{}).Preload("Permissions").Take(&role, rid)
	if result.Error != nil {
		return nil, result.Error
	}
	return role.Permissions, nil
}

func GetPermissionChildren(db *gorm.DB, pid uint) ([]*Permission, error) {
	var permissions []*Permission
	result := db.Model(&Permission{}).Where("parent_id", pid).Find(&permissions)
	if result.Error != nil {
		return nil, result.Error
	}
	return permissions, nil
}

func CheckPermissionInUse(db *gorm.DB, pid uint) (bool, error) {
	var count int64
	result := db.Model(&RolePermission{}).Where("permission_id", pid).Count(&count)
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
func DeletePermission(db *gorm.DB, pid uint) error {
	p, err := GetPermissionByID(db, pid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	if p.ParentID == 0 {
		children, err := GetPermissionChildren(db, pid)
		if err != nil {
			return err
		}
		ids := []uint{pid}
		for _, child := range children {
			ids = append(ids, child.ID)
		}
		result := db.Delete(&Permission{}, "id", ids)
		if result.Error != nil {
			return result.Error
		}
	} else {
		result := db.Delete(&Permission{}, "id", pid)
		if result.Error != nil {
			return result.Error
		}
	}

	return nil
}

// user
func GetUsersByGroup(db *gorm.DB, gid uint) ([]*User, error) {
	var group Group
	result := db.Model(&Group{}).Preload("Users").Take(&group)
	if result.Error != nil {
		return nil, result.Error
	}
	return group.Users, nil
}

func GetUsersByRole(db *gorm.DB, rid uint) ([]*User, error) {
	var role Role
	result := db.Model(&Role{}).Preload("Users").Take(&role)
	if result.Error != nil {
		return nil, result.Error
	}
	return role.Users, nil
}

func AddRoleForUser(db *gorm.DB, uid uint, rid uint) error {
	userRole := UserRole{
		UserID: uid,
		RoleID: rid,
	}
	return db.Model(&userRole).Create(userRole).Error
}

func UpdateRolesForUser(db *gorm.DB, uid uint, rids []uint) (*User, error) {
	user := User{
		ID: uid,
	}

	result := db.Delete(&UserRole{}, "user_id", user.ID)
	if result.Error != nil {
		return nil, result.Error
	}

	for _, roleID := range rids {
		if err := AddRoleForUser(db, user.ID, roleID); err != nil {
			return nil, err
		}
	}

	return &user, nil
}

// check
func CheckRolePermission(db *gorm.DB, rid uint, uri, method string) (bool, error) {
	ps, err := GetPermissionsByRole(db, rid)
	if err != nil {
		return false, err
	}

	for _, p := range ps {
		if p.Anonymous || (p.Uri == uri && p.Method == method) {
			return true, nil
		}
	}
	return false, nil
}

func CheckUserPermission(db *gorm.DB, uid uint, uri, method string) (bool, error) {
	p, err := GetPermission(db, uri, method)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	if p.Anonymous {
		return true, nil
	}

	rs, err := GetRolesByUser(db, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	for _, r := range rs {
		if pass, _ := CheckRolePermission(db, r.ID, uri, method); pass {
			return true, nil
		}
	}

	return false, nil
}

// for test
func CreateRoleWithPermissions(db *gorm.DB, name, label string, ps []*Permission) (*Role, error) {
	role := Role{
		Name:        name,
		Label:       label,
		Permissions: ps,
	}
	result := db.Create(&role)
	if result.Error != nil {
		return nil, result.Error
	}
	return &role, nil
}
