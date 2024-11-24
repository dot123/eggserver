package models

import (
	"context"
	"eggServer/internal/constant"
	"eggServer/pkg/errors"
	"gorm.io/gorm"
)

type Role struct {
	ID                   uint64 `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	UserID               uint64 `gorm:"column:userId;index:idx_userId;NOT NULL"`
	CreatedAt            int64  `gorm:"column:createdAt;"`
	LastLogin            int64  `gorm:"column:lastLogin;"`
	LastLayEggTime       int64  `gorm:"column:lastLayEggTime;"`
	LastAddVitTime       int64  `gorm:"column:lastAddVitTime;"`
	LastClickCount       int32  `gorm:"column:lastClickCount;"`
	LastDailyTime        int64  `gorm:"column:lastDailyTime;"`
	LastDeskId           string `gorm:"column:lastDeskId;"`
	AutoEggCollectETime  int64  `gorm:"column:autoEggCollectETime;"`  // 自动收蛋结束时间
	AutoEggCollectRTime  int64  `gorm:"column:autoEggCollectRTime;"`  // 自动收蛋剩余时间
	BattleCount          int32  `gorm:"column:battleCount;"`          // 战斗次数
	PassPortDeluxeReward byte   `gorm:"column:passPortDeluxeReward;"` // 是否有通行证豪华奖励
	GM                   byte   `gorm:"column:gm;"`
}

var RoleRepo = new(roleRepo)

type roleRepo struct{}

func (s *roleRepo) Get(ctx context.Context, db *gorm.DB, roleId uint64) (*Role, error) {
	role := new(Role)
	err := db.Where("`id` = ?", roleId).First(role).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return role, errors.NewResponseError(constant.DatabaseError, err)
	}
	return role, err
}

func (s *roleRepo) UpdateColum(ctx context.Context, db *gorm.DB, roleId uint64, column string, value interface{}) error {
	if err := db.Model(new(Role)).Where("`id` = ?", roleId).Update(column, value).Error; err != nil {
		return errors.NewResponseError(constant.DatabaseError, err)
	}
	return nil
}

func (s *roleRepo) FindOneByUserId(ctx context.Context, db *gorm.DB, userId uint64) (*Role, error) {
	role := new(Role)
	err := db.Where("`userId` = ?", userId).First(role).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return role, errors.NewResponseError(constant.DatabaseError, err)
	}
	return role, err
}

func (s *roleRepo) Updates(ctx context.Context, db *gorm.DB, roleId uint64, values interface{}) error {
	if err := db.Model(new(Role)).Where("`id` = ?", roleId).Updates(values).Error; err != nil {
		return errors.NewResponseError(constant.DatabaseError, err)
	}
	return nil
}
