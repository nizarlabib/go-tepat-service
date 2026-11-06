package entity

type Permission struct {
	BaseUUIDModel

	Name     string `json:"name" gorm:"not null;"`
	BitIndex int    `json:"bit_index" gorm:"not null;uniqueIndex:unq_idx_permissions;"`
}

type PermissionGroup struct {
	Group       string             `json:"group"`
	Type        string             `json:"type,omitempty"`        // hanya untuk leaf
	Permissions []*PermissionValue `json:"permissions,omitempty"` // hanya untuk leaf
	Children    []*PermissionGroup `json:"children,omitempty"`    // hanya untuk parent
}

type PermissionValue struct {
	Label string `json:"label"`
	Value string `json:"value"`
}
