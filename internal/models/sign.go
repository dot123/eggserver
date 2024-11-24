package models

import (
	"context"
	"eggServer/internal/constant"
	"eggServer/pkg/errors"
	"gorm.io/gorm"
)

type Sign struct {
	ID            uint64 `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	RoleID        uint64 `gorm:"column:roleId;index:idx_roleId;NOT NULL"`
	StartTime     int64  `gorm:"column:startTime;"`      // 签到开始时间
	LoginNum      int    `gorm:"column:loginNum;"`       // 签到登录天数
	LoginTime     int64  `gorm:"column:loginTime;"`      // 签到登录时间
	SignCardETime int64  `gorm:"column:signCardETime;"`  // 豪华签到奖励结束时间
	Day           int    `gorm:"column:day;NOT NULL"`    // 签到天数
	ReSign        int    `gorm:"column:reSign;NOT NULL"` // 补签次数
	IsSign        byte   `gorm:"column:isSign;NOT NULL"` // 今日是否签到
}

var SignRepo = new(signRepo)

type signRepo struct{}

func (s *signRepo) Save(ctx context.Context, db *gorm.DB, sign *Sign) error {
	if err := db.Save(sign).Error; err != nil {
		return errors.NewResponseError(constant.DatabaseError, err)
	}
	return nil
}

func (s *signRepo) Get(ctx context.Context, db *gorm.DB, roleId uint64) (*Sign, error) {
	shop := new(Sign)
	err := db.Where("`roleId` = ?", roleId).First(shop).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return shop, errors.NewResponseError(constant.DatabaseError, err)
	}
	return shop, err
}

func (s *signRepo) Updates(ctx context.Context, db *gorm.DB, roleId uint64, values interface{}) error {
	if err := db.Model(new(Sign)).Where("`roleId` = ?", roleId).Updates(values).Error; err != nil {
		return errors.NewResponseError(constant.DatabaseError, err)
	}
	return nil
}
