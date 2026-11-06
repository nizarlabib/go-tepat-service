package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type modelMemo struct {
	memo
}

func (m *modelMemo) HasOriginal() bool {
	if !m.HasMemo("original") {
		return false
	}

	return m.GetMemo("original") != nil
}

func (m *modelMemo) GetOriginal() any {
	if !m.HasOriginal() {
		return nil
	}
	return m.GetMemo("original")
}

func (m *modelMemo) SetOriginal(original any) {
	if original == nil {
		m.RemoveMemo("original")
		return
	}
	m.SetMemo("original", original)
}

type BaseModelTimestamp struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type BaseModel struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

type BaseDeletableModel struct {
	ID        string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type BaseUUIDModel struct {
	ID uuid.UUID `json:"id" gorm:"primarykey;default:gen_random_uuid();"`
	*BaseModelTimestamp
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

type BaseUUIDDeletableModel struct {
	ID uuid.UUID `json:"id" gorm:"primarykey;default:gen_random_uuid();"`
	*BaseModelTimestamp
}
