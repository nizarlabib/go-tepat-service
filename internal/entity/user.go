package entity

import (
	// "context"
	// "go-memolab-service/internal/value_object/v1/handling"
	"go-memolab-service/internal/pkg/rbac/permission"
	permissionindex "go-memolab-service/internal/pkg/rbac/permission-index"

	// "slices"
	"time"

	"github.com/google/uuid"
)

type (
	UserAccountStatus string
)

const (
	UserAccountStatusPending, UserAccountStatusSuspended, UserAccountStatusVerified, UserAccountStatusRejected, UserAccountStatusDeleted UserAccountStatus = "pending", "suspended", "verified", "rejected", "deleted"
)

type EncryptedString string

type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
)

type User struct {
	modelMemo
	BaseModel
	Fullname          string  `json:"fullname" gorm:"not null;"`
	Email             string  `json:"email" gorm:"not null;type:varchar(50);unique"`
	Phone             string  `json:"phone"`
	Gender            Gender  `json:"gender" gorm:"type:gender;default:male;"`
	Password          string  `json:"-"  gorm:"type:varchar(100);"`
	ProfilePictureURL *string `json:"profile_picture_url" gorm:"type:varchar(100);"`

	AccountStatus UserAccountStatus `json:"account_status" gorm:"type:user_account_status;default:pending;"`
	RememberToken *string           `json:"-"`

	SuspendedAt     *time.Time `json:"suspended_at,omitempty"`
	SuspendedReason *string    `json:"suspended_reason,omitempty"`

	RoleID      *uuid.UUID `json:"role_id" gorm:"type:uuid;validate:required"`
	Role        *Role      `json:"role" gorm:"foreignKey:RoleID;references:ID"`
	Permissions string     `json:"permission_mask" gorm:"-"`

	VerifiedAt *time.Time `json:"verified_at,omitempty"`

	LastLoggedInAt        *time.Time `json:"-" gorm:"default:null"`
	LastLoggedInIP        *string    `json:"-" gorm:"default:null"`
	LastLoggedInUserAgent *string    `json:"-" gorm:"default:null"`
	LastLoggedInRoleID    *uuid.UUID `json:"-" gorm:"type:uuid;default:null;"`

	PasswordChangedAt  *time.Time `json:"password_changed_at" gorm:"default:null"`
	PasswordMustChange bool       `json:"password_must_change" gorm:"default:false"`

	DiskQuota int64 `json:"disk_quota" gorm:"default:314572800;"`

	AccountDeletionReason *string `json:"account_deletion_reason,omitempty"`
	DeletedBy             *uint   `json:"deleted_by,omitempty"`

	permissions uint                `json:"-" gorm:"-"`
	bigPerms    *permission.BigPerm `json:"-" gorm:"-"`

	UserDevice *DeviceLog `json:"-" gorm:"-"`
}

func (u *User) IsVerified() bool {
	return u.AccountStatus == "verified"
}

func NewUserFromID(id uint) *User {
	return &User{BaseModel: BaseModel{
		ID: id,
	}}
}

func (u *User) HasSuperAdminRole() bool {
	return u.Can(permissionindex.SuperAdmin)
}

func (u *User) Can(bitIndices ...int) bool {
	return u.GetPermissions().HasAll(bitIndices...)
}

func (u *User) GetPermissions() *permission.BigPerm {
	if u.bigPerms != nil {
		return u.bigPerms
	}

	var perms *permission.BigPerm
	if u.Permissions != "" {
		perms = permission.NewFromStr(u.Permissions)
		u.bigPerms = perms
		return u.bigPerms
	}

	perms = permission.New()
	if u.RoleID == nil {
		return perms
	}

	u.bigPerms = perms
	return u.bigPerms
}

func (u *User) MaybeCan(bitIndices ...int) bool {
	return u.GetPermissions().HasOneOf(bitIndices...)
}

func (u *User) IsUser() bool {
	return u.Can(permissionindex.User)
}

func (u *User) IsSuperAdmin() bool {
	return u.Can(permissionindex.SuperAdmin)
}

func (u *User) IsSuspended() bool {
	return u.AccountStatus == UserAccountStatusSuspended && u.SuspendedAt != nil
}

func (u *User) IsAccountVerified() bool {
	return u.AccountStatus == UserAccountStatusVerified
}

// func (u *User) GetRole() string {
// 	if u.Can(permission.Maker) {
// 		return UserRoleMaker
// 	}
// 	if u.Can(permission.Checker) {
// 		return UserRoleChecker
// 	}
// 	if u.Can(permission.Approver) {
// 		return UserRoleApprover
// 	}
// 	return UserRoleMaker
// }

// func (u *User) HasAccessToInstitute(userType string, userRole string, instituteIDs ...string) bool {
// 	if len(u.Roles) == 0 {
// 		return false
// 	}

// 	needPerms := permission.New()
// 	switch userType {
// 	case UserTypeProvider:
// 		needPerms.Grant(permission.Provider)
// 	case UserTypeUser:
// 		needPerms.Grant(permission.User)
// 	}

// 	switch userRole {
// 	case UserRoleMaker:
// 		needPerms.Grant(permission.Maker)
// 	case UserRoleChecker:
// 		needPerms.Grant(permission.Checker)
// 	case UserRoleApprover:
// 		needPerms.Grant(permission.Approver)
// 	}

// 	for _, role := range u.Roles {
// 		rolePerms := role.GetPermissions()
// 		roleHasAccess := rolePerms.HasAllPerms(needPerms)
// 		roleSameInstitute := slices.Contains(instituteIDs, role.InstituteID)

// 		if roleHasAccess && roleSameInstitute {
// 			return true
// 		}
// 	}

// 	return false
// }

// func (u *User) FetchUnits(ctx context.Context) error {
// 	return nil
// }

type UsersIDs struct {
	Ids []uint `json:"ids"`
}

type InstituteFindUnits struct {
	Name     string
	MaxLevel int
	ParentID string
}

type UserFindItemAPI struct {
	*User
	Roles     any `json:"roles" gorm:"type:jsonb;serializer:json;"`
	Institute any `json:"institute" gorm:"type:jsonb;serializer:json;"`
}

type UserWithRoles struct {
	*User
	Roles []*Role `json:"roles"`
}

type UserFindAllOptions struct {
	CountOnly bool
	Name      string
	Email     string
	Phone     string
	Status    string
	RoleID    uuid.UUID
	HasPerms  string
	OrderBy   string
	IsDeleted bool
}
