package logic

import (
	"context"
	"eggServer/internal/constant"
	cfg "eggServer/internal/gamedata"
	"eggServer/internal/schema"
	"eggServer/pkg/errors"
	"gorm.io/gorm"
)

type utilsLogic struct {
}

var UtilsLogic = new(utilsLogic)

func (s *utilsLogic) Init() {

}

// AddItem 添加物品
func (s *utilsLogic) AddItem(ctx context.Context, db *gorm.DB, roleId uint64, itemId int32, count int32, itemType int32) (*schema.RewardData, error) {
	if itemType == cfg.RewardType_Item {
		return ItemLogic.AddItem(ctx, db, roleId, itemId, count)
	} else if itemType == cfg.RewardType_Pet {
		return PetLogic.AddPet(ctx, db, roleId, itemId, count)
	}

	return nil, errors.NewResponseError(constant.ParametersInvalid, nil)
}
