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

	u, err := CreateUser(db, "test@example.com", "123456")
	assert.Nil(t, err)

	role, err := CreateRoleWithPermissions(db, "test", []*Permission{
		{Name: "read", Code: "read"},
	})
	assert.Nil(t, err)

	r, err := GetRoleByID(db, role.ID)
	assert.Nil(t, err)
	assert.Equal(t, "test", r.Name)

	r, err = GetRoleByName(db, role.Name)
	assert.Nil(t, err)
	assert.Equal(t, role.ID, r.ID)

	err = AddRoleForUser(db, u.ID, role.ID)
	assert.Nil(t, err)

	rs, err := GetRolesByUser(db, u.ID)
	assert.Nil(t, err)
	assert.Len(t, rs, 1)

	flag, err := CheckRoleInUse(db, role.ID)
	assert.Nil(t, err)
	assert.True(t, flag)
}

func TestPermissions(t *testing.T) {
	db := initDB(t)

	role, err := CreateRoleWithPermissions(db, "test", []*Permission{
		{Name: "read", Code: "read"},
	})
	assert.Nil(t, err)

	p, err := AddPermissionForRole(db, role.ID, "write", "write")
	assert.Nil(t, err)
	assert.NotNil(t, p)

	p1, err := GetPermissionByID(db, p.ID)
	assert.Nil(t, err)
	assert.Equal(t, p.Name, p1.Name)

	p2, err := GetPermissionByCode(db, p.Code)
	assert.Nil(t, err)
	assert.Equal(t, p.Name, p2.Name)

	ps, err := GetPermissionsByRole(db, role.ID)
	assert.Nil(t, err)
	assert.Len(t, ps, 2)

	err = DeletePermissionForRole(db, role.ID, p.Code)
	assert.Nil(t, err)

	ps, err = GetPermissionsByRole(db, role.ID)
	assert.Nil(t, err)
	assert.Len(t, ps, 1)

	flag, err := CheckPermissionInUse(db, p.Code)
	assert.Nil(t, err)
	assert.False(t, flag)
}
