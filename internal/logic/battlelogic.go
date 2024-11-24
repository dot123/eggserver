package logic

import (
	"context"
	"eggServer/internal/constant"
	"eggServer/internal/contextx"
	cfg "eggServer/internal/gamedata"
	"eggServer/internal/models"
	"eggServer/internal/schema"
	"eggServer/pkg/errors"
	"eggServer/pkg/redisbackend"
	"eggServer/pkg/utils"
	Xorshift "eggServer/pkg/utils/xorshift"
	"encoding/json"
	"fmt"
	"github.com/patrickmn/go-cache"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cast"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
	"math"
	"math/rand"
	"strings"
	"time"
)

var (
	BattleDeskIdKey             = "battle:deskId:%d"
	BattleDataKey               = "battle:data:%s"
	BattleMatchKey              = "battle:match:%d"
	BattleRegistrationKey       = "battleRegistration:%s"
	BattleScoreKey              = "battleScore:%s"
	BattleNeedPetNum      int32 = 1
)

var BattleLogic = new(battleLogic)

type battleLogic struct {
	tables   *cfg.Tables
	g        singleflight.Group
	cache    *cache.Cache
	xorshift *Xorshift.Xorshift
}

func (s *battleLogic) Init(tables *cfg.Tables) {
	s.tables = tables

	// 使用当前时间的纳秒部分作为种子
	seed := uint32(time.Now().UnixNano())
	s.xorshift = Xorshift.NewXorshift(seed)

	// 内存缓存
	s.cache = cache.New(time.Second, time.Minute)
}

// 检查无效的战斗数据
func (s *battleLogic) checkInvalidBattleData(ctx context.Context, rb *redisbackend.RedisBackend, battleData *models.BattleData, battleId int32) (bool, error) {
	logger := contextx.FromLogger(ctx)
	battleConfig := s.tables.BattleConfigTb.Get(battleId)
	if battleData != nil && time.Now().Unix() >= battleData.CreateAt+int64(battleConfig.MatchTime) {
		if battleData.State == 0 || battleData.State == 2 {
			if err := rb.Client().Del(ctx, fmt.Sprintf(BattleDataKey, battleData.DeskId)).Err(); err != nil {
				logger.Errorf("BattleLogic.checkInvalidBattleData error:%s", err.Error())
				return true, err
			}
			return true, nil
		}
	}
	return false, nil
}

// Match 匹配
func (s *battleLogic) Match(ctx context.Context, db *gorm.DB, roleId uint64, battleId int32, petId int32) (*schema.BattleMatchResp, error) {
	logger := contextx.FromLogger(ctx)
	rb := contextx.FromRB(ctx)
	battleConfig := s.tables.BattleConfigTb.Get(battleId)
	if battleConfig == nil {
		return nil, errors.NewResponseError(constant.ParametersInvalid, nil)
	}

	// 分布式锁
	m := rb.NewMutex(fmt.Sprintf(BattleMatchKey, battleId))
	if err := m.Lock(ctx); err != nil {
		logger.Errorf("BattleLogic.Match error:%s", err.Error())
		return nil, errors.NewResponseError(constant.ServerBusy, err)
	}

	defer func() {
		if _, err := m.Unlock(context.Background()); err != nil {
			logger.WithError(err).Error("error on mutex unlock")
		}
	}()

	key := fmt.Sprintf(BattleDeskIdKey, battleId)

	// 获取最后的房间号
	ret, err := rb.Client().Get(ctx, key).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}

	if ret == "" {
		// 设置初始房间号为1
		ret, err = rb.Client().Set(ctx, key, 1, redis.KeepTTL).Result()
		if err != nil {
			return nil, err
		}
	}

	// 最后的房间号
	deskId := fmt.Sprintf("%d-%s", battleId, ret)
	battleData, err := s.getBattleData(ctx, rb, deskId)
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}

	isInvalid, err := s.checkInvalidBattleData(ctx, rb, battleData, battleId)
	if err != nil {
		return nil, err
	}

	if battleConfig.IsGuide == 1 {
		guide, err := models.GuideRepo.Get(ctx, db, roleId)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		// 已经进行新手战斗引导了
		if utils.InArray(guide.Step, 306) {
			return nil, errors.NewResponseError(constant.ParametersInvalid, nil)
		}
		// 记录新手引导
		if err := GuideLogic.Step(ctx, db, roleId, 306); err != nil {
			return nil, err
		}
	}

	// 战斗引导或战斗房间无效或战斗还没开始并且人数已经满了或战斗已经开始了
	if battleConfig.IsGuide == 1 || battleData == nil || isInvalid || (!s.isBattleStart(battleData) && len(battleData.Players) == int(battleConfig.PlayerNum)) || s.isBattleStart(battleData) {
		// 房间号自加
		ret, err := rb.Client().Incr(ctx, key).Result()
		if err != nil {
			return nil, err
		}
		deskId = fmt.Sprintf("%d-%d", battleId, ret)
		// 创建新房间
		battleData, err = s.createBattleData(ctx, rb, battleId, deskId)
		if err != nil {
			logger.Errorf("BattleLogic.Match error:%s", err.Error())
			return nil, err
		}
	}

	return s.join(ctx, db, rb, roleId, battleData, petId)
}

// MatchState 匹配状态
func (s *battleLogic) MatchState(ctx context.Context, roleId uint64, deskId string) (*schema.BattleMatchStateResp, error) {
	logger := contextx.FromLogger(ctx)
	rb := contextx.FromRB(ctx)
	// 分布式锁
	m := rb.NewMutex(deskId)
	if err := m.Lock(ctx); err != nil {
		logger.Errorf("BattleLogic.MatchState error:%s", err.Error())
		return nil, errors.NewResponseError(constant.ServerBusy, err)
	}

	defer func() {
		if _, err := m.Unlock(context.Background()); err != nil {
			logger.WithError(err).Error("error on mutex unlock")
		}
	}()

	battleData, err := s.getBattleData(ctx, rb, deskId)
	// 检查错误
	if err := s.checkError(battleData, roleId, err); err != nil {
		logger.Errorf("BattleLogic.MatchState error:%s", err.Error())
		return nil, err
	}

	// 机器人逻辑
	if err := s.robotJoin(ctx, rb, battleData, true); err != nil {
		return nil, err
	}

	resp := s.getBattleMatchStateResp(roleId, battleData)
	return resp, nil
}

func (s *battleLogic) getBattleMatchStateResp(roleId uint64, battleData *models.BattleData) *schema.BattleMatchStateResp {
	battleConfig := s.tables.BattleConfigTb.Get(battleData.BattleId)

	resp := new(schema.BattleMatchStateResp)
	resp.BattleId = battleData.BattleId
	resp.CreatedAt = battleData.CreateAt
	resp.StartAt = battleData.StartAt
	playerData := s.getPlayerData(battleData, roleId)
	resp.JoinAt = playerData.JoinAt

	if battleData.StartAt == 0 {
		if time.Now().Unix() >= battleData.CreateAt+int64(battleConfig.MatchTime) {
			resp.StartAt = battleData.CreateAt + int64(battleConfig.MatchTime)
		}
	}

	resp.PlayerNum = len(battleData.Players)
	resp.DeskId = battleData.DeskId
	return resp
}

// 机器人加入逻辑
func (s *battleLogic) robotJoin(ctx context.Context, rb *redisbackend.RedisBackend, battleData *models.BattleData, isSaveBattleData bool) error {
	logger := contextx.FromLogger(ctx)
	if battleData == nil {
		return nil
	}

	battleConfig := s.tables.BattleConfigTb.Get(battleData.BattleId)
	if battleConfig.AddRobot == 0 {
		return nil
	}

	if len(battleData.Players) < int(battleConfig.PlayerNum) && len(battleData.PlayerData) < int(battleConfig.PlayerNum) {
		t := time.Now().Unix() - battleData.CreateAt
		p := int(math.Ceil((float64(t) / float64(battleConfig.MatchTime)) * float64(battleConfig.PlayerNum)))
		playerNum := len(battleData.Players)
		if p > int(battleConfig.PlayerNum) {
			p = int(battleConfig.PlayerNum)
		}
		// 人数不足自动补机器人
		if p > playerNum {
			n := p - playerNum
			petConfigList := s.tables.PetTb.GetDataList()
			for i := 0; i < n; i++ {
				petConfig, _ := utils.RandomElement(petConfigList)
				battleData.RobotRoleId = battleData.RobotRoleId + 1
				s.dealJoin(true, battleData.RobotRoleId, battleData, petConfig.Id)
			}

			if isSaveBattleData {
				if err := s.saveBattleData(ctx, rb, battleData); err != nil {
					logger.Errorf("BattleLogic.robotJoin error:%s", err.Error())
					return err
				}
			}
		}
	}
	return nil
}

// 是否是机器人
func (s *battleLogic) isRobot(roleId uint64) bool {
	return roleId <= 10000
}

// 是否所有都是机器人
func (s *battleLogic) isAllRobot(players []uint64) bool {
	for _, roleId := range players {
		if !s.isRobot(roleId) {
			return false
		}
	}
	return true
}

// 机器人人数
func (s *battleLogic) getRobotLeftNum(players []uint64) int {
	num := 0
	for _, roleId := range players {
		if s.isRobot(roleId) {
			num++
		}
	}
	return num
}

// 机器人下注人数
func (s *battleLogic) getRobotBetNum(battleData *models.BattleData, round int) int {
	i := 0
	for _, roleId := range battleData.Players {
		if s.isRobot(roleId) {
			playerData := s.getPlayerData(battleData, roleId)
			if len(playerData.Bet) >= round {
				i++
			}
		}
	}
	return i
}

// 机器人下注
func (s *battleLogic) robotBet(ctx context.Context, rb *redisbackend.RedisBackend, battleData *models.BattleData, isSaveBattleData bool, round int) error {
	logger := contextx.FromLogger(ctx)
	if battleData == nil {
		return nil
	}
	battleConfig := s.tables.BattleConfigTb.Get(battleData.BattleId)
	if battleConfig.AddRobot == 1 {
		roundTime := battleConfig.RoundTimes[round-1]
		robotNum := s.getRobotLeftNum(battleData.Players)
		t := int64(roundTime) - (s.getRoundStartTime(battleData, round) - time.Now().Unix())
		n := int(math.Ceil(float64(t) / float64(roundTime) * float64(robotNum)))
		robotBetNum := s.getRobotBetNum(battleData, round)

		if n > robotNum {
			n = robotNum
		}

		if n > robotBetNum {
			gridList := s.getGridList(battleData)
			i := 0
			for _, roleId := range battleData.Players {
				if s.isRobot(roleId) {
					i++
					if i <= n {
						playerData := s.getPlayerData(battleData, roleId)
						if len(playerData.Bet) < round {
							// 随机格子
							grid := gridList[s.xorshift.Intn(len(gridList))]
							s.bet(roleId, grid, round, battleData)
						}
					}
				}
			}

			if isSaveBattleData {
				if err := s.saveBattleData(ctx, rb, battleData); err != nil {
					logger.Errorf("BattleLogic.robotBet error:%s", err.Error())
					return err
				}
			}
		}
	}
	return nil
}

func (s *battleLogic) checkError(battleData *models.BattleData, roleId uint64, err error) error {
	if err != nil {
		if errors.Is(err, redis.Nil) {
			// 房间已经解散
			return errors.NewResponseError(constant.BattleAlreadyDismiss, nil)
		}
		return err
	}

	// 你不在本回合
	if !utils.InArray(battleData.Players, roleId) {
		return errors.NewResponseError(constant.BattleYouNotInRound, nil)
	}

	return nil
}

// 检查参赛资格
func (s *battleLogic) checkQualified(ctx context.Context, db *gorm.DB, roleId uint64, needItemList []*cfg.GlobalItemData) (bool, error) {
	ret := true
	for _, need := range needItemList {
		if need.Type == cfg.RewardType_Item {
			item, err := models.ItemRepo.Get(ctx, db, roleId, need.Id)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return false, err
			}
			if !(item.ItemNum >= need.Num) {
				ret = false
			}
		} else if need.Type == cfg.RewardType_Pet {
			item, err := models.PetRepo.Get(ctx, db, roleId, need.Id)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return false, err
			}
			if !(item.PetNum >= need.Num) {
				ret = false
			}
		}
	}
	return ret, nil
}

// join 加入房间
func (s *battleLogic) join(ctx context.Context, db *gorm.DB, rb *redisbackend.RedisBackend, roleId uint64, battleData *models.BattleData, petId int32) (*schema.BattleMatchResp, error) {
	logger := contextx.FromLogger(ctx)
	if s.tables.PetTb.Get(petId) == nil {
		return nil, errors.NewResponseError(constant.ParametersInvalid, nil)
	}

	battleConfig := s.tables.BattleConfigTb.Get(battleData.BattleId)
	if battleConfig == nil {
		return nil, errors.NewResponseError(constant.ParametersInvalid, nil)
	}

	if x, found := s.cache.Get(fmt.Sprintf(BattleRegistrationKey, battleData.DeskId)); found {
		// 报名人数已满
		if int(battleConfig.PlayerNum) <= x.(int) {
			return nil, errors.NewResponseError(constant.BattleRegistrationFull, nil)
		}
	}

	// 已经开始战斗
	if s.isBattleStart(battleData) {
		return nil, errors.NewResponseError(constant.BattleAlreadyStarted, nil)
	}

	// 报名人数已满
	if int(battleConfig.PlayerNum) <= len(battleData.Players) {
		// 内存缓存
		s.cache.Set(fmt.Sprintf(BattleRegistrationKey, battleData.DeskId), len(battleData.Players), time.Second)
		return nil, errors.NewResponseError(constant.BattleRegistrationFull, nil)
	}

	ret, err := s.checkQualified(ctx, db, roleId, battleConfig.Need)
	if err != nil {
		return nil, err
	}

	// 不符合报名资格
	if !ret {
		return nil, errors.NewResponseError(constant.NotQualified, nil)
	}

	role, err := models.RoleRepo.Get(ctx, db, roleId)
	if err != nil {
		logger.Errorf("BattleLogic.join error:%s", err.Error())
		return nil, errors.NewResponseError(constant.DatabaseError, err)
	}

	// 已报名其他战场
	if role.LastDeskId != "" && role.LastDeskId != battleData.DeskId {
		return nil, errors.NewResponseError(constant.AlreadyJoinOtherBattle, nil)
	}

	resp := new(schema.BattleMatchResp)

	if !utils.InArray(battleData.Players, roleId) {
		err = db.Transaction(func(tx *gorm.DB) error {
			if err := models.RoleRepo.Updates(ctx, db, roleId, map[string]interface{}{"lastDeskId": battleData.DeskId, "battleCount": role.BattleCount + 1}); err != nil {
				return err
			}

			reward, err := UtilsLogic.AddItem(ctx, db, roleId, petId, -BattleNeedPetNum, cfg.RewardType_Pet)
			if err != nil {
				return err
			}
			resp.Reward = reward
			return nil
		})

		if err != nil {
			logger.Errorf("BattleLogic.join error:%s", err.Error())
			return nil, err
		}

		s.dealJoin(false, roleId, battleData, petId)
		if err := s.saveBattleData(ctx, rb, battleData); err != nil {
			logger.Errorf("BattleLogic.join error:%s", err.Error())
			return nil, err
		}
	}

	resp.MatchState = s.getBattleMatchStateResp(roleId, battleData)
	return resp, nil
}

// 处理加入战斗
func (s *battleLogic) dealJoin(isRobot bool, roleId uint64, battleData *models.BattleData, petId int32) {
	// 保存参战的宠物
	playerData := s.getPlayerData(battleData, roleId)
	playerData.PetId = petId

	// 加入房间时间
	playerData.JoinAt = 0
	if !isRobot {
		playerData.JoinAt = time.Now().Unix()
	}

	// 加入本局战斗
	battleData.Players = append(battleData.Players, roleId)
	battleData.State = 1
}

// Leave 离开房间
func (s *battleLogic) Leave(ctx context.Context, db *gorm.DB, roleId uint64, deskId string) (*schema.BattleLeaveResp, error) {
	logger := contextx.FromLogger(ctx)
	rb := contextx.FromRB(ctx)
	// 分布式锁
	m := rb.NewMutex(deskId)
	if err := m.Lock(ctx); err != nil {
		logger.Errorf("BattleLogic.Leave error:%s", err.Error())
		return nil, errors.NewResponseError(constant.ServerBusy, err)
	}

	defer func() {
		if _, err := m.Unlock(context.Background()); err != nil {
			logger.WithError(err).Error("error on mutex unlock")
		}
	}()

	battleData, err := s.getBattleData(ctx, rb, deskId)
	// 检查错误
	if err := s.checkError(battleData, roleId, err); err != nil {
		logger.Errorf("BattleLogic.Leave error:%s", err.Error())
		return nil, err
	}

	// 已经开始战斗
	if s.isBattleStart(battleData) {
		return nil, errors.NewResponseError(constant.BattleAlreadyStarted, nil)
	}

	resp := new(schema.BattleLeaveResp)
	players, ret := utils.RemoveElement(battleData.Players, roleId)
	if ret {
		playerData := s.getPlayerData(battleData, roleId)
		if playerData.PetId > 0 {
			role, err := models.RoleRepo.Get(ctx, db, roleId)
			if err != nil {
				logger.Errorf("BattleLogic.Leave error:%s", err.Error())
				return nil, errors.NewResponseError(constant.DatabaseError, err)
			}

			err = db.Transaction(func(tx *gorm.DB) error {
				if err := models.RoleRepo.Updates(ctx, db, roleId, map[string]interface{}{"lastDeskId": "", "battleCount": role.BattleCount - 1}); err != nil {
					return err
				}
				reward, err := UtilsLogic.AddItem(ctx, db, roleId, playerData.PetId, BattleNeedPetNum, cfg.RewardType_Pet)
				if err != nil {
					return err
				}
				resp.Reward = reward
				return nil
			})

			if err != nil {
				logger.Errorf("BattleLogic.Leave error:%s", err.Error())
				return nil, err
			}
		}

		delete(battleData.PlayerData, roleId)
		// 参加的总人数
		battleData.Players = players
		if len(battleData.Players) == 0 {
			battleData.State = 0
		}

		// 都是机器人
		if s.isAllRobot(battleData.Players) {
			battleData.State = 0
		}

		// 内存缓存
		s.cache.Set(fmt.Sprintf(BattleRegistrationKey, deskId), len(battleData.Players), time.Second)

		if err := s.saveBattleData(ctx, rb, battleData); err != nil {
			logger.Errorf("BattleLogic.Leave error:%s", err.Error())
			return nil, err
		}
	}

	return resp, nil
}

// Exit 退出战斗
func (s *battleLogic) Exit(ctx context.Context, roleId uint64, deskId string) (*schema.BattleExitResp, error) {
	logger := contextx.FromLogger(ctx)
	resp := new(schema.BattleExitResp)
	resp.State = 1
	rb := contextx.FromRB(ctx)
	// 分布式锁
	m := rb.NewMutex(deskId)
	if err := m.Lock(ctx); err != nil {
		logger.Errorf("BattleLogic.Exit error:%s", err.Error())
		return resp, errors.NewResponseError(constant.ServerBusy, err)
	}

	defer func() {
		if _, err := m.Unlock(context.Background()); err != nil {
			logger.WithError(err).Error("error on mutex unlock")
		}
	}()

	battleData, err := s.getBattleData(ctx, rb, deskId)
	// 检查错误
	if err := s.checkError(battleData, roleId, err); err != nil {
		logger.Errorf("BattleLogic.Exit error:%s", err.Error())
		resp.State = 0 // 立即退出
		return resp, err
	}

	// 战斗还没开始
	if !s.isBattleStart(battleData) {
		return nil, errors.NewResponseError(constant.ParametersInvalid, nil)
	}

	// 战斗已经结束
	round := s.getRound(battleData)
	if s.isBattleEnd(battleData, round) {
		return nil, errors.NewResponseError(constant.ParametersInvalid, nil)
	}

	if utils.InArray(battleData.Players, roleId) {
		playerData := s.getPlayerData(battleData, roleId)
		round := s.getRound(battleData)
		if !s.isRoundStart(battleData, round) { // 回合还没开始
			// 可以结算
			playerData.Settlement = 1

			battleData.Players, _ = utils.RemoveElement(battleData.Players, roleId)

			if !utils.InArray(battleData.ExitPlayer, roleId) {
				battleData.ExitPlayer = append(battleData.ExitPlayer, roleId)
			}

			if len(battleData.Players) == 0 || s.isAllRobot(battleData.Players) {
				if round == 1 {
					// 第一回合还没开始就退出
					battleData.SettlementRound = -1
				} else {
					// 上一回合结算
					battleData.SettlementRound = round - 1
				}
			}

			if err := s.saveBattleData(ctx, rb, battleData); err != nil {
				logger.Errorf("BattleLogic.Exit error:%s", err.Error())
				return resp, err
			}

			// 立即退出
			resp.State = 0
		}
	}

	return resp, nil
}

// Bet 下注
func (s *battleLogic) Bet(ctx context.Context, roleId uint64, req *schema.BattleBetReq) (*schema.BattleSyncScoreResp, error) {
	logger := contextx.FromLogger(ctx)
	rb := contextx.FromRB(ctx)
	// 分布式锁
	m := rb.NewMutex(req.DeskId)
	if err := m.Lock(ctx); err != nil {
		logger.Errorf("BattleLogic.Bet error:%s", err.Error())
		return nil, errors.NewResponseError(constant.ServerBusy, err)
	}

	defer func() {
		if _, err := m.Unlock(context.Background()); err != nil {
			logger.WithError(err).Error("error on mutex unlock")
		}
	}()

	battleData, err := s.getBattleData(ctx, rb, req.DeskId)
	// 检查错误
	if err := s.checkError(battleData, roleId, err); err != nil {
		logger.Errorf("BattleLogic.Bet error:%s", err.Error())
		return nil, err
	}

	// 战斗还没开始
	if !s.isBattleStart(battleData) {
		return nil, errors.NewResponseError(constant.ParametersInvalid, nil)
	}

	// 战斗已经结束
	round := s.getRound(battleData)
	if s.isBattleEnd(battleData, round) {
		return nil, errors.NewResponseError(constant.ParametersInvalid, nil)
	}

	// 获取玩家战斗数据
	playerData := s.getPlayerData(battleData, roleId)

	if utils.InArray(battleData.Result, req.Grid) || req.Grid < 1 {
		return nil, errors.NewResponseError(constant.ParametersInvalid, nil)
	}

	// 保存本局本回合下注数据
	s.bet(roleId, req.Grid, round, battleData)

	// 机器人下注
	if err := s.robotBet(ctx, rb, battleData, false, round); err != nil {
		return nil, err
	}

	if err := s.saveBattleData(ctx, rb, battleData); err != nil {
		logger.Errorf("BattleLogic.Bet error:%s", err.Error())
		return nil, errors.NewResponseError(constant.RDBError, nil)
	}

	resp := s.getBattleSyncScoreResp(battleData, playerData, round)
	return resp, nil
}

// 下注
func (s *battleLogic) bet(roleId uint64, grid int32, round int, battleData *models.BattleData) {
	// 保存下注信息和参战宠物数量
	playerData := s.getPlayerData(battleData, roleId)

	if len(playerData.Bet) < round {
		playerData.Bet = append(playerData.Bet, grid)
	} else {
		playerData.Bet[round-1] = grid
	}

	// 使内存缓存失效
	s.cache.Delete(fmt.Sprintf(BattleScoreKey, battleData.DeskId))
}

// 获取玩家数据
func (s *battleLogic) getPlayerData(battleData *models.BattleData, roleId uint64) *models.BattlePlayerData {
	playerData := battleData.PlayerData[roleId]
	if playerData == nil {
		battleData.PlayerData[roleId] = new(models.BattlePlayerData)
		battleData.PlayerData[roleId].Bet = make([]int32, 0)
	}
	return battleData.PlayerData[roleId]
}

// 创建战斗数据
func (s *battleLogic) createBattleData(ctx context.Context, rb *redisbackend.RedisBackend, battleId int32, deskId string) (*models.BattleData, error) {
	battleData := new(models.BattleData)
	battleData.CreateAt = time.Now().Unix()
	battleData.BattleId = battleId
	battleData.DeskId = deskId
	battleData.Players = make([]uint64, 0)
	battleData.SettlementRound = 0
	battleData.Result = make([]int32, 0)
	battleData.ExitPlayer = make([]uint64, 0)
	battleData.PlayerData = make(map[uint64]*models.BattlePlayerData)
	battleData.Bonus = make(map[int]int32)

	jsonData, err := json.Marshal(battleData)
	if err != nil {
		return nil, errors.NewResponseError(constant.JsonMarshalError, err)
	}

	if err := rb.Client().Set(ctx, fmt.Sprintf(BattleDataKey, deskId), jsonData, redis.KeepTTL).Err(); err != nil {
		return nil, errors.NewResponseError(constant.RDBError, err)
	}

	return battleData, nil
}

// 获取战斗数据
func (s *battleLogic) getBattleData(ctx context.Context, rb *redisbackend.RedisBackend, deskId string) (*models.BattleData, error) {
	result, err := rb.Client().Get(ctx, fmt.Sprintf(BattleDataKey, deskId)).Result()
	if err != nil {
		return nil, err
	}

	battleData := new(models.BattleData)
	if err := json.Unmarshal([]byte(result), battleData); err != nil {
		return nil, errors.NewResponseError(constant.JsonUnmarshalError, err)
	}
	return battleData, nil
}

// 保存战斗数据
func (s *battleLogic) saveBattleData(ctx context.Context, rb *redisbackend.RedisBackend, battleData *models.BattleData) error {
	val, err := json.Marshal(battleData)
	if err != nil {
		return errors.NewResponseError(constant.JsonMarshalError, err)
	}

	if err := rb.Client().Set(ctx, fmt.Sprintf(BattleDataKey, battleData.DeskId), val, redis.KeepTTL).Err(); err != nil {
		return errors.NewResponseError(constant.RDBError, err)
	}

	return nil
}

// 当前的回合
func (s *battleLogic) getRound(battleData *models.BattleData) int {
	if battleData.StartAt == 0 {
		return 0
	}
	battleConfig := s.tables.BattleConfigTb.Get(battleData.BattleId)

	var temp int32
	n := int32(time.Now().Unix() - battleData.StartAt)
	for i := 0; i < int(battleConfig.TotalRound); i++ {
		if n >= temp && n <= temp+battleConfig.RoundTimes[i]+battleConfig.RoundInterval {
			return i + 1
		}
		temp = temp + battleConfig.RoundTimes[i] + battleConfig.RoundInterval
	}

	if n > temp {
		return int(battleConfig.TotalRound)
	}
	return 0
}

func (s *battleLogic) calcPlayerBonus(battleData *models.BattleData, round int) int32 {
	bonus := s.getBaseBonus(battleData) * int32(round)
	battleConfig := s.tables.BattleConfigTb.Get(battleData.BattleId)
	if int32(round) == battleConfig.TotalRound {
		bonus = battleConfig.TotalBonus
	}
	// 离开的人带走的奖金
	for _, roleId := range battleData.ExitPlayer {
		playerData := s.getPlayerData(battleData, roleId)
		bonus = bonus - playerData.Bonus
	}

	p := len(battleData.Players)
	if p > 0 {
		bonus = bonus / int32(p)
	}
	return bonus
}

func (s *battleLogic) getBaseBonus(battleData *models.BattleData) int32 {
	battleConfig := s.tables.BattleConfigTb.Get(battleData.BattleId)
	bonus := battleConfig.TotalBonus / battleConfig.GridNum
	return bonus
}

func (s *battleLogic) BattleSyncScore(ctx context.Context, roleId uint64, deskId string) (*schema.BattleSyncScoreResp, error) {
	logger := contextx.FromLogger(ctx)
	rb := contextx.FromRB(ctx)
	// 分布式锁
	m := rb.NewMutex(deskId)
	if err := m.Lock(ctx); err != nil {
		logger.Errorf("BattleLogic.BattleSyncScore error:%s", err.Error())
		return nil, errors.NewResponseError(constant.ServerBusy, err)
	}

	defer func() {
		if _, err := m.Unlock(context.Background()); err != nil {
			logger.WithError(err).Error("error on mutex unlock")
		}
	}()

	battleData, err := s.getBattleData(ctx, rb, deskId)
	// 检查错误
	if err != nil {
		if errors.Is(err, redis.Nil) {
			// 房间已经解散
			return nil, nil
		}
		return nil, err
	}

	round := s.getRound(battleData)
	// 游戏还没开始
	if round == 0 {
		return nil, errors.NewResponseError(constant.ParametersInvalid, nil)
	}

	// 机器人下注
	if err := s.robotBet(ctx, rb, battleData, true, round); err != nil {
		return nil, err
	}

	playerData := s.getPlayerData(battleData, roleId)
	return s.getBattleSyncScoreResp(battleData, playerData, round), err
}

func (s *battleLogic) getBattleSyncScoreResp(battleData *models.BattleData, playerData *models.BattlePlayerData, round int) *schema.BattleSyncScoreResp {
	resp := new(schema.BattleSyncScoreResp)
	resp.ScoreList = make(map[int32]int32)
	resp.PetList = make(map[int32][]int32)
	resp.Grid = -1

	// 统计分数
	for _, playerData := range battleData.PlayerData {
		if len(playerData.Bet) >= round {
			grid := playerData.Bet[round-1]
			resp.PetList[grid] = append(resp.PetList[grid], playerData.PetId)
			resp.ScoreList[grid] = resp.ScoreList[grid] + battleData.Bonus[round-1]
		}
	}

	// 排除自己的宠物
	if len(playerData.Bet) >= round {
		grid := playerData.Bet[round-1]
		resp.Grid = grid
		resp.PetList[grid], _ = utils.RemoveElement(resp.PetList[grid], playerData.PetId)
	}

	// 最后下注的回合
	if len(battleData.Result) >= round {
		resp.Grid = playerData.Bet[len(playerData.Bet)-1]
	}

	resp.PlayerBonus = battleData.Bonus[round-1]
	resp.BaseBonus = s.getBaseBonus(battleData)

	return resp
}

// 回合开始时间
func (s *battleLogic) getRoundStartTime(battleData *models.BattleData, round int) int64 {
	battleConfig := s.tables.BattleConfigTb.Get(battleData.BattleId)
	var totalRoundTime int32 = 0
	for i, roundTime := range battleConfig.RoundTimes {
		if i < round {
			totalRoundTime = totalRoundTime + roundTime
		}
	}
	if battleData.StartAt == 0 {
		return battleData.StartAt + int64(battleConfig.MatchTime+battleConfig.RoundInterval*int32(round-1)) + int64(totalRoundTime) // 每回合开始时间
	}
	return battleData.StartAt + int64(battleConfig.RoundInterval*int32(round-1)) + int64(totalRoundTime) // 每回合开始时间
}

// 回合是否开始
func (s *battleLogic) isRoundStart(battleData *models.BattleData, round int) bool {
	if battleData.StartAt == 0 {
		return false
	}
	if time.Now().Unix() > s.getRoundStartTime(battleData, round) {
		return true
	}

	return false
}

// 回合是否结束
func (s *battleLogic) isRoundEnd(battleData *models.BattleData, round int) bool {
	if battleData.StartAt == 0 {
		return false
	}
	battleConfig := s.tables.BattleConfigTb.Get(battleData.BattleId)
	if time.Now().Unix() > s.getRoundStartTime(battleData, round)+int64(battleConfig.RoundInterval) {
		return true
	}

	return false
}

// 战斗是否开始
func (s *battleLogic) isBattleStart(battleData *models.BattleData) bool {
	battleConfig := s.tables.BattleConfigTb.Get(battleData.BattleId)
	if time.Now().Unix() >= battleData.CreateAt+int64(battleConfig.MatchTime) {
		return true
	}
	return battleData.StartAt > 0
}

// 战斗是否结束
func (s *battleLogic) isBattleEnd(battleData *models.BattleData, round int) bool {
	if battleData.StartAt == 0 {
		return false
	}

	battleConfig := s.tables.BattleConfigTb.Get(battleData.BattleId)
	if int32(round) != battleConfig.TotalRound {
		return false
	}

	return s.isRoundEnd(battleData, int(battleConfig.TotalRound))
}

// GetRoundResult 获取回合结果
func (s *battleLogic) GetRoundResult(ctx context.Context, roleId uint64, deskId string) (*schema.BattleRoundResultResp, error) {
	logger := contextx.FromLogger(ctx)
	rb := contextx.FromRB(ctx)
	// 分布式锁
	m := rb.NewMutex(deskId)
	if err := m.Lock(ctx); err != nil {
		logger.Errorf("BattleLogic.GetRoundResult error:%s", err.Error())
		return nil, errors.NewResponseError(constant.ServerBusy, err)
	}

	defer func() {
		if _, err := m.Unlock(context.Background()); err != nil {
			logger.WithError(err).Error("error on mutex unlock")
		}
	}()

	resp := new(schema.BattleRoundResultResp)
	battleData, err := s.getBattleData(ctx, rb, deskId)
	// 检查错误
	if err != nil {
		if errors.Is(err, redis.Nil) {
			// 战斗已经结束，房间已经解散
			resp.State = 1
			return resp, nil
		}
		return nil, err
	}

	// 机器人逻辑
	if err := s.robotJoin(ctx, rb, battleData, true); err != nil {
		return nil, err
	}

	if battleData.StartAt == 0 {
		// 人数已满则直接开始
		battleConfig := s.tables.BattleConfigTb.Get(battleData.BattleId)
		if len(battleData.Players) == int(battleConfig.PlayerNum) {
			// 倒计时到了
			if time.Now().Unix() >= battleData.CreateAt+int64(battleConfig.MatchTime) {
				battleData.StartAt = battleData.CreateAt + int64(battleConfig.MatchTime)
			} else {
				battleData.StartAt = time.Now().Unix()
			}
			// 保存战斗数据
			if err := s.saveBattleData(ctx, rb, battleData); err != nil {
				logger.Errorf("BattleLogic.GetRoundResult error:%s", err.Error())
				return nil, err
			}
		} else {
			// 倒计时到了
			if time.Now().Unix() >= battleData.CreateAt+int64(battleConfig.MatchTime) {
				battleData.StartAt = battleData.CreateAt + int64(battleConfig.MatchTime)
				// 保存战斗数据
				if err := s.saveBattleData(ctx, rb, battleData); err != nil {
					logger.Errorf("BattleLogic.GetRoundResult error:%s", err.Error())
					return nil, err
				}
			}
		}
	}

	round := s.getRound(battleData)

	resp.Round = round
	if round == 0 {
		// 游戏还没开始
		resp.State = 2
		return resp, nil
	}

	logger.Infof("BattleLogic.GetRoundResult deskId=%s, round:%d", deskId, round)

	roundStartTime := s.getRoundStartTime(battleData, round)
	resp.RoundStartTime = roundStartTime

	n := time.Now().Unix() - roundStartTime
	if len(battleData.Result) < round && len(battleData.Players) > 0 {
		calcRound := 0
		if n >= 0 {
			calcRound = round
		} else if len(battleData.Result) < round-1 { // 上一回合
			calcRound = round - 1
		}

		logger.Infof("BattleLogic.GetRoundResult deskId=%s, n=%d,calcRound=%d", deskId, n, calcRound)

		if calcRound > 0 {
			for i := 1; i <= calcRound; i++ {
				s.checkBet(battleData, i)

				s.calcRoundResult(roleId, battleData, i)
			}
			logger.Infof("BattleLogic.GetRoundResult deskId=%s, result=%v", deskId, battleData.Result)
			// 保存战斗数据
			if err := s.saveBattleData(ctx, rb, battleData); err != nil {
				logger.Errorf("BattleLogic.GetRoundResult error:%s", err.Error())
				return nil, err
			}
		}
	}

	playerData := s.getPlayerData(battleData, roleId)
	isEnd := s.isBattleEnd(battleData, round)
	if isEnd || (battleData.SettlementRound > 0 && s.isRoundEnd(battleData, battleData.SettlementRound) || battleData.SettlementRound == -1) {
		// 战斗已经结束
		resp.State = 1
		// 结束的回合
		resp.Round = len(playerData.Bet)
	} else {
		if !utils.InArray(battleData.Players, roleId) {
			// 可以进行结算
			resp.State = 3
			// 结束的回合
			resp.Round = len(playerData.Bet)
		}
	}

	resp.SyncScore = s.getBattleSyncScoreResp(battleData, playerData, resp.Round)
	resp.PetId = playerData.PetId
	resp.LeftPlayerNum = len(battleData.Players)

	// 只发送到结束的回合
	for i := 0; i < len(battleData.Result); i++ {
		if resp.Round > i {
			resp.Result = append(resp.Result, battleData.Result[i])
		}
	}

	resp.BattleId = battleData.BattleId

	return resp, nil
}

func (s *battleLogic) getGridList(battleData *models.BattleData) []int32 {
	battleConfig := s.tables.BattleConfigTb.Get(battleData.BattleId)
	gridList := make([]int32, 0)
	var k int32
	for k = 1; k <= battleConfig.GridNum; k++ {
		if utils.IndexOf(battleData.Result, k) == -1 {
			gridList = append(gridList, k)
		}
	}
	return gridList
}

// 检查下注
func (s *battleLogic) checkBet(battleData *models.BattleData, round int) {
	gridList := s.getGridList(battleData)
	for _, roleId := range battleData.Players {
		playerData := s.getPlayerData(battleData, roleId)

		var grid int32
		// 没有下注，随机下注
		if len(playerData.Bet) < round {
			// 随机格子
			grid = gridList[s.xorshift.Intn(len(gridList))]
			// 保存下注信息
			s.bet(roleId, grid, round, battleData)
		}
	}
}

func (s *battleLogic) calcRoundResult(rid uint64, battleData *models.BattleData, round int) {
	// 本回合还没结算
	if len(battleData.Result) < round {
		gridList := s.getGridList(battleData)
		battleData.Result = append(battleData.Result, gridList[s.xorshift.Intn(len(gridList))]) // 杀死的格子
		battleConfig := s.tables.BattleConfigTb.Get(battleData.BattleId)
		n := rand.Int() % 8
		for i := 0; i < n; i++ {
			s.xorshift.Intn(len(gridList))
		}

		players := utils.DeepCopyArray(battleData.Players)
		for _, roleId := range players {
			playerData := s.getPlayerData(battleData, roleId)
			grid := playerData.Bet[round-1]

			// 战斗引导则一直胜利
			if battleConfig.IsGuide == 1 && roleId == rid {
				var k int32
				for k = 1; k <= battleConfig.GridNum; k++ {
					if utils.IndexOf(battleData.Result, k) == -1 && grid != k {
						battleData.Result[round-1] = k
						break
					}
				}
			}

			// 要杀的格子
			result := battleData.Result[round-1]

			// 胜利
			if grid != result {
				playerData.Win = 1
			} else {
				// 个人战斗结果保存
				playerData.Win = 0
				playerData.Bonus = 0
				playerData.Settlement = 1
				battleData.Players, _ = utils.RemoveElement(battleData.Players, roleId)
			}
		}

		bonus := s.calcPlayerBonus(battleData, round)
		for _, roleId := range players {
			playerData := s.getPlayerData(battleData, roleId)
			if playerData.Win == 1 {
				playerData.Bonus = bonus
			}
		}
		// 记录每回合奖金
		battleData.Bonus[round] = bonus

		if len(battleData.Players) == 0 || s.isAllRobot(battleData.Players) {
			battleData.SettlementRound = round
		}
	}
}

// 是否所有玩家都结算了
func (s *battleLogic) isAllPlayersSettlement(battleData *models.BattleData) bool {
	for roleId, playerData := range battleData.PlayerData {
		if !s.isRobot(roleId) && playerData.Settlement != 2 {
			return false
		}
	}
	return true
}

// Settlement 结算
func (s *battleLogic) Settlement(ctx context.Context, db *gorm.DB, roleId uint64, deskId string) (*schema.BattleSettlementResp, error) {
	logger := contextx.FromLogger(ctx)
	rb := contextx.FromRB(ctx)
	// 分布式锁
	m := rb.NewMutex(deskId)
	if err := m.Lock(ctx); err != nil {
		logger.Errorf("BattleLogic.Settlement error:%s", err.Error())
		return nil, errors.NewResponseError(constant.ServerBusy, err)
	}

	defer func() {
		if _, err := m.Unlock(context.Background()); err != nil {
			logger.WithError(err).Error("error on mutex unlock")
		}
	}()

	battleData, err := s.getBattleData(ctx, rb, deskId)
	// 检查错误
	if err != nil {
		if errors.Is(err, redis.Nil) {
			role, err := models.RoleRepo.Get(ctx, db, roleId)
			if err != nil {
				return nil, err
			}

			if err := models.RoleRepo.UpdateColum(ctx, db, roleId, "lastDeskId", ""); err != nil {
				return nil, err
			}

			// redis数据被清理则直接判断输
			if role.LastDeskId == deskId {
				resp := new(schema.BattleSettlementResp)
				resp.Win = 0
				resp.Bonus = 0
				return resp, nil
			}
			// 房间已经解散
			return nil, errors.NewResponseError(constant.BattleAlreadyDismiss, nil)
		}
		return nil, err
	}

	round := s.getRound(battleData)
	isEnd := s.isBattleEnd(battleData, round)
	if battleData.Settlement == 0 && (isEnd || (len(battleData.Players) == 0) ||
		(battleData.SettlementRound > 0 && s.isRoundEnd(battleData, battleData.SettlementRound)) || battleData.SettlementRound == -1) {
		// 进行结算
		if err := s.doSettlement(ctx, db, rb, battleData); err != nil {
			return nil, err
		}
	}

	// 获取个人战斗结果
	playerData := s.getPlayerData(battleData, roleId)

	resp := new(schema.BattleSettlementResp)
	resp.Win = playerData.Win
	resp.Bonus = playerData.Bonus

	battleId := cast.ToInt32(strings.Split(deskId, "-")[0])
	battleConfig := s.tables.BattleConfigTb.Get(battleId)

	// 还未结算进行结算
	if playerData.Settlement == 1 {
		playerData.Settlement = 2

		role, err := models.RoleRepo.Get(ctx, db, roleId)
		if err != nil {
			return nil, err
		}

		// 检查是否合法
		if role.LastDeskId != deskId {
			return nil, errors.NewResponseError(constant.BattleAlreadyDismiss, nil)
		}

		err = db.Transaction(func(db *gorm.DB) error {
			if err := models.RoleRepo.UpdateColum(ctx, db, roleId, "lastDeskId", ""); err != nil {
				return err
			}
			// 胜利奖励
			if playerData.Win == 1 {
				for _, rewardData := range battleConfig.Reward {
					reward, err := UtilsLogic.AddItem(ctx, db, roleId, rewardData.Id, rewardData.Num*playerData.Bonus, rewardData.Type)
					if err != nil {
						return err
					}
					resp.Reward = reward
				}
			}

			return nil
		})

		if err != nil {
			logger.Errorf("BattleLogic.Settlement error:%s", err.Error())
			return nil, err
		}

		// 所有人都结算了则清除redis中的战斗数据
		if s.isAllPlayersSettlement(battleData) {
			battleData.State = 2
			if err := rb.Client().Del(ctx, fmt.Sprintf(BattleDataKey, deskId)).Err(); err != nil {
				logger.Errorf("BattleLogic.Settlement error:%s", err.Error())
				return nil, err
			}
		} else {
			if err := s.saveBattleData(ctx, rb, battleData); err != nil {
				return nil, err
			}
		}
	}

	return resp, nil
}

func (s *battleLogic) doSettlement(ctx context.Context, db *gorm.DB, rb *redisbackend.RedisBackend, battleData *models.BattleData) error {
	if battleData.Settlement == 1 {
		return nil
	}
	battleData.Settlement = 1

	logger := contextx.FromLogger(ctx)

	// 最后胜利的人
	for _, roleId := range battleData.Players {
		playerData := s.getPlayerData(battleData, roleId)
		playerData.Settlement = 1
	}

	// 保存战斗数据
	if err := s.saveBattleData(ctx, rb, battleData); err != nil {
		logger.Errorf("BattleLogic.doSettlement error:%s", err.Error())
		return err
	}

	// 保存战斗结果到数据库
	battleResult := new(models.BattleResult)
	utils.Copy(battleResult, battleData)
	if err := models.BattleResultRepo.Save(ctx, db, battleResult); err != nil {
		logger.Errorf("BattleLogic.doSettlement error:%s", err.Error())
		return err
	}
	return nil
}
