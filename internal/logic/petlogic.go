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
)

var PetLogic = new(petLogic)

type petLogic struct {
	tables *cfg.Tables
}

func (s *petLogic) Init(tables *cfg.Tables) {
	s.tables = tables
}

func (s *petLogic) AddPet(ctx context.Context, db *gorm.DB, roleId uint64, petId int32, count int32) (*schema.RewardData, error) {
	logger := contextx.FromLogger(ctx)
	if petConfig := s.tables.PetTb.Get(petId); petConfig == nil {
		logger.Errorf("PetLogic.AddPet petId=%d pet not found", petId)
		return nil, errors.NewResponseError(constant.MinionsNotFound, errors.New("minions not found"))
	}

	pet, err := models.PetRepo.Get(ctx, db, roleId, petId)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Errorf("PetLogic.AddPet petId=%d error: %s", petId, err.Error())
			return nil, err
		}
	}

	pet.RoleID = roleId
	pet.PetID = petId
	pet.PetNum = pet.PetNum + count

	if pet.PetNum >= 0 {
		if err := models.PetRepo.Save(ctx, db, pet); err != nil {
			logger.Errorf("PetLogic.AddPet petId=%d error: %s", pet.PetID, err.Error())
			return nil, err
		}
		if count > 0 {
			// 记录收集宠物进度
			if err := TaskLogic.RecordTaskProgress(ctx, db, roleId, 4, count); err != nil {
				return nil, err
			}
		}
	} else {
		return nil, errors.NewResponseError(constant.MinionsNotEnough, errors.New("minions not enough"))
	}

	resp := new(schema.RewardData)
	itemData := new(schema.ItemData)
	itemData.ID = petId
	itemData.Num = count
	itemData.Type = cfg.RewardType_Pet
	resp.Item = itemData
	return resp, nil
}
