package logic

import (
	"context"
	"eggServer/internal/constant"
	"eggServer/internal/contextx"
	cfg "eggServer/internal/gamedata"
	"eggServer/internal/models"
	"eggServer/internal/schema"
	"eggServer/pkg/errors"
	"eggServer/pkg/utils"
	"gorm.io/gorm"
	"time"
)

var SignLogic = new(signLogic)

type signLogic struct {
	tables *cfg.Tables
}

func (s *signLogic) Init(tables *cfg.Tables) {
	s.tables = tables
}

func (s *signLogic) Daily(ctx context.Context, db *gorm.DB, role *models.Role) (*schema.SignDataResp, *schema.RewardData, error) {
	now := time.Now().Unix()
	sign, err := models.SignRepo.Get(ctx, db, role.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, err
	}

	// 判断是否需要重置签到
	if sign.LoginTime == 0 || utils.IsDifferentDays(time.Unix(sign.LoginTime, 0), time.Unix(now, 0), "Asia/Shanghai") {
		signConfigList := s.tables.SignTb.GetDataList()
		if len(signConfigList)-1 == sign.LoginNum || errors.Is(err, gorm.ErrRecordNotFound) {
			sign.LoginNum = 0
			sign.Day = 0
			sign.ReSign = 0
			sign.StartTime = now
		}

		sign.RoleID = role.ID
		sign.LoginTime = now
		sign.IsSign = 0

		reSignConfigList := s.tables.ReSignTb.GetDataList()
		if sign.LoginNum < sign.Day+len(reSignConfigList)-sign.ReSign {
			sign.LoginNum = utils.CalculateDaysDifference(sign.StartTime, now)
		} else {
			sign.LoginNum = sign.Day + 1
		}

		if sign.LoginNum > len(signConfigList) {
			sign.LoginNum = len(signConfigList)
		}

		if sign.SignCardETime < now {
			sign.SignCardETime = 0
		}

		if err := models.SignRepo.Save(ctx, db, sign); err != nil {
			return nil, nil, err
		}
	}

	signDataResp, err := s.Data(ctx, db, role.ID)
	return signDataResp, nil, err
}

// SignIn 签到
func (s *signLogic) SignIn(ctx context.Context, db *gorm.DB, roleId uint64, id int32, reSignIn bool) (*schema.SignInResp, error) {
	logger := contextx.FromLogger(ctx)
	signConfig := s.tables.SignTb.Get(id)
	if signConfig == nil {
		logger.Errorf("SignLogic.SignIn id=%d not found", id)
		return nil, errors.NewResponseError(constant.ItemNotFound, errors.New("id not found"))
	}

	sign, err := models.SignRepo.Get(ctx, db, roleId)
	if err != nil {
		logger.Errorf("SignLogic.SignIn error:%s", err.Error())
		return nil, err
	}

	// 提前签到无效
	if int32(sign.LoginNum) < id || sign.Day > int(id) || (id > 1 && sign.LoginNum > 1 && sign.Day < int(id-1)) {
		return nil, errors.NewResponseError(constant.ParametersInvalid, nil)
	}

	resp := &schema.SignInResp{
		RewardList: make([]*schema.RewardData, 0),
	}

	if sign.IsSign == 1 {
		if reSignIn {
			reSignConfigList := s.tables.ReSignTb.GetDataList()
			if len(reSignConfigList) > sign.ReSign {
				reSignConfig := reSignConfigList[sign.ReSign]
				cost := reSignConfig.Cost
				item, err := UtilsLogic.AddItem(ctx, db, roleId, cost.Id, -cost.Num, cost.Type)
				if err != nil {
					return nil, err
				}
				resp.RewardList = append(resp.RewardList, item)
				sign.ReSign++
			} else {
				return nil, errors.NewResponseError(constant.ParametersInvalid, nil)
			}
		} else {
			return nil, errors.NewResponseError(constant.ParametersInvalid, nil)
		}
	} else {
		sign.IsSign = 1
		if int32(sign.Day)+1 != id {
			return nil, errors.NewResponseError(constant.ParametersInvalid, nil)
		}
	}

	sign.Day++

	// 签到卡奖励双倍
	rewardCount := 1
	if sign.SignCardETime > time.Now().Unix() {
		rewardCount = 2
	}

	for k := 0; k < rewardCount; k++ {
		reward, err := UtilsLogic.AddItem(ctx, db, roleId, signConfig.Reward.Id, signConfig.Reward.Num, signConfig.Reward.Type)
		if err != nil {
			return nil, err
		}
		resp.RewardList = append(resp.RewardList, reward)
	}

	sign.RoleID = roleId
	if err := models.SignRepo.Save(ctx, db, sign); err != nil {
		logger.Errorf("SignLogic.SignIn error:%s", err.Error())
		return nil, err
	}

	return resp, nil
}

// Data 签到数据
func (s *signLogic) Data(ctx context.Context, db *gorm.DB, roleId uint64) (*schema.SignDataResp, error) {
	logger := contextx.FromLogger(ctx)
	sign, err := models.SignRepo.Get(ctx, db, roleId)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Errorf("SignLogic.Data error:%s", err.Error())
		return nil, err
	}

	resp := new(schema.SignDataResp)
	utils.Copy(resp, sign)
	return resp, nil
}
