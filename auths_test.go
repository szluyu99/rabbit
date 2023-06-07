package rabbit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func initDB(t *testing.T) *gorm.DB {
	db := InitDatabase("", "", nil)
	err := InitMigrate(db)
	assert.Nil(t, err)

	return db
}

func TestGroups(t *testing.T) {
	db := initDB(t)

	u, err := CreateUser(db, "test@example.com", "123456")
	assert.Nil(t, err)

	gp1, err := CreateGroupByUser(db, u.ID, "group1")
	assert.Nil(t, err)

	gp2, err := CreateGroupByUser(db, u.ID, "group2")
	assert.Nil(t, err)

	gp, err := GetGroupByID(db, gp2.ID)
	assert.Nil(t, err)
	assert.Equal(t, gp2.Name, gp.Name)

	gp, err = GetFirstGroupByUser(db, u.ID)
	assert.Nil(t, err)
	assert.Equal(t, gp1.Name, gp.Name)

	gps, err := GetGroupsByUser(db, u.ID)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(gps))

	users, err := GetUsersByGroup(db, gp1.ID)
	assert.Nil(t, err)
	assert.Len(t, users, 1)

	// non-exist group
	negp, err := GetGroupByID(db, 999)
	assert.Nil(t, negp)
	assert.NotNil(t, err)

	uu, err := CreateUser(db, "test2@example.com", "123456")
	assert.Nil(t, err)

	// user without group
	gp, err = GetFirstGroupByUser(db, uu.ID)
	assert.NotNil(t, err)
	assert.Nil(t, gp)

	gp, err = GetGroupByName(db, "group1")
	assert.Nil(t, err)
	assert.NotNil(t, gp)

	flag, err := CheckGroupInUse(db, gp1.ID)
	assert.Nil(t, err)
	assert.True(t, flag)
}

func TestRoles(t *testing.T) {
	db := initDB(t)

	// prepare data
	user, err := CreateUser(db, "test@qq.com", "123456")
	assert.Nil(t, err)

	role, err := CreateRole(db, "admin", "ADMIN")
	assert.Nil(t, err)

	//
	v, err := GetRoleByID(db, role.ID)
	assert.Nil(t, err)
	assert.Equal(t, role.Name, v.Name)

	v, err = GetRoleByName(db, role.Name)
	assert.Nil(t, err)
	assert.Equal(t, role.Name, v.Name)

	// check role name exist
	{
		flag, err := CheckRoleNameExist(db, "admin")
		assert.Nil(t, err)
		assert.True(t, flag)

		flag, err = CheckRoleNameExist(db, "not-exist")
		assert.Nil(t, err)
		assert.False(t, flag)
	}

	// check role in use
	{
		flag, err := CheckRoleInUse(db, role.ID)
		assert.Nil(t, err)
		assert.False(t, flag)
	}

	// get user roles
	{
		err := AddRoleForUser(db, user.ID, role.ID)
		assert.Nil(t, err)

		roles, err := GetRolesByUser(db, role.ID)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(roles))
	}

	// update role
	{
		p1, err := SavePermission(db, 0, 0, "p1", "/p1", "GET", false)
		assert.Nil(t, err)
		p2, err := SavePermission(db, 0, 0, "p2", "/p2", "POST", false)
		assert.Nil(t, err)

		_, err = UpdateRoleWithPermissions(db, role.ID, role.Name, role.Label, []uint{p1.ID, p2.ID})
		assert.Nil(t, err)

		ps, err := GetPermissionsByRole(db, role.ID)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(ps))
	}

	err = DeleteRole(db, role.ID)
	assert.Nil(t, err)
}

func TestPermissions(t *testing.T) {
	db := initDB(t)

	// prepare data
	// user, err := CreateUser(db, "test@qq.com", "123456")
	// assert.Nil(t, err)

	p1, err := SavePermission(db, 0, 0, "p1", "/p1", "GET", false)
	assert.Nil(t, err)

	//
	v, err := GetPermissionByID(db, p1.ID)
	assert.Nil(t, err)
	assert.Equal(t, p1.Name, v.Name)

	v, err = GetPermissionByName(db, p1.Name)
	assert.Nil(t, err)
	assert.Equal(t, p1.Name, v.Name)

	// check permission name exist
	{
		flag, err := CheckPermissionNameExist(db, "p1")
		assert.Nil(t, err)
		assert.True(t, flag)

		flag, err = CheckPermissionNameExist(db, "not-exist")
		assert.Nil(t, err)
		assert.False(t, flag)
	}

	// check permission in use
	{
		_, err := AddRoleWithPermissions(db, "admin", "ADMIN", []uint{p1.ID})
		assert.Nil(t, err)

		flag, err := CheckPermissionInUse(db, p1.ID)
		assert.Nil(t, err)
		assert.True(t, flag)
	}

	// delete child permissions
	{
		// create children
		p11, _ := SavePermission(db, 0, p1.ID, "p1-1", "/p11", "GET", false)
		p12, _ := SavePermission(db, 0, p1.ID, "p1-2", "/p12", "GET", false)
		p13, _ := SavePermission(db, 0, p1.ID, "p1-3", "/p13", "GET", false)

		children, err := GetPermissionChildren(db, p11.ID)
		assert.Nil(t, err)
		assert.Len(t, children, 0)

		children, err = GetPermissionChildren(db, p1.ID)
		assert.Nil(t, err)
		assert.Len(t, children, 3)

		// delete parent with children
		err = DeletePermission(db, p1.ID)
		assert.Nil(t, err)

		var count int64
		db.Model(&Permission{}).Where("id in (?)", []uint{p11.ID, p12.ID, p13.ID}).Count(&count)
		assert.Equal(t, int64(0), count)
	}

}

func TestCheckPermission(t *testing.T) {
	// not anonymous
	{
		db := initDB(t)
		u, _ := CreateUser(db, "test@example.com", "123456")
		r, _ := CreateRoleWithPermissions(db, "admin", "ADMIN", []*Permission{
			{
				Name:   "p1",
				Uri:    "GET",
				Method: "/api/v1/users",
			},
		})

		AddRoleForUser(db, u.ID, r.ID)

		pass, err := CheckUserPermission(db, u.ID, "GET", "/api/v1/users")
		assert.Equal(t, true, pass)
		assert.Nil(t, err)

		pass, err = CheckUserPermission(db, u.ID, "POST", "/api/v1/users")
		assert.Equal(t, false, pass)
		assert.Nil(t, err)

		pass, err = CheckUserPermission(db, u.ID, "GET", "/api/v2/users")
		assert.Equal(t, false, pass)
		assert.Nil(t, err)
	}

	// anonymous
	{
		db := initDB(t)
		u, _ := CreateUser(db, "test@example.com", "123456")

		SavePermission(db, 0, 0, "p1", "/api/v1/users", "GET", true)
		SavePermission(db, 0, 0, "p2", "/api/v1/users", "POST", true)

		count, _ := Count[Permission](db)
		assert.Equal(t, 2, count)

		pass, err := CheckUserPermission(db, u.ID, "/api/v1/users", "GET")
		assert.Equal(t, true, pass)
		assert.Nil(t, err)

		pass, err = CheckUserPermission(db, u.ID, "/api/v1/users", "POST")
		assert.Equal(t, true, pass)
		assert.Nil(t, err)

		pass, err = CheckUserPermission(db, u.ID, "/api/v2/users", "GET")
		assert.Equal(t, false, pass)
		assert.Nil(t, err)
	}
}
