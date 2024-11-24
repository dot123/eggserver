package models

import (
	"context"
	"eggServer/internal/constant"
	"eggServer/pkg/errors"
	"gorm.io/gorm"
)

type Egg struct {
	ID     uint64 `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	RoleID uint64 `gorm:"column:roleId;index:idx_roleId;NOT NULL"`
	Part1  int32  `gorm:"column:part1;NOT NULL"`
	Part2  int32  `gorm:"column:part2;NOT NULL"`
	Part3  int32  `gorm:"column:part3;NOT NULL"`
	IsDel  byte   `gorm:"column:isDel;NOT NULL"`
}

var EggRepo = new(eggRepo)

type eggRepo struct{}

func (s *eggRepo) Get(ctx context.Context, db *gorm.DB, roleId uint64, eggId int32) (*Egg, error) {
	egg := new(Egg)
	err := db.Where("roleId=? and id=?", roleId, eggId).First(egg).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return egg, errors.NewResponseError(constant.DatabaseError, err)
	}
	return egg, err
}

func (s *eggRepo) Create(ctx context.Context, db *gorm.DB, egg *Egg) error {
	if err := db.Create(egg).Error; err != nil {
		return errors.NewResponseError(constant.DatabaseError, err)
	}
	return nil
}

func (s *eggRepo) Delete(ctx context.Context, db *gorm.DB, egg *Egg) error {
	if err := db.Delete(egg).Error; err != nil {
		return errors.NewResponseError(constant.DatabaseError, err)
	}
	return nil
}

func (s *eggRepo) FindAllByRoleId(ctx context.Context, db *gorm.DB, roleId uint64) ([]*Egg, error) {
	list := make([]*Egg, 0)
	err := db.Model(new(Egg)).Where("roleId=? and isDel=0", roleId).Find(&list).Error
	return list, err
}

func (s *eggRepo) UpdateColum(ctx context.Context, db *gorm.DB, id uint64, column string, value interface{}) error {
	if err := db.Model(new(Egg)).Where("`id` = ?", id).Update(column, value).Error; err != nil {
		return errors.NewResponseError(constant.DatabaseError, err)
	}
	return nil
}
