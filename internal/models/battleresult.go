package models

import (
	"context"
	"eggServer/internal/constant"
	"eggServer/pkg/errors"
	"gorm.io/gorm"
)

type BattlePlayerData struct {
	PetId      int32   `json:"1,omitempty"` // 参战的宠物
	Bet        []int32 `json:"2,omitempty"`
	Win        byte    `json:"3,omitempty"`
	Bonus      int32   `json:"4,omitempty"`
	Settlement byte    `json:"5,omitempty"` // 0还不可结算 1可以结算 2已经结算
	JoinAt     int64   `json:"6,omitempty"` // 加入房间时间
}

type BattleData struct {
	CreateAt        int64
	StartAt         int64
	BattleId        int32
	DeskId          string
	Players         []uint64                     // 剩余的玩家
	SettlementRound int                          // 没有真实玩家了 大于0表示结算的回合 -1表示第一回合还没开始就退出
	Result          []int32                      // 每回合的结果
	ExitPlayer      []uint64                     // 每轮退出的玩家数
	PlayerData      map[uint64]*BattlePlayerData // 所有玩家数据
	RobotRoleId     uint64                       // 机器人roleId自增
	Settlement      byte                         // 0未结算 1已结算
	State           byte                         // 0未使用 1已使用 2可以销毁
	Bonus           map[int]int32                // 每回合的奖励
}

type BattleResult struct {
	ID         uint64                       `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	CreateAt   int64                        `gorm:"column:createAt;NOT NULL"`
	StartAt    int64                        `gorm:"column:startAt;NOT NULL"`
	BattleId   int32                        `gorm:"column:battleId;NOT NULL"`
	DeskId     string                       `gorm:"column:deskId;NOT NULL"`
	Players    []uint64                     `gorm:"column:players;type:TEXT;serializer:json"`    // 剩余的玩家
	Result     []int32                      `gorm:"column:result;serializer:json"`               // 每回合的结果
	ExitPlayer []uint64                     `gorm:"column:exitPlayer;type:TEXT;serializer:json"` // 每轮退出的玩家数
	PlayerData map[uint64]*BattlePlayerData `gorm:"column:playerData;type:TEXT;serializer:json"` // 所有玩家数据
	Bonus      map[int]int32                `gorm:"column:bonus;serializer:json"`                // 每回合的奖励
}

var BattleResultRepo = new(battleResultRepo)

type battleResultRepo struct{}

func (s *battleResultRepo) Save(ctx context.Context, db *gorm.DB, battleResult *BattleResult) error {
	if err := db.Save(battleResult).Error; err != nil {
		return errors.NewResponseError(constant.DatabaseError, err)
	}
	return nil
}
