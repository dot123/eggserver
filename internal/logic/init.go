package logic

import cfg "eggServer/internal/gamedata"

func Init(tables *cfg.Tables) {
	TonapiLogic.Init()
	UtilsLogic.Init()
	UserLogic.Init(tables)
	GuideLogic.Init(tables)
	LeaderboardLogic.Init(tables)
	ItemLogic.Init(tables)
	PetLogic.Init(tables)
	TaskLogic.Init(tables)
	EggLogic.Init(tables)
	BattleLogic.Init(tables)
	ShopLogic.Init(tables)
	SignLogic.Init(tables)
	PassPortLogic.Init(tables)
}
