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
	"math/rand"
	"time"
)

var EggLogic = new(eggLogic)

type eggLogic struct {
	tables             *cfg.Tables
	eggPartTotalWeight map[int32]int32
	eggPartTb          map[int32]map[int32][]*cfg.IEgg
	eggOpenWeightTb    map[int32][]*cfg.IEggOpenWeight
	eggOpenTotalWeight map[int32]int32
}

func (s *eggLogic) Init(tables *cfg.Tables) {
	s.tables = tables

	s.eggPartTotalWeight = make(map[int32]int32)
	s.eggPartTb = make(map[int32]map[int32][]*cfg.IEgg)
	s.eggOpenWeightTb = make(map[int32][]*cfg.IEggOpenWeight)
	s.eggOpenTotalWeight = make(map[int32]int32)

	s.eggPartTb[cfg.EggPartType_Part1] = make(map[int32][]*cfg.IEgg)
	s.eggPartTb[cfg.EggPartType_Part2] = make(map[int32][]*cfg.IEgg)
	s.eggPartTb[cfg.EggPartType_Part3] = make(map[int32][]*cfg.IEgg)

	eggList := tables.EggTb.GetDataList()
	for _, v := range eggList {
		s.eggPartTb[v.PartType][v.Quality] = append(s.eggPartTb[v.PartType][v.Quality], v)
	}

	eggPart1List := tables.EggPart1Tb.GetDataList()
	for _, v := range eggPart1List {
		s.eggPartTotalWeight[cfg.EggPartType_Part1] = s.eggPartTotalWeight[cfg.EggPartType_Part1] + v.Weight
	}

	eggPart2List := tables.EggPart1Tb.GetDataList()
	for _, v := range eggPart2List {
		s.eggPartTotalWeight[cfg.EggPartType_Part2] = s.eggPartTotalWeight[cfg.EggPartType_Part2] + v.Weight
	}

	eggPart3List := tables.EggPart1Tb.GetDataList()
	for _, v := range eggPart3List {
		s.eggPartTotalWeight[cfg.EggPartType_Part3] = s.eggPartTotalWeight[cfg.EggPartType_Part3] + v.Weight
	}

	eggOpenWeightList := tables.EggOpenWeightTb.GetDataList()
	for _, v := range eggOpenWeightList {
		s.eggOpenWeightTb[v.GroupId] = append(s.eggOpenWeightTb[v.GroupId], v)
		s.eggOpenTotalWeight[v.GroupId] = s.eggOpenTotalWeight[v.GroupId] + v.Weight
	}
}

func (s *eggLogic) ClickScreen(ctx context.Context, db *gorm.DB, roleId uint64, req *schema.ClickScreenReq) (*schema.ClickScreenResp, error) {
	logger := contextx.FromLogger(ctx)
	globalConfig := s.tables.GlobalTb.GetDataList()[0]
	resp := new(schema.ClickScreenResp)

	role, err := models.RoleRepo.Get(ctx, db, roleId)
	if err != nil {
		logger.Errorf("EggLogic.ClickScreen error: %s", err.Error())
		return resp, err
	}

	lastClickCount := role.LastClickCount

	// 超过一次请求最大点击次数
	if req.ClickCount > globalConfig.ClickLayEgg {
		req.ClickCount = globalConfig.ClickLayEgg
	}

	// 点击次数没有变化
	if lastClickCount > req.ClickCount {
		resp.LastClickCount = lastClickCount
		return resp, err
	}

	clickCount := req.ClickCount - lastClickCount // 增加的点击次数
	item, err := models.ItemRepo.Get(ctx, db, roleId, constant.VIT)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Errorf("EggLogic.ClickScreen itemId=%d error: %s", constant.VIT, err.Error())
		return nil, err
	}
	num := item.ItemNum / globalConfig.ClickTakeOffVit
	if clickCount > num {
		clickCount = num
	}

	// 扣除体力
	cost := globalConfig.ClickTakeOffVit * clickCount
	lastClickCount = lastClickCount + clickCount
	eggNum := lastClickCount / globalConfig.ClickLayEgg
	lastClickCount = lastClickCount - eggNum*globalConfig.ClickLayEgg
	resp.LastClickCount = lastClickCount

	err = db.Transaction(func(db *gorm.DB) error {
		if err := models.RoleRepo.UpdateColum(ctx, db, roleId, "lastClickCount", lastClickCount); err != nil {
			logger.Errorf("EggLogic.ClickScreen error: %s", err.Error())
			return err
		}

		if cost > 0 {
			reward, err := ItemLogic.AddItem(ctx, db, roleId, constant.VIT, -cost)
			if err != nil {
				return err
			}
			resp.Reward = reward
		}

		if eggNum > 0 {
			var i int32
			for i = 0; i < eggNum; i++ {
				egg, err := s.AddEgg(ctx, db, roleId) // 添加蛋
				if err != nil {
					return err
				}

				e := new(schema.Egg)
				utils.Copy(e, egg)

				resp.Reward.Egg = e
			}
		}

		return nil
	})

	return resp, err
}

// TryLayEgg 自动产蛋
func (s *eggLogic) TryLayEgg(ctx context.Context, db *gorm.DB, roleId uint64) (*schema.AutoLayEggResp, error) {
	logger := contextx.FromLogger(ctx)
	globalConfig := s.tables.GlobalTb.GetDataList()[0]

	role, err := models.RoleRepo.Get(ctx, db, roleId)
	if err != nil {
		return nil, err
	}

	lastLayEggTime := role.LastLayEggTime
	resp := new(schema.AutoLayEggResp)
	resp.LastLayEggTime = lastLayEggTime

	now := time.Now().Unix()
	t := now - lastLayEggTime

	layEggTime := int64(globalConfig.LayEggTime)
	if t > layEggTime {
		num := t / layEggTime

		if role.AutoEggCollectRTime > 0 { // 自动收蛋的剩余时间
			if role.AutoEggCollectRTime >= t {
				role.AutoEggCollectRTime = role.AutoEggCollectRTime - t
			} else {
				num = role.AutoEggCollectRTime / layEggTime
				role.AutoEggCollectRTime = 0
				role.AutoEggCollectETime = 0
			}
			n := (now - role.LastLayEggTime) / layEggTime
			role.LastLayEggTime = role.LastLayEggTime + n*layEggTime
		} else {
			if num > 0 {
				num = 1
			}
			role.LastLayEggTime = now
		}

		resp.LastLayEggTime = role.LastLayEggTime

		err = db.Transaction(func(db *gorm.DB) error {
			if err := models.RoleRepo.Updates(ctx, db, roleId, map[string]interface{}{"lastLayEggTime": role.LastLayEggTime, "autoEggCollectRTime": role.AutoEggCollectRTime, "autoEggCollectETime": role.AutoEggCollectETime}); err != nil {
				logger.Errorf("EggLogic.TryLayEgg error: %s", err.Error())
				return err
			}

			for i := 0; i < int(num); i++ {
				egg, err := s.AddEgg(ctx, db, roleId) // 添加蛋
				if err != nil {
					return err
				}
				e := new(schema.Egg)
				utils.Copy(e, egg)
				resp.Eggs = append(resp.Eggs, e)
			}

			return nil
		})

		if err != nil {
			return nil, err
		}
	}
	resp.AutoEggCollectRTime = role.AutoEggCollectRTime
	resp.AutoEggCollectETime = role.AutoEggCollectETime
	return resp, nil
}

func (s *eggLogic) randOneGggPart(part int32) int32 {
	var temp int32 = 0
	var quality int32 = 0
	var n = rand.Int31n(s.eggPartTotalWeight[part]) + 1

	if part == cfg.EggPartType_Part1 {
		temp = 0
		tempList := s.tables.EggPart1Tb.GetDataList()
		for i := 0; i < len(tempList); i++ {
			if n > temp && n <= temp+tempList[i].Weight {
				quality = tempList[i].Quality
				break
			}
			temp = temp + tempList[i].Weight
		}
	} else if part == cfg.EggPartType_Part2 {
		temp = 0
		tempList := s.tables.EggPart2Tb.GetDataList()
		for i := 0; i < len(tempList); i++ {
			if n > temp && n <= temp+tempList[i].Weight {
				quality = tempList[i].Quality
				break
			}
			temp = temp + tempList[i].Weight
		}
	} else if part == cfg.EggPartType_Part3 {
		temp = 0
		tempList := s.tables.EggPart3Tb.GetDataList()
		for i := 0; i < len(tempList); i++ {
			if n > temp && n <= temp+tempList[i].Weight {
				quality = tempList[i].Quality
				break
			}
			temp = temp + tempList[i].Weight
		}
	}

	tempList := s.eggPartTb[part]
	eggList := tempList[quality]

	eggPart, err := utils.RandomElement(eggList)
	if err != nil {
		return 0
	}
	return eggPart.Id
}

func (s *eggLogic) AddEgg(ctx context.Context, db *gorm.DB, roleId uint64) (*models.Egg, error) {
	egg := new(models.Egg)
	egg.RoleID = roleId
	egg.Part1 = s.randOneGggPart(cfg.EggPartType_Part1)
	egg.Part2 = s.randOneGggPart(cfg.EggPartType_Part2)
	egg.Part3 = s.randOneGggPart(cfg.EggPartType_Part3)

	if err := models.EggRepo.Create(ctx, db, egg); err != nil {
		return egg, err
	}
	// 记录蛋进度
	if err := TaskLogic.RecordTaskProgress(ctx, db, roleId, 3, 1); err != nil {
		return egg, err
	}
	return egg, nil
}

func (s *eggLogic) EggOpen(ctx context.Context, db *gorm.DB, roleId uint64, eggId int32, guideId int32) (*schema.RewardData, error) {
	logger := contextx.FromLogger(ctx)
	egg, err := models.EggRepo.Get(ctx, db, roleId, eggId)
	if err != nil {
		return nil, err
	}

	// 蛋评分
	score := s.tables.EggTb.Get(egg.Part1).EggScore +
		s.tables.EggTb.Get(egg.Part2).EggScore +
		s.tables.EggTb.Get(egg.Part3).EggScore

	eggScoreList := s.tables.EggScoreTb.GetDataList()

	// 从评分中选择
	var index = -1
	for i := 0; i < len(eggScoreList); i++ {
		if len(eggScoreList[i].EggScore) == 2 && eggScoreList[i].EggScore[0] < score && score <= eggScoreList[i].EggScore[1] {
			index = i
			break
		}
	}

	// 在引导中特殊处理砸蛋
	if guideId == 202 || guideId == 206 {
		guide, err := models.GuideRepo.Get(ctx, db, roleId)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if !utils.InArray(guide.Step, guideId) {
			index = 10 // 索引
		}
	}
	var resp *schema.RewardData

	if index == -1 {
		// 如果是空蛋则标记删除，可以使用挽回卡重新获得该蛋
		if err := models.EggRepo.UpdateColum(ctx, db, egg.ID, "isDel", 1); err != nil {
			logger.Errorf("EggLogic.EggOpen error: %s", err.Error())
			return nil, err
		}
		// 空蛋
		return new(schema.RewardData), nil
	}

	itemType, id, num := s.eggOpenWeight(index)

	if id != 0 {
		err := db.Transaction(func(db *gorm.DB) error {
			if itemType == cfg.RewardType_Pet {
				reward, err := PetLogic.AddPet(ctx, db, roleId, id, num)
				if err != nil {
					return err
				}
				resp = reward
			} else if itemType == cfg.RewardType_Item {
				reward, err := ItemLogic.AddItem(ctx, db, roleId, id, num)
				if err != nil {
					return err
				}
				resp = reward
			}

			if err := models.EggRepo.Delete(ctx, db, egg); err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			logger.Errorf("EggLogic.EggOpen error: %s", err.Error())
			return nil, err
		}
		return resp, nil
	} else {
		// 如果是空蛋则标记删除，可以使用挽回卡重新获得该蛋
		if err := models.EggRepo.UpdateColum(ctx, db, egg.ID, "isDel", 1); err != nil {
			logger.Errorf("EggLogic.EggOpen error: %s", err.Error())
			return nil, err
		}
	}

	// 空蛋
	return new(schema.RewardData), nil
}

func (s *eggLogic) EggOpenByIndex(ctx context.Context, db *gorm.DB, roleId uint64, index int) (*schema.RewardData, error) {
	logger := contextx.FromLogger(ctx)

	var resp *schema.RewardData

	itemType, id, num := s.eggOpenWeight(index)
	if id != 0 {
		err := db.Transaction(func(db *gorm.DB) error {
			if itemType == cfg.RewardType_Pet {
				reward, err := PetLogic.AddPet(ctx, db, roleId, id, num)
				if err != nil {
					return err
				}
				resp = reward
			} else if itemType == cfg.RewardType_Item {
				reward, err := ItemLogic.AddItem(ctx, db, roleId, id, num)
				if err != nil {
					return err
				}
				resp = reward
			}
			return nil
		})

		if err != nil {
			logger.Errorf("EggLogic.EggOpen error: %s", err.Error())
			return nil, err
		}
		return resp, nil
	}

	// 空蛋
	return new(schema.RewardData), nil
}

func (s *eggLogic) eggOpenWeight(index int) (int32, int32, int32) {
	eggScoreList := s.tables.EggScoreTb.GetDataList()
	// 总的组概率
	groupPb := eggScoreList[index].GroupPb
	var totalGroupPb int32 = 0
	for i := 0; i < len(groupPb); i++ {
		totalGroupPb = totalGroupPb + groupPb[i]
	}

	// 选出组id
	var groupId int32
	var temp int32 = 0
	var n = rand.Int31n(totalGroupPb) + 1
	for i := 0; i < len(groupPb); i++ {
		if n > temp && n <= temp+groupPb[i] {
			groupId = eggScoreList[index].OpenGroupId[i]
			break
		}
		temp = temp + groupPb[i]
	}

	if groupId != 0 {
		list := s.eggOpenWeightTb[groupId]
		var tempData *cfg.IEggOpenWeight
		var temp int32 = 0
		n := rand.Int31n(s.eggOpenTotalWeight[groupId]) + 1
		for i := 0; i < len(list); i++ {
			if n > temp && n <= temp+list[i].Weight {
				tempData = list[i]
				break
			}
			temp = temp + list[i].Weight
		}

		if tempData != nil && tempData.ArticleId != 0 {
			return tempData.ItemType, tempData.ArticleId, tempData.ArticleNum
		}
	}
	return 0, 0, 0
}
