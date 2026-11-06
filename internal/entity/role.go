package entity

import (
	"go-starter-service/internal/pkg/rbac/permission"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role struct {
	ID       uuid.UUID `json:"id" gorm:"type:uuid;not null;column:uuid;default:gen_random_uuid();unique;"`
	IDLegacy uint      `json:"-" gorm:"column:id;primarykey;"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	Name           string `json:"name" gorm:"not null;"`
	Description    string `json:"description"`
	PermissionMask string `json:"permissions" gorm:"default:0;not null;"`
	IsDefault      bool   `json:"default" gorm:"default:false"`
}

func (r *Role) GetPermissions() *permission.BigPerm {
	return permission.NewFromStr(r.PermissionMask)
}

// SetPermissions is a flexible method to set permissions for the role.
// *permission.BigPerm, uint, uint64, string
func (r *Role) SetPermissions(permissions any) {
	switch v := permissions.(type) {
	case *permission.BigPerm:
		r.PermissionMask = v.Value().String()
	case uint:
		r.PermissionMask = permission.NewFromUint64(uint64(v)).Value().String()
	case uint64:
		r.PermissionMask = permission.NewFromUint64(v).Value().String()
	case string:
		r.PermissionMask = permission.NewFromStr(v).Value().String()
	}
}

func (r *Role) Can(permIndices ...int) bool {
	return r.GetPermissions().HasAll(permIndices...)
}
