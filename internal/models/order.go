package models

import (
	"context"
	"eggServer/internal/constant"
	"eggServer/pkg/errors"
	"gorm.io/gorm"
)

type Order struct {
	ID          uint64 `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	OrderId     int64  `gorm:"column:orderId;index:idx_orderId;NOT NULL"`
	RoleID      uint64 `gorm:"column:roleId;NOT NULL"`
	Platform    int    `gorm:"column:platform"` // 平台标识 0 官方
	TotalAmount int64  `gorm:"column:totalAmount;NOT NULL"`
	OrderStatus byte   `gorm:"column:orderStatus;NOT NULL"`
	ShopId      int32  `gorm:"column:shopId;NOT NULL"`
	ShopType    byte   `gorm:"column:shopType;NOT NULL"`
	Currency    int    `gorm:"column:currency;NOT NULL"`
	CreatedAt   int64  `gorm:"column:createdAt;NOT NULL"`
	UpdatedAt   int64  `gorm:"column:updatedAt;"`
}

var OrderRepo = new(orderRepo)

type orderRepo struct{}

func (s *orderRepo) Create(ctx context.Context, db *gorm.DB, order *Order) error {
	if err := db.Create(order).Error; err != nil {
		return errors.NewResponseError(constant.DatabaseError, err)
	}
	return nil
}

func (s *orderRepo) Updates(ctx context.Context, db *gorm.DB, orderId int64, values interface{}) error {
	if err := db.Model(new(Order)).Where("`orderId` = ?", orderId).Updates(values).Error; err != nil {
		return errors.NewResponseError(constant.DatabaseError, err)
	}
	return nil
}

func (s *orderRepo) Get(ctx context.Context, db *gorm.DB, orderId int64) (*Order, error) {
	order := new(Order)
	err := db.Model(new(Order)).Where("`orderId` = ?", orderId).First(order).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return order, errors.NewResponseError(constant.DatabaseError, err)
	}
	return order, err
}
