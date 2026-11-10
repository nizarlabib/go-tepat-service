package entity

import (
	"go-starter-service/internal/value_object/v1/price"
	utilutils "go-starter-service/internal/value_object/v1/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Product represents a product.
type Product struct {
	BaseUUIDModel

	Name        string      `json:"name" gorm:"not null;type:varchar(255);"`
	Slug        string      `json:"slug" gorm:"not null;type:varchar(255);index;"`
	Brand       *string     `json:"brand" gorm:"type:varchar(100);"`
	CategoryID  uuid.UUID   `json:"category_id" gorm:"type:uuid;not null;index;"`
	Category    *Category   `json:"category,omitempty" gorm:"foreignKey:CategoryID;references:ID"`
	Description *string     `json:"description" gorm:"type:text;"`
	ImageURL    *string     `json:"image_url" gorm:"type:varchar(500);"`
	ImageAlt    *string     `json:"image_alt" gorm:"type:varchar(255);"`
	Price       price.Price `json:"price" gorm:"type:numeric(15,2);"`
	IsActive    bool        `json:"is_active" gorm:"default:true;index"`

	CategorySlug string `json:"category_slug" gorm:"-"`
}

func (p *Product) BeforeSave(tx *gorm.DB) error {
	if p.Name != "" && p.Slug == "" {
		p.Slug = utilutils.Slugify(p.Name)
	}
	return nil
}
