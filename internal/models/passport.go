package models

import (
	"context"
	"eggServer/internal/constant"
	"eggServer/pkg/errors"
	"gorm.io/gorm"
)

type PassPort struct {
	ID         uint64 `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	RoleID     uint64 `gorm:"column:roleId;index:idx_roleId;NOT NULL"`
	PassPortId int32  `gorm:"column:passPortId;NOT NULL"` // id
	State      byte   `gorm:"column:state;NOT NULL"`      // 0未领取 1普通奖励已领取 2豪华奖励已领取 3全部领取
}

var PassPortRepo = new(passPortRepo)

type passPortRepo struct{}

func (s *passPortRepo) Save(ctx context.Context, db *gorm.DB, passPort *PassPort) error {
	if err := db.Save(passPort).Error; err != nil {
		return errors.NewResponseError(constant.DatabaseError, err)
	}
	return nil
}

func (s *passPortRepo) Get(ctx context.Context, db *gorm.DB, roleId uint64, passPortId int32) (*PassPort, error) {
	passPort := new(PassPort)
	err := db.Where("roleId=? and passPortId=?", roleId, passPortId).First(passPort).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return passPort, errors.NewResponseError(constant.DatabaseError, err)
	}
	return passPort, err
}

func (s *passPortRepo) FindAllByRoleId(ctx context.Context, db *gorm.DB, roleId uint64) ([]*PassPort, error) {
	passPorts := make([]*PassPort, 0)
	err := db.Where("roleId=?", roleId).Find(&passPorts).Error
	return passPorts, err
}
