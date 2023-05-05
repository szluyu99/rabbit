package rabbit

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// const (
// 	PermissionAll    = "all"
// 	PermissionCreate = "create"
// 	PermissionUpdate = "update"
// 	PermissionRead   = "read"
// 	PermissionDelete = "delete"
// )

// const (
// 	GroupRoleAdmin  = "admin"
// 	GroupRoleMember = "member"
// )

type Config struct {
	ID    uint   `json:"id" gorm:"primaryKey"`
	Key   string `json:"key" gorm:"size:128,uniqueIndex"`
	Value string `json:"value"`
	Desc  string `json:"desc" gorm:"size: 200"`
}

type Profile struct {
	Avatar  string         `json:"avatar,omitempty"`
	Gender  string         `json:"gender,omitempty"`
	City    string         `json:"city,omitempty"`
	Region  string         `json:"region,omitempty"`
	Country string         `json:"country,omitempty"`
	Extra   map[string]any `json:"extra,omitempty"`
}

func (p *Profile) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	return json.Unmarshal(value.([]byte), p)
}

func (p Profile) Value() (driver.Value, error) {
	return json.Marshal(p)
}

type User struct {
	ID        uint      `json:"-" gorm:"primarykey"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`

	Email       string     `json:"email" gorm:"size:128;uniqueIndex"`
	Password    string     `json:"-" gorm:"size:128"`
	Phone       string     `json:"phone,omitempty" gorm:"size:64;index"`
	FirstName   string     `json:"firstName,omitempty" gorm:"size:128"`
	LastName    string     `json:"lastName,omitempty" gorm:"size:128"`
	DisplayName string     `json:"displayName,omitempty" gorm:"size:128"`
	IsSuperUser bool       `json:"-"`
	IsStaff     bool       `json:"-"`
	Enabled     bool       `json:"-"`
	Activated   bool       `json:"-"`
	LastLogin   *time.Time `json:"lastLogin,omitempty"`
	LastLoginIP string     `json:"-" gorm:"size:128"`

	Source    string   `json:"-" gorm:"size:64;index"`
	Locale    string   `json:"locale,omitempty" gorm:"size:20"`
	Timezone  string   `json:"timezone,omitempty" gorm:"size:200"`
	Profile   *Profile `json:"profile,omitempty"`
	AuthToken string   `json:"token,omitempty" gorm:"-"`

	Groups []*Group `json:"groups" gorm:"many2many:group_members;"`
	Roles  []*Role  `json:"roles" gorm:"many2many:user_roles;"`
}

type Group struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Name  string `json:"name" gorm:"size:200;uniqueIndex"`
	Extra string `json:"extra"`

	Users []*User `json:"users" gorm:"many2many:group_members;"`
}

type Role struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Name  string `json:"name" gorm:"size:50;uniqueIndex"`
	Label string `json:"label" gorm:"size:200;uniqueIndex"`

	Users       []*User       `json:"users" gorm:"many2many:user_roles;"`
	Permissions []*Permission `json:"permissions" gorm:"many2many:role_permissions;"`
}

type Permission struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Name string `json:"name" gorm:"uniqueIndex"`
	Code string `json:"code" gorm:"size:200;uniqueIndex"`
	P1   string `json:"p1" gorm:"size:200"`
	P2   string `json:"p2" gorm:"size:200"`
	P3   string `json:"p3" gorm:"size:200"`

	Groups []*Group `json:"groups" gorm:"many2many:group_permissions;"`
	Roles  []*Role  `json:"roles" gorm:"many2many:role_permissions;"`
}

type UserRole struct {
	UserID uint `json:"-" gorm:"primarykey"`
	RoleID uint `json:"-" gorm:"primarykey"`

	User User `json:"user"`
	Role Role `json:"role"`
}

type GroupMember struct {
	UserID  uint `json:"-" gorm:"primarykey"`
	GroupID uint `json:"-" gorm:"primarykey"`

	User  User  `json:"user"`
	Group Group `json:"group"`
}

type RolePermission struct {
	RoleID       uint `json:"-" gorm:"primarykey"`
	PermissionID uint `json:"-" gorm:"primarykey"`

	Role       Role       `json:"role"`
	Permission Permission `json:"permission"`
}

func (u *User) GetVisibleName() string {
	if u.DisplayName != "" {
		return u.DisplayName
	}
	if u.FirstName != "" {
		return u.FirstName
	}
	return u.LastName
}

func (u *User) GetProfile() Profile {
	if u.Profile != nil {
		return *u.Profile
	}
	return Profile{}
}

func InitMigrate(db *gorm.DB) error {
	if err := db.SetupJoinTable(&User{}, "Roles", &UserRole{}); err != nil {
		return err
	}

	if err := db.SetupJoinTable(&User{}, "Groups", &GroupMember{}); err != nil {
		return err
	}

	if err := db.SetupJoinTable(&Permission{}, "Roles", &RolePermission{}); err != nil {
		return err
	}

	return db.AutoMigrate(
		&Config{},
		&User{},
		&Group{},
		&Role{},
		&UserRole{},
		&GroupMember{},
		&RolePermission{},
	)
}
