package logic

import (
	"context"
	"eggServer/internal/constant"
	"eggServer/internal/contextx"
	cfg "eggServer/internal/gamedata"
	"eggServer/internal/models"
	"eggServer/internal/schema"
	"eggServer/pkg/errors"
	"gorm.io/gorm"
	"time"
)

var PassPortLogic = new(passPortLogic)

type passPortLogic struct {
	tables *cfg.Tables
}

func (s *passPortLogic) Init(tables *cfg.Tables) {
	s.tables = tables
}

func (s *passPortLogic) getIssue() int32 {
	passPortIssueList := s.tables.PassPortIssueTb.GetDataList()
	now := time.Now().Unix()
	for _, passPortIssue := range passPortIssueList {
		if int64(passPortIssue.StartTime) <= now && now <= int64(passPortIssue.EndTime) {
			return passPortIssue.Issue
		}
	}
	return 0
}

func (s *passPortLogic) Get(ctx context.Context, db *gorm.DB, roleId uint64, id int32, deluxe bool) (*schema.PassPortRewardGetResp, error) {
	logger := contextx.FromLogger(ctx)
	passPortRewardConfig := s.tables.PassPortRewardTb.Get(id)
	if passPortRewardConfig == nil {
		logger.Errorf("PassPortLogic.Get id=%d not found", id)
		return nil, errors.NewResponseError(constant.ParametersInvalid, errors.New("id not found"))
	}

	if passPortRewardConfig.Issue != s.getIssue() {
		return nil, errors.NewResponseError(constant.ParametersInvalid, errors.New("issue not match"))
	}

	resp := new(schema.PassPortRewardGetResp)
	role, err := models.RoleRepo.Get(ctx, db, roleId)
	if err != nil {
		logger.Errorf("PassPortLogic.Get error:%s", err.Error())
		return nil, err
	}

	if passPortRewardConfig.Exp <= role.BattleCount {
		passPort, err := models.PassPortRepo.Get(ctx, db, roleId, id)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Errorf("PassPortLogic.Get error:%s", err.Error())
			return nil, err
		}

		passPort.PassPortId = id
		passPort.RoleID = roleId

		if deluxe && role.PassPortDeluxeReward == 1 {
			if passPort.State != 2 && passPort.State != 3 {
				resp.Reward, err = UtilsLogic.AddItem(ctx, db, roleId, passPortRewardConfig.DeluxeReward.Id, passPortRewardConfig.DeluxeReward.Num, passPortRewardConfig.DeluxeReward.Type)
				if err != nil {
					return nil, err
				}

				if passPort.State == 1 {
					passPort.State = 3
				} else {
					passPort.State = 2
				}

				if err := models.PassPortRepo.Save(ctx, db, passPort); err != nil {
					logger.Errorf("PassPortLogic.Get error:%s", err.Error())
					return nil, err
				}
			}
		} else {
			if passPort.State != 1 && passPort.State != 3 {
				resp.Reward, err = UtilsLogic.AddItem(ctx, db, roleId, passPortRewardConfig.OrdinaryReward.Id, passPortRewardConfig.OrdinaryReward.Num, passPortRewardConfig.OrdinaryReward.Type)
				if err != nil {
					return nil, err
				}

				if passPort.State == 2 {
					passPort.State = 3
				} else {
					passPort.State = 1
				}

				if err := models.PassPortRepo.Save(ctx, db, passPort); err != nil {
					logger.Errorf("PassPortLogic.Get error:%s", err.Error())
					return nil, err
				}
			}
		}
	} else {
		return nil, errors.NewResponseError(constant.ParametersInvalid, errors.New("exp not match"))
	}
	return resp, nil
}

func (s *passPortLogic) Data(ctx context.Context, db *gorm.DB, roleId uint64) (*schema.PassPortRewardDataResp, error) {
	logger := contextx.FromLogger(ctx)

	passPorts, err := models.PassPortRepo.FindAllByRoleId(ctx, db, roleId)
	if err != nil {
		logger.Errorf("PassPortLogic.Data error:%s", err.Error())
		return nil, err
	}

	resp := new(schema.PassPortRewardDataResp)
	resp.Data = make(map[int32]byte)
	for _, passPort := range passPorts {
		resp.Data[passPort.PassPortId] = passPort.State
	}
	return resp, nil
}
