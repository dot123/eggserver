package logic

import (
	"context"
	"eggServer/internal/config"
	"eggServer/internal/constant"
	"eggServer/internal/contextx"
	cfg "eggServer/internal/gamedata"
	"eggServer/internal/models"
	"eggServer/internal/schema"
	"eggServer/pkg/errors"
	"eggServer/pkg/utils"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/cast"
	"gorm.io/gorm"
	"strings"
	"time"
)

// userLogic 处理用户相关的操作
type userLogic struct {
	tables *cfg.Tables // 游戏配置数据
}

// UserLogic 是 userLogic 的全局实例
var UserLogic = new(userLogic)

// Init 初始化 userLogic，设置游戏表数据
func (s *userLogic) Init(tables *cfg.Tables) {
	s.tables = tables
}

// Login 处理用户登录，验证令牌并在必要时创建新用户
func (s *userLogic) Login(ctx context.Context, db *gorm.DB, req *schema.LoginReq) (string, error) {
	logger := contextx.FromLogger(ctx)
	rb := contextx.FromRB(ctx)
	// 分布式锁
	mutex := rb.NewMutex(req.UserUid)
	if err := mutex.Lock(ctx); err != nil {
		logger.Errorf("UserLogic.Login error:%s", err.Error())
		return "", errors.NewResponseError(constant.ServerBusy, err)
	}

	defer func() {
		if _, err := mutex.Unlock(context.Background()); err != nil {
			logger.WithError(err).Error("error on mutex unlock")
		}
	}()

	userUid := req.UserUid

	if userUid == "" {
		return "", errors.NewResponseError(constant.ParametersInvalid, nil)
	}

	var inviterUserUid string
	var p, m, d, shopType, shopIdx int

	// 携带启动参数
	if req.StartParam != "" {
		num, err := utils.Decrypt(req.StartParam)
		if err != nil {
			logger.Errorf("UserLogic.Login startParam:%s error: %s", req.StartParam, err.Error())
		} else {
			startParam := cast.ToString(num)
			strLen := len(startParam)
			if strLen > 0 && strLen <= 3 {
				p = utils.ToInt(startParam)
			} else if strLen >= 18 {
				split := strLen - 9 // 平台标识长度3+时间月长度2日长度2+商品类型1商品索引1
				inviterUserUid = startParam[:split]
				p = utils.ToInt(startParam[split : split+3])
				m = utils.ToInt(startParam[split+3 : split+5])
				d = utils.ToInt(startParam[split+5 : split+7])
				shopType = utils.ToInt(startParam[split+7 : split+8])
				shopIdx = utils.ToInt(startParam[split+8 : split+9])
				logger.Infof("startParam:%s-%d-%d-%d-%d-%d", inviterUserUid, p, m, d, shopType, shopIdx)
			}
		}
	}

	isNew := false
	// 查找或创建用户
	user, err := models.UserRepo.FindOneByUserUid(ctx, db, userUid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			user.UserUid = userUid
			user.FirstName = req.FirstName
			user.LastName = req.LastName
			user.PhotoUrl = req.PhotoUrl
			user.UserName = req.UserName
			user.LanguageCode = req.LanguageCode
			user.OS = req.OS
			user.Platform = p
			if err := models.UserRepo.Create(ctx, db, user); err != nil {
				logger.Errorf("UserLogic.Login error: %s", err.Error())
				return "", errors.NewResponseError(constant.TokenGenerateFail, err)
			}
			isNew = true
		} else {
			return "", err
		}
	} else {
		if err := models.UserRepo.Updates(ctx, db, userUid, map[string]interface{}{
			"userName":     req.UserName,
			"firstName":    req.FirstName,
			"lastName":     req.LastName,
			"photoUrl":     req.PhotoUrl,
			"languageCode": req.LanguageCode,
			"os":           req.OS,
		}); err != nil {
			logger.Errorf("UserLogic.Login error: %s", err.Error())
			return "", errors.NewResponseError(constant.TokenGenerateFail, err)
		}
	}

	if inviterUserUid != "" {
		inviter, err := models.UserRepo.FindOneByUserUid(ctx, db, inviterUserUid) // 邀请人
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Errorf("UserLogic.Login error: %s", err.Error())
		} else {
			if inviter.ID > 0 && inviter.ID != user.ID {
				if isNew {
					s.checkInvite(ctx, db, inviter.ID, user.ID, shopType, shopIdx)
				} else {
					s.checkShare(ctx, db, inviter.ID, user.ID, shopType, shopIdx, m, d)
				}
			}
		}
	}

	// 生成 JWT 令牌
	return s.generateToken(ctx, user.ID, 0)
}

// 检查邀请
func (s *userLogic) checkInvite(ctx context.Context, db *gorm.DB, inviterId uint64, invitedId uint64, shopType int, shopIdx int) {
	logger := contextx.FromLogger(ctx)

	role, err := models.RoleRepo.FindOneByUserId(ctx, db, inviterId) // 查找邀请人roleId
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Errorf("UserLogic.checkInvite error: %s", err.Error())
		return
	}

	if role.ID > 0 {
		_, err := models.InviteRepo.FindOne(ctx, db, inviterId, invitedId)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Errorf("UserLogic.checkInvite error: %s", err.Error())
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) { // 没有被邀请
			err = db.Transaction(func(db *gorm.DB) error {
				// 更新邀请人任务数据
				if err := TaskLogic.RecordTaskProgress(ctx, db, role.ID, 2, 1); err != nil {
					return err
				}
				if err := TaskLogic.RecordTaskProgress(ctx, db, role.ID, 8, 1); err != nil {
					return err
				}

				// 记录商店分享
				if err := ShopLogic.RecordShare(ctx, db, role.ID, shopType, shopIdx); err != nil {
					return err
				}

				// 被邀请的人获得奖励
				globalConfig := s.tables.GlobalTb.GetDataList()[0]
				_, err := UtilsLogic.AddItem(ctx, db, invitedId, globalConfig.InvitedReward.Id, globalConfig.InvitedReward.Num, globalConfig.InvitedReward.Type)
				if err != nil {
					return err
				}

				invite := new(models.Invite)
				invite.Inviter = inviterId
				invite.Invited = invitedId
				invite.InvitedLastLogin = time.Now().Unix()
				return models.InviteRepo.Create(ctx, db, invite) // 记录被邀请人
			})
			if err != nil {
				logger.Errorf("UserLogic.checkInvite error: %s", err.Error())
				return
			}
		}
	}
}

// 检查分享
func (s *userLogic) checkShare(ctx context.Context, db *gorm.DB, inviterId uint64, invitedId uint64, shopType int, shopIdx int, m int, d int) {
	logger := contextx.FromLogger(ctx)
	invite, err := models.InviteRepo.Get(ctx, db, inviterId, invitedId)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Errorf("UserLogic.checkShare error: %s", err.Error())
		return
	}

	t1 := time.Now()
	if int(t1.Month()) == m && t1.Day() == d { // 链接创建的时间
		t2 := time.Unix(invite.InvitedLastLogin, 0)
		if utils.IsDifferentDays(t1, t2, "Asia/Shanghai") { // 上次通过链接进入的时间
			role, err := models.RoleRepo.FindOneByUserId(ctx, db, inviterId) // 查找邀请人roleId
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Errorf("UserLogic.checkShare error: %s", err.Error())
				return
			}

			err = db.Transaction(func(db *gorm.DB) error {
				// 更新邀请人任务数据
				if err := TaskLogic.RecordTaskProgress(ctx, db, role.ID, 8, 1); err != nil {
					return err
				}

				// 记录商店分享
				if err := ShopLogic.RecordShare(ctx, db, role.ID, shopType, shopIdx); err != nil {
					return err
				}

				// 被邀请的人获得奖励
				globalConfig := s.tables.GlobalTb.GetDataList()[0]
				_, err := UtilsLogic.AddItem(ctx, db, invitedId, globalConfig.InvitedReward.Id, globalConfig.InvitedReward.Num, globalConfig.InvitedReward.Type)
				if err != nil {
					return err
				}

				invite.Inviter = inviterId
				invite.Invited = invitedId
				invite.InvitedLastLogin = t1.Unix()            // 更新被邀请登录时间
				return models.InviteRepo.Save(ctx, db, invite) // 记录被邀请人
			})
			if err != nil {
				logger.Errorf("UserLogic.checkShare error: %s", err.Error())
				return
			}
		}
	}
}

// Enter 处理用户进入游戏的请求
func (s *userLogic) Enter(ctx context.Context, db *gorm.DB, userId uint64) (*schema.RoleDataResp, error) {
	logger := contextx.FromLogger(ctx)
	rb := contextx.FromRB(ctx)
	// 分布式锁
	mutex := rb.NewMutex(fmt.Sprintf("%d", userId))
	if err := mutex.Lock(ctx); err != nil {
		logger.Errorf("UserLogic.Enter error:%s", err.Error())
		return nil, errors.NewResponseError(constant.ServerBusy, err)
	}

	defer func() {
		if _, err := mutex.Unlock(context.Background()); err != nil {
			logger.WithError(err).Error("error on mutex unlock")
		}
	}()

	now := time.Now().Unix()
	role, err := models.RoleRepo.FindOneByUserId(ctx, db, userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			role.UserID = userId
			role.CreatedAt = now
			role.LastLayEggTime = now
			role.LastAddVitTime = now
			if err := db.Create(role).Error; err != nil {
				logger.Errorf("UserLogic.Enter error: %s", err.Error())
				return nil, err
			}
			// 初始化角色的默认值
			s.initAddValue(ctx, db, role.ID)
		} else {
			logger.Errorf("UserLogic.Enter error: %s", err.Error())
			return nil, err
		}
	}

	// 执行每日任务
	if role.LastDailyTime == 0 || utils.IsDifferentDays(time.Unix(role.LastDailyTime, 0), time.Unix(now, 0), "Asia/Shanghai") {
		role.LastDailyTime = now
		_, err := s.Daily(ctx, db, role)
		if err != nil {
			logger.Errorf("UserLogic.Enter error: %s", err.Error())
			return nil, err
		}
	}

	_, _, err = SignLogic.Daily(ctx, db, role)
	if err != nil {
		logger.Errorf("UserLogic.Enter error: %s", err.Error())
		return nil, err
	}

	role.LastLogin = now
	if err := models.RoleRepo.Updates(ctx, db, role.ID, map[string]interface{}{"lastDailyTime": role.LastDailyTime,
		"lastLogin": role.LastLogin}); err != nil {
		logger.Errorf("UserLogic.Enter error: %s", err.Error())
		return nil, err
	}

	// 玩家进入时的回调
	s.onEnter(ctx, db, role.ID)

	// 准备响应数据
	resp := new(schema.RoleDataResp)
	utils.Copy(resp, role)
	resp.RoleID = role.ID
	resp.ServerTime = now
	resp.LastLoginTime = role.LastLogin

	// 生成 JWT 令牌
	tokenString, err := s.generateToken(ctx, userId, role.ID)
	if err != nil {
		return nil, err
	}
	resp.Token = tokenString

	// 自动增加体力
	autoAddVitResp, err := s.AutoAddVit(ctx, db, role.ID, role.LastAddVitTime)
	if err != nil {
		return nil, err
	}
	role.LastAddVitTime = autoAddVitResp.LastAddVitTime

	// 获取物品、宠物、任务和蛋
	items, err := models.ItemRepo.FindAllByRoleId(ctx, db, role.ID)
	if err != nil {
		return nil, err
	}

	pets, err := models.PetRepo.FindAllByRoleId(ctx, db, role.ID)
	if err != nil {
		return nil, err
	}

	eggs, err := models.EggRepo.FindAllByRoleId(ctx, db, role.ID)
	if err != nil {
		return nil, err
	}

	taskDataResp, err := TaskLogic.Data(ctx, db, role.ID)
	if err != nil {
		return nil, err
	}

	shopDataResp, err := ShopLogic.Data(ctx, db, role.ID)
	if err != nil {
		return nil, err
	}

	signDataResp, err := SignLogic.Data(ctx, db, role.ID)
	if err != nil {
		return nil, err
	}

	passPortRewardDataResp, err := PassPortLogic.Data(ctx, db, role.ID)
	if err != nil {
		return nil, err
	}

	// 填充响应数据
	utils.Copy(&resp.Items, items)
	utils.Copy(&resp.Pets, pets)
	utils.Copy(&resp.Eggs, eggs)

	resp.Task = taskDataResp
	resp.Shop = shopDataResp
	resp.Sign = signDataResp
	resp.PassPortReward = passPortRewardDataResp

	guideDataResp, err := GuideLogic.GuideData(ctx, db, role.ID)
	if err != nil {
		return nil, err
	}
	resp.Guide = guideDataResp

	return resp, nil
}

// initAddValue 初始化角色的默认值
func (s *userLogic) initAddValue(ctx context.Context, db *gorm.DB, roleId uint64) {
	logger := contextx.FromLogger(ctx)
	globalConfig := s.tables.GlobalTb.GetDataList()[0]
	initVit := globalConfig.InitVit
	initGold := globalConfig.InitGold
	initDiamond := globalConfig.InitDiamond
	if initVit >= 0 {
		_, err := ItemLogic.AddItem(ctx, db, roleId, constant.VIT, initVit)
		if err != nil {
			logger.Errorf("UserLogic.initAddValue error: %s", err.Error())
		}
	}

	if initGold >= 0 {
		_, err := ItemLogic.AddItem(ctx, db, roleId, constant.Gold, initGold)
		if err != nil {
			logger.Errorf("UserLogic.initAddValue error: %s", err.Error())
		}
	}

	if initDiamond >= 0 {
		_, err := ItemLogic.AddItem(ctx, db, roleId, constant.Diamond, initDiamond)
		if err != nil {
			logger.Errorf("UserLogic.initAddValue error: %s", err.Error())
		}
	}
}

// generateToken 创建 JWT 令牌
func (s *userLogic) generateToken(ctx context.Context, userId uint64, roleId uint64) (string, error) {
	logger := contextx.FromLogger(ctx)
	rb := contextx.FromRB(ctx)
	var sessionId int64
	if config.C.JWTAuth.UseSession {
		id, err := rb.Client().Incr(ctx, "sessionId").Result()
		if err != nil {
			logger.Errorf("Error generating ID:%s", err.Error())
			return "", err
		}

		// 将令牌存储到 Redis
		if err := rb.Client().Set(ctx, cast.ToString(userId), id, time.Hour*config.C.JWTAuth.Expired).Err(); err != nil {
			return "", err
		}
		sessionId = id
	}

	// 创建包含声明的 JWT 令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId":    userId,
		"roleId":    roleId,
		"sessionId": sessionId,
		"exp":       time.Now().Add(time.Hour * config.C.JWTAuth.Expired).Unix(),
	})

	// 签名令牌并返回字符串
	tokenString, err := token.SignedString([]byte(config.C.JWTAuth.Key))
	if err != nil {
		logger.Errorf("UserLogic.generateToken error: %s", err.Error())
		return "", errors.NewResponseError(constant.TokenGenerateFail, err)
	}

	return tokenString, nil
}

// Daily 执行每日任务
func (s *userLogic) Daily(ctx context.Context, db *gorm.DB, role *models.Role) (*schema.DailyResp, error) {
	logger := contextx.FromLogger(ctx)
	resp := new(schema.DailyResp)

	if err := TaskLogic.Daily(ctx, db, role); err != nil {
		return nil, err
	}

	shopRefreshResp, err := ShopLogic.Daily(ctx, db, role)
	if err != nil {
		logger.Errorf("UserLogic.Daily error: %s", err.Error())
		return resp, nil
	}
	resp.Shop = shopRefreshResp.Shop

	signDataResp, reward, err := SignLogic.Daily(ctx, db, role)
	if err != nil {
		logger.Errorf("UserLogic.Daily error: %s", err.Error())
		return resp, nil
	}
	resp.Sign = signDataResp
	resp.Reward = reward
	resp.ServerTime = time.Now().Unix()
	return resp, nil
}

// AutoAddVit 自动增加体力
func (s *userLogic) AutoAddVit(ctx context.Context, db *gorm.DB, roleId uint64, lastAddVitTime int64) (*schema.AutoAddVitResp, error) {
	logger := contextx.FromLogger(ctx)
	globalConfig := s.tables.GlobalTb.GetDataList()[0]
	resp := new(schema.AutoAddVitResp)
	resp.LastAddVitTime = lastAddVitTime
	now := time.Now().Unix()
	// 计算自上次增加体力以来的时间
	t := now - lastAddVitTime
	num := int32(t / int64(globalConfig.VitAddAuto))
	if num > 0 {
		item, err := models.ItemRepo.Get(ctx, db, roleId, constant.VIT)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Errorf("VitLogic.AutoAddVit itemId=%d error: %s", constant.VIT, err.Error())
			return resp, err
		}

		// 自动回复体力最大值
		vitAddAuto := globalConfig.VitMaxAuto
		if item.ItemNum >= vitAddAuto {
			resp.LastAddVitTime = now
			return resp, nil
		} else {
			if item.ItemNum+num > vitAddAuto {
				num = vitAddAuto - item.ItemNum
			}
		}

		// 最大体力区间
		if item.ItemNum > globalConfig.VitMinMax[1] {
			return resp, nil
		}

		// 在事务中增加体力
		lastAddVitTime = now
		err = db.Transaction(func(db *gorm.DB) error {
			if err := models.RoleRepo.UpdateColum(ctx, db, roleId, "lastAddVitTime", lastAddVitTime); err != nil {
				logger.Errorf("VitLogic.AutoAddVit error: %s", err.Error())
				return err
			}
			reward, err := ItemLogic.AddItem(ctx, db, roleId, constant.VIT, num)
			resp.Reward = reward
			return err
		})

		if err != nil {
			resp.LastAddVitTime = lastAddVitTime
			return resp, err
		}
	}
	return resp, nil
}

// GM 执行 GM 指令
func (s *userLogic) GM(ctx context.Context, db *gorm.DB, roleId uint64, cmdStr string) (*schema.GMResp, error) {
	logger := contextx.FromLogger(ctx)
	strArr := strings.Split(cmdStr, " ")
	if len(strArr) < 1 {
		return nil, errors.NewResponseError(constant.ParametersInvalid, errors.New("cmd format error"))
	}

	if !config.C.GM {
		return nil, errors.NewResponseError(constant.PermissionDenied, nil)
	}

	// 检查角色权限
	role, err := models.RoleRepo.Get(ctx, db, roleId)
	if err != nil {
		logger.Errorf("GmLogic.GM error:%s", err.Error())
		return nil, err
	}

	if role.GM != 1 {
		return nil, errors.NewResponseError(constant.PermissionDenied, nil)
	}

	resp := new(schema.GMResp)

	// 处理 GM 指令
	if cast.ToInt64(strArr[0]) != 0 {
		roleId = cast.ToUint64(strArr[0])
		strArr = strArr[1:]
		_, err := models.RoleRepo.Get(ctx, db, roleId)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.NewResponseError(constant.ParametersInvalid, nil)
			}
			return nil, err
		}
	}
	cmd := strArr[0]
	if cmd == "addItem" {
		itemId := cast.ToInt32(strArr[1])
		itemNum := cast.ToInt32(strArr[2])
		reward, err := ItemLogic.AddItem(ctx, db, roleId, itemId, itemNum)
		if err != nil {
			return nil, err
		}
		if reward != nil {
			resp.RewardList = append(resp.RewardList, reward)
		}
	} else if cmd == "addEgg" {
		eggNum := cast.ToInt32(strArr[1])
		if eggNum > 0 {
			var i int32
			for i = 0; i < eggNum; i++ {
				egg, err := EggLogic.AddEgg(ctx, db, roleId) // 添加蛋
				if err != nil {
					logger.Errorf("GmLogic.GM error:%s", err.Error())
				}

				e := new(schema.Egg)
				utils.Copy(e, egg)
				rewardData := new(schema.RewardData)
				rewardData.Egg = e
				resp.RewardList = append(resp.RewardList, rewardData)
			}
		}
	} else if cmd == "restGuide" {
		if err := models.GuideRepo.Delete(ctx, db, roleId); err != nil {
			return nil, err
		}
	} else if cmd == "userCount" {
		var userCount int64
		oneHourAgo := time.Now().Add(-1 * time.Hour).Unix()

		// 查询最近一小时注册的用户数量
		db.Model(new(models.User)).
			Where("createdAt >= ?", oneHourAgo).
			Count(&userCount)

		return nil, errors.NewResponseError(constant.RegisteredUsers, errors.New(fmt.Sprintf("%d", userCount)))
	} else if cmd == "addBattleCount" {
		battleCount := cast.ToInt32(strArr[1])
		if err := models.RoleRepo.Updates(ctx, db, roleId, map[string]interface{}{"battleCount": role.BattleCount + battleCount}); err != nil {
			return nil, err
		}
		resp.BattleCount = battleCount
	} else {
		return nil, errors.NewResponseError(constant.ParametersInvalid, nil)
	}
	return resp, nil
}

// SyncData 同步数据
func (s *userLogic) SyncData(ctx context.Context, db *gorm.DB, roleId uint64) (*schema.SyncDataResp, error) {
	taskDataResp, err := TaskLogic.Data(ctx, db, roleId)
	if err != nil {
		return nil, err
	}

	shopDataResp, err := ShopLogic.Data(ctx, db, roleId)
	if err != nil {
		return nil, err
	}

	resp := new(schema.SyncDataResp)
	resp.ServerTime = time.Now().Unix()
	resp.Task = taskDataResp
	resp.Shop = shopDataResp

	return resp, nil
}

// onEnter 玩家进入时的回调
func (s *userLogic) onEnter(ctx context.Context, db *gorm.DB, roleId uint64) {

}
