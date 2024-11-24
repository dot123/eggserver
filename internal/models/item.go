package models

import (
	"context"
	"eggServer/internal/constant"
	"eggServer/pkg/errors"
	"gorm.io/gorm"
)

type Item struct {
	ID      uint64 `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	RoleID  uint64 `gorm:"column:roleId;index:idx_roleId;NOT NULL"`
	ItemID  int32  `gorm:"column:itemId;NOT NULL"`
	ItemNum int32  `gorm:"column:itemNum;"`
}

var ItemRepo = new(itemRepo)

type itemRepo struct{}

func (s *itemRepo) Get(ctx context.Context, db *gorm.DB, roleId uint64, itemId int32) (*Item, error) {
	item := new(Item)
	err := db.Where("roleId=? and itemId=?", roleId, itemId).First(item).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return item, errors.NewResponseError(constant.DatabaseError, err)
	}
	return item, err
}

func (s *itemRepo) FindAllByRoleId(ctx context.Context, db *gorm.DB, roleId uint64) ([]*Item, error) {
	items := make([]*Item, 0)
	err := db.Where("roleId=? and itemNum>0", roleId).Find(&items).Error
	return items, err
}

func (s *itemRepo) Save(ctx context.Context, db *gorm.DB, item *Item) error {
	if err := db.Save(item).Error; err != nil {
		return errors.NewResponseError(constant.DatabaseError, err)
	}
	return nil
}
