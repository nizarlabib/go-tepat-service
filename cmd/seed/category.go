package main

import (
	"go-starter-service/internal/entity"
	"go-starter-service/internal/pkg/dd"
	"go-starter-service/internal/value_object/v1/convert"

	"gorm.io/gorm"
)

func seedCategories(db *gorm.DB) {
	logger.Info("Seeding categories")

	categories := []*entity.Category{
		{Name: "Mata Bor", Slug: "mata-bor", Description: convert.PointerStr("Kategori mata bor"), IsActive: true, SortOrder: 1},
		{Name: "Hand Tap", Slug: "hand-tap", Description: convert.PointerStr("Kategori hand tap"), IsActive: true, SortOrder: 2},
		{Name: "Insert Widia", Slug: "insert-widia", Description: convert.PointerStr("Kategori insert widia"), IsActive: true, SortOrder: 3},
		{Name: "Aksesoris Machining", Slug: "aksesoris-machining", Description: convert.PointerStr("Kategori aksesoris machining"), IsActive: true, SortOrder: 4},
	}

	for _, cat := range categories {
		var existing entity.Category
		result := db.Where("slug = ? AND deleted_at IS NULL", cat.Slug).First(&existing)
		if result.Error == nil {
			logger.Info("category already exists, skipping", "slug", cat.Slug)
			continue
		}

		result = db.Create(cat)
		if result.Error != nil {
			dd.PrettyPrint(result.Error)
			continue
		}
	}

	logger.Info("Categories seeded successfully", "count", len(categories))
}
