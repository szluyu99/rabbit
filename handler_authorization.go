package rabbit

import (
	"net/http"
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

func RegisterAuthorizationHandlers(db *gorm.DB, r gin.IRoutes) {
	r.PUT("role", handleCreateRole)
	r.PATCH("role/:key", handleUpdateRole)
	r.DELETE("role/:key", handleDeleteRole)
	r.PUT("permission", handleAddPermission)
	r.PATCH("permission/:key", handleEditPermission)
	r.DELETE("permission/:key", handleDeletePermission)
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
		HandleErrorMessage(c, http.StatusBadRequest, "role name exists")
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
	roleID, err := strconv.Atoi(c.Param("key"))
	if err != nil {
		HandleErrorMessage(c, http.StatusBadRequest, "role id invalid")
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
	roleID, err := strconv.Atoi(c.Param("key"))
	if err != nil {
		HandleErrorMessage(c, http.StatusBadRequest, "role id invalid")
		return
	}

	db := c.MustGet(DbField).(*gorm.DB)

	flag, err := CheckRoleInUse(db, uint(roleID))
	if err != nil || flag {
		HandleErrorMessage(c, http.StatusBadRequest, "role in use")
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

	p, err := SavePermission(db, 0, form.ParentID, form.Name, form.Uri, form.Method, form.Anonymous)
	if err != nil {
		HandleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, p)
}

func handleDeletePermission(c *gin.Context) {
	pID, err := strconv.Atoi(c.Param("key"))
	if err != nil {
		HandleErrorMessage(c, http.StatusBadRequest, "permission id invalid")
		return
	}

	db := c.MustGet(DbField).(*gorm.DB)

	flag, err := CheckPermissionInUse(db, uint(pID))
	if err != nil || flag {
		HandleErrorMessage(c, http.StatusBadRequest, "permission in use")
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
