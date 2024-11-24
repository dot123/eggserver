package models

import (
	"context"
	"eggServer/internal/constant"
	"eggServer/pkg/errors"
	"gorm.io/gorm"
)

type Invite struct {
	ID               uint64 `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	Inviter          uint64 `gorm:"column:inviter;index:idx;NOT NULL"` // 邀请人的用户id
	Invited          uint64 `gorm:"column:invited;index:idx;NOT NULL"` // 被邀请的用户id
	InvitedLastLogin int64  `gorm:"column:invitedLastLogin;"`
}

var InviteRepo = new(inviteRepo)

type inviteRepo struct{}

func (s *inviteRepo) Create(ctx context.Context, db *gorm.DB, invite *Invite) error {
	if err := db.Create(invite).Error; err != nil {
		return errors.NewResponseError(constant.DatabaseError, err)
	}
	return nil
}

func (s *inviteRepo) Save(ctx context.Context, db *gorm.DB, invite *Invite) error {
	if err := db.Save(invite).Error; err != nil {
		return errors.NewResponseError(constant.DatabaseError, err)
	}
	return nil
}

func (s *inviteRepo) FindOne(ctx context.Context, db *gorm.DB, inviter uint64, invited uint64) (*Invite, error) {
	invite := new(Invite)
	err := db.Where("(inviter = ? AND invited = ?) OR (inviter = ? AND invited = ?)", inviter, invited, invited, inviter).First(&invite).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return invite, errors.NewResponseError(constant.DatabaseError, err)
	}
	return invite, err
}

func (s *inviteRepo) Get(ctx context.Context, db *gorm.DB, inviter uint64, invited uint64) (*Invite, error) {
	invite := new(Invite)
	err := db.Where("inviter = ? AND invited = ?", inviter, invited).First(&invite).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return invite, errors.NewResponseError(constant.DatabaseError, err)
	}
	return invite, err
}

func (s *inviteRepo) UpdateColum(ctx context.Context, db *gorm.DB, id uint64, column string, value interface{}) error {
	if err := db.Model(new(Invite)).Where("`id` = ?", id).Update(column, value).Error; err != nil {
		return errors.NewResponseError(constant.DatabaseError, err)
	}
	return nil
}
