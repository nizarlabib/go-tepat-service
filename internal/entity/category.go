package entity

import (
	utilutils "go-starter-service/internal/value_object/v1/utils"

	"gorm.io/gorm"
)

type Category struct {
	BaseUUIDModel

	Name        string     `json:"name" gorm:"not null;type:varchar(100);"`
	Slug        string     `json:"slug" gorm:"not null;type:varchar(100);index;"`
	Description *string    `json:"description" gorm:"type:text;"`
	IsActive    bool       `json:"is_active" gorm:"default:true;"`
	SortOrder   int        `json:"sort_order" gorm:"default:0;"`
	Products    []*Product `json:"products,omitempty" gorm:"foreignKey:CategoryID;references:ID"`
}

func (p *Category) BeforeSave(tx *gorm.DB) error {
	if p.Name != "" && p.Slug == "" {
		p.Slug = utilutils.Slugify(p.Name)
	}
	return nil
}
