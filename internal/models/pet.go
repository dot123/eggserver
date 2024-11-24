package models

import (
	"context"
	"eggServer/internal/constant"
	"eggServer/pkg/errors"
	"gorm.io/gorm"
)

type Pet struct {
	ID     uint64 `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	RoleID uint64 `gorm:"column:roleId;index:idx_roleId;NOT NULL"`
	PetID  int32  `gorm:"column:petId;NOT NULL"`
	PetNum int32  `gorm:"column:petNum;"`
}

var PetRepo = new(petRepo)

type petRepo struct{}

func (s *petRepo) Get(ctx context.Context, db *gorm.DB, roleId uint64, petId int32) (*Pet, error) {
	pet := new(Pet)
	err := db.Where("roleId=? and petId=?", roleId, petId).First(pet).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return pet, errors.NewResponseError(constant.DatabaseError, err)
	}
	return pet, err
}

func (s *petRepo) FindAllByRoleId(ctx context.Context, db *gorm.DB, roleId uint64) ([]*Pet, error) {
	pets := make([]*Pet, 0)
	err := db.Where("roleId=? and petNum>0", roleId).Find(&pets).Error
	return pets, err
}

func (s *petRepo) Save(ctx context.Context, db *gorm.DB, pet *Pet) error {
	if err := db.Save(pet).Error; err != nil {
		return errors.NewResponseError(constant.DatabaseError, err)
	}
	return nil
}
