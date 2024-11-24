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
)

var GuideLogic = new(guideLogic)

type guideLogic struct {
	tables *cfg.Tables
}

func (s *guideLogic) Init(tables *cfg.Tables) {
	s.tables = tables
}

func (s *guideLogic) Step(ctx context.Context, db *gorm.DB, roleId uint64, guideId int32) error {
	logger := contextx.FromLogger(ctx)
	guideConfig := s.tables.GuideListTb.Get(guideId)
	if guideConfig == nil {
		logger.Errorf("GuideLogic.Step step=%d guide not found", guideId)
		return errors.NewResponseError(constant.ParametersInvalid, nil)
	}

	guide, err := models.GuideRepo.Get(ctx, db, roleId)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Errorf("GuideLogic.Step error:%s", err.Error())
		return err
	}

	if !utils.InArray(guide.Step, guideConfig.Id) && guideConfig.IsSave == 1 {
		guide.Step = append(guide.Step, guideConfig.Id)
		if err := models.GuideRepo.Save(ctx, db, roleId, guide); err != nil {
			logger.Errorf("GuideLogic.Step error:%s", err.Error())
			return err
		}
	}

	return nil
}

func (s *guideLogic) GuideData(ctx context.Context, db *gorm.DB, roleId uint64) (*schema.GuideDataResp, error) {
	logger := contextx.FromLogger(ctx)
	resp := new(schema.GuideDataResp)
	guide, err := models.GuideRepo.Get(ctx, db, roleId)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Errorf("GuideLogic.GuideData error:%s", err.Error())
		return nil, err
	}

	resp.Step = make([]int32, 0)
	utils.Copy(&resp.Step, guide.Step)

	return resp, nil
}
