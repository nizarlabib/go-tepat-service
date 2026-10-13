package main

import (
	"go-memolab-service/internal/entity"
	"go-memolab-service/internal/pkg/dd"
	permissionindex "go-memolab-service/internal/pkg/rbac/permission-index"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func seedPermissions(db *gorm.DB) []*entity.Permission {
	logger.Info("Seeding permissions")

	permissions := []*entity.Permission{
		{Name: "Super Admin", BitIndex: permissionindex.SuperAdmin},

		{Name: "Read User", BitIndex: permissionindex.ReadUser},
		{Name: "Create User", BitIndex: permissionindex.CreateUser},
		{Name: "Update User", BitIndex: permissionindex.UpdateUser},
		{Name: "Delete User", BitIndex: permissionindex.DeleteUser},
	}

	err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "bit_index"}},
		DoUpdates: clause.AssignmentColumns([]string{"name"}),
	}).Create(&permissions).Error
	if err != nil {
		dd.PrettyPrint(err)
	}

	logger.Info("Permissions seeded successfully", "count", len(permissions))

	return permissions
}
