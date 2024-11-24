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
	"github.com/spf13/cast"
	"gorm.io/gorm"
	"time"
)

var ItemLogic = new(itemLogic)

type itemLogic struct {
	tables *cfg.Tables
}

func (s *itemLogic) Init(tables *cfg.Tables) {
	s.tables = tables
}

func (s *itemLogic) AddItem(ctx context.Context, db *gorm.DB, roleId uint64, itemId int32, count int32) (*schema.RewardData, error) {
	logger := contextx.FromLogger(ctx)
	itemConfig := s.tables.ItemTb.Get(itemId)
	if itemConfig == nil {
		logger.Errorf("ItemLogic.AddItem itemId=%d item not found", itemId)
		return nil, errors.NewResponseError(constant.ItemNotFound, errors.New("item not found"))
	}

	if itemConfig.UseType == 1 {
		return s.UseItem(ctx, db, roleId, itemId, itemConfig.Param, false)
	}

	// 更新道具数量
	_, err := s.UpdateItemCount(ctx, db, roleId, itemId, count)
	if err != nil {
		return nil, err
	}

	resp := new(schema.RewardData)
	itemData := new(schema.ItemData)
	itemData.ID = itemId
	itemData.Num = count
	itemData.Type = cfg.RewardType_Item
	resp.Item = itemData
	return resp, nil
}

func (s *itemLogic) UpdateItemCount(ctx context.Context, db *gorm.DB, roleId uint64, itemId int32, count int32) (int32, error) {
	logger := contextx.FromLogger(ctx)

	item, err := models.ItemRepo.Get(ctx, db, roleId, itemId)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Errorf("ItemLogic.AddItem itemId=%d error: %s", itemId, err.Error())
		return 0, err
	}

	itemNum := item.ItemNum
	item.RoleID = roleId
	item.ItemID = itemId
	item.ItemNum = item.ItemNum + count

	if item.ItemNum >= 0 {
		if err := models.ItemRepo.Save(ctx, db, item); err != nil {
			logger.Errorf("ItemLogic.AddItem itemId=%d error: %s", item.ItemID, err.Error())
			return item.ItemNum, err
		}
	} else {
		if itemId == constant.Diamond {
			return 0, errors.NewResponseError(constant.DiamondNotEnough, errors.New("not enough diamonds"))
		}
		return 0, errors.NewResponseError(constant.ItemNotEnough, errors.New("not enough items"))
	}

	// 体力小于最大恢复值
	if item.ItemID == constant.VIT {
		globalConfig := s.tables.GlobalTb.GetDataList()[0]
		if itemNum >= globalConfig.VitMaxAuto && count < 0 && item.ItemNum < globalConfig.VitMaxAuto {
			lastAddVitTime := time.Now().Unix()
			if err := models.RoleRepo.UpdateColum(ctx, db, roleId, "lastAddVitTime", lastAddVitTime); err != nil {
				logger.Errorf("VitLogic.AutoAddVit error: %s", err.Error())
				return 0, err
			}
		}
	}
	return 0, nil
}

func (s *itemLogic) UseItem(ctx context.Context, db *gorm.DB, roleId uint64, itemId int32, param string, checkCount bool) (*schema.RewardData, error) {
	logger := contextx.FromLogger(ctx)
	itemConfig := s.tables.ItemTb.Get(itemId)
	if itemConfig == nil {
		logger.Errorf("ItemLogic.UseItem itemId=%d item not found", itemId)
		return nil, errors.NewResponseError(constant.ItemNotFound, errors.New("item not found"))
	}

	if checkCount {
		// 扣除道具
		_, err := s.UpdateItemCount(ctx, db, roleId, itemId, -1)
		if err != nil {
			return nil, err
		}
	}

	if itemConfig.Cmd == "returnEgg" {
		egg, err := models.EggRepo.Get(ctx, db, roleId, cast.ToInt32(param))
		if err != nil {
			logger.Errorf("ItemLogic.UseItem itemId=%d error:%s", itemId, err.Error())
			return nil, err
		}

		if err := models.EggRepo.UpdateColum(ctx, db, egg.ID, "isDel", 0); err != nil {
			logger.Errorf("ItemLogic.UseItem itemId=%d error:%s", itemId, err.Error())
			return nil, err
		}

		e := new(schema.Egg)
		utils.Copy(e, egg)
		resp := new(schema.RewardData)
		resp.Egg = e
		return resp, nil
	} else if itemConfig.Cmd == "openBox" {
		return EggLogic.EggOpenByIndex(ctx, db, roleId, cast.ToInt(itemConfig.Param)-1)
	} else if itemConfig.Cmd == "autoEggCollect" {
		role, err := models.RoleRepo.Get(ctx, db, roleId)
		if err != nil {
			return nil, err
		}
		if role.AutoEggCollectRTime == 0 {
			role.AutoEggCollectETime = time.Now().Unix()
		}
		role.AutoEggCollectETime = role.AutoEggCollectETime + cast.ToInt64(itemConfig.Param)
		role.AutoEggCollectRTime = role.AutoEggCollectRTime + cast.ToInt64(itemConfig.Param)
		if err := models.RoleRepo.Updates(ctx, db, roleId, map[string]interface{}{"autoEggCollectETime": role.AutoEggCollectETime, "autoEggCollectRTime": role.AutoEggCollectRTime}); err != nil {
			return nil, err
		}
		return nil, nil
	} else if itemConfig.Cmd == "passPortDeluxeReward" {
		if err := models.RoleRepo.Updates(ctx, db, roleId, map[string]interface{}{"passPortDeluxeReward": 1}); err != nil {
			return nil, err
		}
		return nil, nil
	} else if itemConfig.Cmd == "signCard" {
		if err := models.SignRepo.Updates(ctx, db, roleId, map[string]interface{}{"signCardETime": time.Now().Unix() + cast.ToInt64(itemConfig.Param)}); err != nil {
			return nil, err
		}
		return nil, nil
	}

	return nil, errors.NewResponseError(constant.ParametersInvalid, nil)
}
