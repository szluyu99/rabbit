package rabbit

import (
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RoleForm struct {
	ID            uint   `json:"id"`
	Name          string `json:"name"`
	Label         string `json:"label"`
	PermissionIds []uint `json:"permission_ids"`
}

func RegisterAuthorizationHandlers(prefix string, db *gorm.DB, r *gin.Engine) {
	if prefix == "" {
		prefix = GetEnv(ENV_AUTH_PREFIX)
	}

	r.PUT(filepath.Join(prefix, "role"), handleCreateRole)
	r.PATCH(filepath.Join(prefix, "role/:id"), handleUpdateRole)
	r.DELETE(filepath.Join(prefix, "role/:id"), handleDeleteRole)
	r.PUT(filepath.Join(prefix, "permission"), handleAddPermission)
	r.PATCH(filepath.Join(prefix, "permission/:key"), handleEditPermission)
	r.DELETE(filepath.Join(prefix, "permission/:id"), handleDeletePermission)
}

// role
func handleCreateRole(c *gin.Context) {
	var form RoleForm
	if err := c.BindJSON(&form); err != nil {
		HandleError(c, http.StatusBadRequest, err)
		return
	}

	db := c.MustGet(DbField).(*gorm.DB)

	exist, err := CheckRoleNameExist(db, form.Name)
	if exist || err != nil {
		HandleErrorMsg(c, http.StatusBadRequest, "role name exists")
		return
	}

	role, err := AddRoleWithPermissions(db, form.Name, form.Label, form.PermissionIds)

	if err != nil {
		HandleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, role)
}

func handleUpdateRole(c *gin.Context) {
	roleID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		HandleErrorMsg(c, http.StatusBadRequest, "role id invalid")
		return
	}

	var form RoleForm
	if err := c.BindJSON(&form); err != nil {
		HandleError(c, http.StatusBadRequest, err)
		return
	}

	db := c.MustGet(DbField).(*gorm.DB)

	role, err := UpdateRoleWithPermissions(db, uint(roleID), form.Name, form.Label, form.PermissionIds)
	if err != nil {
		HandleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, role)
}

func handleDeleteRole(c *gin.Context) {
	roleID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		HandleErrorMsg(c, http.StatusBadRequest, "role id invalid")
		return
	}

	db := c.MustGet(DbField).(*gorm.DB)

	flag, err := CheckRoleInUse(db, uint(roleID))
	if err != nil || flag {
		HandleErrorMsg(c, http.StatusBadRequest, "role in use")
		return
	}

	if err := DeleteRole(db, uint(roleID)); err != nil {
		HandleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, true)
}

// permission
func handleAddPermission(c *gin.Context) {
	var form Permission
	if err := c.BindJSON(&form); err != nil {
		HandleError(c, http.StatusBadRequest, err)
		return
	}

	db := c.MustGet(DbField).(*gorm.DB)

	p, err := SavePermission(db, 0, form.ParentID, form.Name, form.Anonymous, form.P1, form.P2, form.P3)
	if err != nil {
		HandleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, p)
}

func handleDeletePermission(c *gin.Context) {
	pID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		HandleErrorMsg(c, http.StatusBadRequest, "permission id invalid")
		return
	}

	db := c.MustGet(DbField).(*gorm.DB)

	flag, err := CheckPermissionInUse(db, uint(pID))
	if err != nil || flag {
		HandleErrorMsg(c, http.StatusBadRequest, "permission in use")
		return
	}

	if err := DeletePermission(db, uint(pID)); err != nil {
		HandleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, true)
}

func handleEditPermission(c *gin.Context) {
	db := c.MustGet(DbField).(*gorm.DB)
	HandleEdit[Permission](c, db, []string{"Name", "Anonymous", "P1", "P2", "P3"}, nil)
}