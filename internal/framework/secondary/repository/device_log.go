package repository

import (
	"context"
	"errors"
	"net/http"

	"go-memolab-service/internal/entity"
	"go-memolab-service/internal/infrastructure"
	"go-memolab-service/internal/value_object/v1/constant"
	"go-memolab-service/internal/value_object/v1/handling"

	"gorm.io/gorm"
)

type DeviceLogs interface {
	Create(ctx context.Context, e *entity.DeviceLog) error
	Delete(ctx context.Context, e *entity.DeviceLog) error
	FindAllByUserID(ctx context.Context, uid uint) ([]*entity.DeviceLog, error)
	FindByID(ctx context.Context, e *entity.DeviceLog) error
	Update(ctx context.Context, e *entity.DeviceLog) error
}

type deviceLog struct {
	db *gorm.DB
}

func NewdeviceLog(infra infrastructure.Postgres) DeviceLogs {
	return deviceLog{db: infra.GetClientGorm()}
}

func (r deviceLog) Create(ctx context.Context, e *entity.DeviceLog) error {
	var err error
	var temp entity.DeviceLog
	_ = r.db.WithContext(ctx).Where("user_id = ? AND user_agent = ? AND ip_address = ?", e.UserID, e.UserAgent, e.IpAddress).First(&temp).Error
	if temp.ID != 0 {
		if err = r.db.WithContext(ctx).Where("id = ?", temp.ID).Updates(e).Error; err != nil {
			return handling.NewHttpError500(err)
		}
	} else {
		if err = r.db.WithContext(ctx).Create(e).Error; err != nil {
			return handling.NewHttpError500(err)
		}
	}

	return nil
}

func (r deviceLog) Delete(ctx context.Context, e *entity.DeviceLog) error {
	err := r.db.WithContext(ctx).Unscoped().Where("session_id = ? AND user_id = ?", e.SessionID, e.UserID).First(e).Delete(e).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return handling.NewHttpError(err, http.StatusNotFound, "Sesi tidak ditemukan.", constant.ERR_RECORD_NOT_FOUND)
		}
		return handling.NewHttpError500(err)
	}
	return nil
}

func (r deviceLog) FindAllByUserID(ctx context.Context, uid uint) ([]*entity.DeviceLog, error) {
	var devices []*entity.DeviceLog
	if err := r.db.WithContext(ctx).Where("user_id = ?", uid).Order("last_activity_at DESC").Limit(10).Find(&devices).Error; err != nil {
		return nil, handling.NewHttpError500(err)
	}
	return devices, nil
}

func (r deviceLog) FindByID(ctx context.Context, e *entity.DeviceLog) error {
	if err := r.db.WithContext(ctx).Where("session_id = ?", e.SessionID).First(e).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return handling.NewHttpError(err, http.StatusNotFound, "Sesi tidak ditemukan.", constant.ERR_RECORD_NOT_FOUND)
		}
		return handling.NewHttpError500(err)
	}
	return nil
}

func (r deviceLog) Update(ctx context.Context, e *entity.DeviceLog) error {
	if err := r.db.WithContext(ctx).Save(e).Error; err != nil {
		return handling.NewHttpError500(err)
	}
	return nil
}
