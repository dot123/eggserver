package models

import (
	"context"
	"eggServer/internal/constant"
	"eggServer/pkg/errors"
	"gorm.io/gorm"
)

//自增优点-索引性能优势
//1 查询性能
//使用自增类型主键的表在执行范围查询时，由于数据的有序性，数据库引擎可以更好地利用B+树的结构进行范围扫描，从而提高查询效率。这对于需要按主键范围进行检索的场景尤为重要。
//
//2 插入性能
//自增类型主键的另一个优势是在数据插入时的性能表现。由于新数据总是追加到索引末尾，不会触发频繁的页面分裂和数据移动，插入性能更为稳定，减少了因为主键冲突而引起的性能瓶颈。

type User struct {
	ID           uint64 `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	UserUid      string `gorm:"column:userUid;index:idx_userUid;NOT NULL"`
	UserName     string `gorm:"column:userName"`
	FirstName    string `gorm:"column:firstName"`
	LastName     string `gorm:"column:lastName"`
	PhotoUrl     string `gorm:"column:photoUrl"`
	LanguageCode string `gorm:"column:languageCode"`
	OS           string `gorm:"column:os"`
	Platform     int    `gorm:"column:platform"` // 平台标识 0 官方
	CreatedAt    int64  `gorm:"column:createdAt;"`
}

var UserRepo = new(userRepo)

type userRepo struct{}

func (s *userRepo) Create(ctx context.Context, db *gorm.DB, user *User) error {
	if err := db.Create(user).Error; err != nil {
		return errors.NewResponseError(constant.DatabaseError, err)
	}
	return nil
}

func (s *userRepo) FindOneByUserUid(ctx context.Context, db *gorm.DB, userUid string) (*User, error) {
	user := new(User)
	err := db.Where("`userUid` = ?", userUid).First(user).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return user, errors.NewResponseError(constant.DatabaseError, err)
	}
	return user, err
}

func (s *userRepo) FindOneByUserId(ctx context.Context, db *gorm.DB, userId uint64) (*User, error) {
	user := new(User)
	err := db.Where("`id` = ?", userId).First(user).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return user, errors.NewResponseError(constant.DatabaseError, err)
	}
	return user, err
}

func (s *userRepo) Updates(ctx context.Context, db *gorm.DB, userUid string, values interface{}) error {
	if err := db.Model(new(User)).Where("`userUid` = ?", userUid).Updates(values).Error; err != nil {
		return errors.NewResponseError(constant.DatabaseError, err)
	}
	return nil
}
