package models

import (
	"context"
	"eggServer/internal/constant"
	"eggServer/pkg/errors"
	"gorm.io/gorm"
)

type Guide struct {
	ID     uint64  `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	RoleID uint64  `gorm:"column:roleId;index:idx_roleId;NOT NULL"`
	Step   []int32 `gorm:"column:data;serializer:json"`
}

var GuideRepo = new(guideRepo)

type guideRepo struct{}

func (s *guideRepo) Save(ctx context.Context, db *gorm.DB, roleId uint64, guide *Guide) error {
	guide.RoleID = roleId
	if err := db.Save(guide).Error; err != nil {
		return errors.NewResponseError(constant.DatabaseError, err)
	}
	return nil
}

func (s *guideRepo) Get(ctx context.Context, db *gorm.DB, roleId uint64) (*Guide, error) {
	guide := new(Guide)
	err := db.Where("roleId=?", roleId).First(guide).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.NewResponseError(constant.DatabaseError, err)
	}
	return guide, err
}

func (s *guideRepo) Delete(ctx context.Context, db *gorm.DB, roleId uint64) error {
	if err := db.Where("roleId=?", roleId).Delete(new(Guide)).Error; err != nil {
		return errors.NewResponseError(constant.DatabaseError, err)
	}
	return nil
}
