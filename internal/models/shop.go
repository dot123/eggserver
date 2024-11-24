package models

import (
	"context"
	"eggServer/internal/constant"
	"eggServer/pkg/errors"
	"gorm.io/gorm"
)

type Shop struct {
	ID           uint64        `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	RoleID       uint64        `gorm:"column:roleId;index:idx_roleId;NOT NULL"`
	Discount     []int32       `gorm:"column:discount;serializer:json"`
	List         []int32       `gorm:"column:list;serializer:json"`
	Buy          []int         `gorm:"column:buy;serializer:json"`
	Share        map[int][]int `gorm:"column:share;serializer:json"`
	RefreshTimes byte          `gorm:"column:refreshTimes"`
}

var ShopRepo = new(shopRepo)

type shopRepo struct{}

func (s *shopRepo) Save(ctx context.Context, db *gorm.DB, shop *Shop) error {
	if err := db.Save(shop).Error; err != nil {
		return errors.NewResponseError(constant.DatabaseError, err)
	}
	return nil
}

func (s *shopRepo) Get(ctx context.Context, db *gorm.DB, roleId uint64) (*Shop, error) {
	shop := new(Shop)
	err := db.Where("roleId=?", roleId).First(shop).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return shop, errors.NewResponseError(constant.DatabaseError, err)
	}
	return shop, err
}
