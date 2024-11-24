package logic

import (
	"context"
	"eggServer/internal/constant"
	"eggServer/internal/contextx"
	cfg "eggServer/internal/gamedata"
	"eggServer/internal/models"
	"eggServer/internal/schema"
	"eggServer/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cast"
	"gorm.io/gorm"
	"time"
)

var LeaderboardLogic = new(leaderboardLogic)

type leaderboardLogic struct {
	tables *cfg.Tables
}

func (s *leaderboardLogic) Init(tables *cfg.Tables) {
	s.tables = tables
}

func (s *leaderboardLogic) Data(ctx context.Context, db *gorm.DB, roleId uint64) (*schema.LeaderboardResp, error) {
	logger := contextx.FromLogger(ctx)
	rb := contextx.FromRB(ctx)
	// 分布式锁，防止多个更新并发执行
	m := rb.NewMutex("{leaderboard}_lock")
	if err := m.Lock(ctx); err != nil {
		logger.Errorf("LeaderboardLogic.Data error:%s", err.Error())
		return nil, errors.NewResponseError(constant.ServerBusy, err)
	}

	defer func() {
		if _, err := m.Unlock(context.Background()); err != nil {
			logger.WithError(err).Error("error on mutex unlock")
		}
	}()

	// 从 Redis 检查排行榜是否过期
	result, err := rb.Client().Get(ctx, "{leaderboard}_expiration").Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}

	if result == "" {
		// 如果排行榜过期，更新排行榜
		if err := s.updateLeaderboard(ctx, db); err != nil {
			return nil, err
		}

		// 设置排行榜的过期时间
		globalConfig := s.tables.GlobalTb.GetDataList()[0]
		if err := rb.Client().Set(ctx, "{leaderboard}_expiration", 1, time.Duration(globalConfig.LeaderboardExpiration)*time.Second).Err(); err != nil {
			return nil, err
		}
	}

	// 获取排行榜数据
	leaderboard, err := s.getLeaderboard(ctx, 0, 199)
	if err != nil {
		return nil, err
	}

	resp := new(schema.LeaderboardResp)
	for i, v := range leaderboard {
		rankData, err := s.getRoleRankData(ctx, db, cast.ToUint64(v))
		if err != nil {
			return nil, err
		}
		rankData.Rank = int64(i + 1)
		resp.List = append(resp.List, rankData)
	}

	// 获取当前玩家的排名
	rank, err := s.getUserRank(ctx, cast.ToString(roleId))
	if err != nil {
		return nil, err
	}
	resp.MyRankData, err = s.getRoleRankData(ctx, db, cast.ToUint64(roleId))
	if err != nil {
		return nil, err
	}
	resp.MyRankData.Rank = rank

	return resp, nil
}

func (s *leaderboardLogic) getRoleRankData(ctx context.Context, db *gorm.DB, roleId uint64) (*schema.RankData, error) {
	logger := contextx.FromLogger(ctx)
	rankData := new(schema.RankData)
	rankData.RoleId = roleId

	item, err := models.ItemRepo.Get(ctx, db, roleId, constant.Gold)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	rankData.GoldNum = item.ItemNum

	role, err := models.RoleRepo.Get(ctx, db, roleId)
	if err != nil {
		logger.Errorf("LeaderboardLogic.getRoleRankData error:%s", err.Error())
		return nil, err
	}

	user, err := models.UserRepo.FindOneByUserId(ctx, db, role.UserID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Errorf("LeaderboardLogic.getRoleRankData error:%s", err.Error())
		return nil, err
	}
	rankData.FirstName = user.FirstName
	rankData.LastName = user.LastName
	return rankData, nil
}

func (s *leaderboardLogic) updateLeaderboard(ctx context.Context, db *gorm.DB) error {
	rb := contextx.FromRB(ctx)
	const batchSize = 1000
	var items []models.Item
	// 使用双缓冲更新，临时排行榜存储在 {leaderboard}_pending
	pendingLeaderboardKey := "{leaderboard}_pending"

	// 分批读取数据并更新到 leaderboard_pending
	for offset := 0; ; offset += batchSize {
		if err := db.Model(new(models.Item)).
			Where("itemId = ?", constant.Gold).
			Limit(batchSize).
			Offset(offset).
			Find(&items).Error; err != nil {
			return err
		}

		if len(items) == 0 {
			break // 没有更多数据了
		}

		// 批量写入 Redis
		pipeline := rb.Client().Pipeline()
		for _, item := range items {
			pipeline.ZAdd(ctx, pendingLeaderboardKey, redis.Z{
				Score:  float64(item.ItemNum),
				Member: item.RoleID,
			})
		}
		_, err := pipeline.Exec(ctx)
		if err != nil {
			return err
		}
	}

	// 当 leaderboard_pending 数据准备好后，执行原子键名切换
	err := rb.Client().Rename(ctx, pendingLeaderboardKey, "{leaderboard}").Err()
	if err != nil {
		return err
	}
	return nil
}

func (s *leaderboardLogic) getLeaderboard(ctx context.Context, start, stop int64) ([]string, error) {
	rb := contextx.FromRB(ctx)
	return rb.Client().ZRevRange(ctx, "{leaderboard}", start, stop).Result()
}

// 根据 RoleId 获取玩家的排名
func (s *leaderboardLogic) getUserRank(ctx context.Context, roleId string) (int64, error) {
	rb := contextx.FromRB(ctx)
	rank, err := rb.Client().ZRevRank(ctx, "{leaderboard}", roleId).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, nil
		}
		return 0, err
	}
	return rank + 1, nil // Redis 排名是从 0 开始，返回时加 1
}

// 清除旧排行榜 (可选)
func (s *leaderboardLogic) clearLeaderboard(ctx context.Context) error {
	rb := contextx.FromRB(ctx)
	_, err := rb.Client().Del(ctx, "{leaderboard}").Result()
	return err
}
