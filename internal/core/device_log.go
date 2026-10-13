package core

import (
	"context"
	"go-memolab-service/internal/entity"
	"go-memolab-service/internal/framework/secondary/repository"
)

type DeviceLog interface {
	Create(ctx context.Context, e *entity.DeviceLog) error
	Delete(ctx context.Context, e *entity.DeviceLog) error
	FindAllByUserID(ctx context.Context, uid uint) ([]*entity.DeviceLog, error)
	FindByID(ctx context.Context, e *entity.DeviceLog) error
	Update(ctx context.Context, e *entity.DeviceLog) error
}

type deviceLog struct {
	repo repository.DeviceLogs
}

func NewDeviceLog(repo repository.DeviceLogs) DeviceLog {
	return &deviceLog{repo: repo}
}

func (d *deviceLog) Create(ctx context.Context, e *entity.DeviceLog) error {
	return d.repo.Create(ctx, e)
}

func (d *deviceLog) Delete(ctx context.Context, e *entity.DeviceLog) error {
	return d.repo.Delete(ctx, e)
}

func (d *deviceLog) FindAllByUserID(ctx context.Context, uid uint) ([]*entity.DeviceLog, error) {
	return d.repo.FindAllByUserID(ctx, uid)
}

func (d *deviceLog) FindByID(ctx context.Context, e *entity.DeviceLog) error {
	return d.repo.FindByID(ctx, e)
}

func (d *deviceLog) Update(ctx context.Context, e *entity.DeviceLog) error {
	return d.repo.Update(ctx, e)
}
