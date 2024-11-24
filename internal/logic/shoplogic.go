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
	"encoding/json"
	"fmt"
	"github.com/spf13/cast"
	"github.com/tonkeeper/tonapi-go"
	"gorm.io/gorm"
	"math/rand"
	"time"
)

var ShopLogic = new(shopLogic)

type shopLogic struct {
	tables          *cfg.Tables
	shopTotalWeight map[int32]int32
	shopWeightTb    map[int32][]*cfg.IDailyShop
}

func (s *shopLogic) Init(tables *cfg.Tables) {
	s.tables = tables
	s.shopTotalWeight = make(map[int32]int32)
	s.shopWeightTb = make(map[int32][]*cfg.IDailyShop)
	shopList := s.tables.DailyShopTb.GetDataList()
	for _, v := range shopList {
		s.shopTotalWeight[v.GroupId] = s.shopTotalWeight[v.GroupId] + v.Weight
		s.shopWeightTb[v.GroupId] = append(s.shopWeightTb[v.GroupId], v)
	}
}

func (s *shopLogic) Daily(ctx context.Context, db *gorm.DB, role *models.Role) (*schema.ShopRefreshResp, error) {
	return s.Refresh(ctx, db, role.ID, true)
}

func (s *shopLogic) calcRefresh(roleId uint64, shop *models.Shop, free bool) *models.Shop {
	shopList := make([]*cfg.IDailyShop, 0)
	for groupId, v := range s.shopTotalWeight {
		var n = rand.Int31n(v) + 1
		var temp int32 = 0
		tempList := s.shopWeightTb[groupId]
		for i := 0; i < len(tempList); i++ {
			if n > temp && n <= temp+tempList[i].Weight {
				shopList = append(shopList, tempList[i])
				break
			}
			temp = temp + tempList[i].Weight
		}
	}
	shop.RoleID = roleId
	shop.Discount = make([]int32, 0)
	shop.List = make([]int32, 0)
	for _, v := range shopList {
		shop.List = append(shop.List, v.Id)
		if v.Discount[1] == v.Discount[0] {
			shop.Discount = append(shop.Discount, v.Discount[0])
		} else {
			shop.Discount = append(shop.Discount, (rand.Int31n((v.Discount[1]-v.Discount[0])/10+1)+v.Discount[0]/10)*10)
		}
	}
	shop.Buy = make([]int, len(shopList))
	shop.Share = make(map[int][]int)

	if free {
		shop.RefreshTimes = 0
	} else {
		shop.RefreshTimes = shop.RefreshTimes + 1
	}

	return shop
}

func (s *shopLogic) Refresh(ctx context.Context, db *gorm.DB, roleId uint64, free bool) (*schema.ShopRefreshResp, error) {
	logger := contextx.FromLogger(ctx)
	resp := new(schema.ShopRefreshResp)
	shopRefresh := s.tables.GlobalTb.GetDataList()[0].ShopRefresh

	shop, err := models.ShopRepo.Get(ctx, db, roleId)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Errorf("ShopLogic.Refresh error:%s", err.Error())
		return nil, err
	}

	if (len(shopRefresh) > int(shop.RefreshTimes) && !free) || free {
		shop = s.calcRefresh(roleId, shop, free)
		err := db.Transaction(func(db *gorm.DB) error {
			if !free {
				item := shopRefresh[shop.RefreshTimes-1]
				reward, err := UtilsLogic.AddItem(ctx, db, roleId, item.Id, -item.Num, item.Type)
				if err != nil {
					return err
				}
				resp.Cost = reward
			}

			if err := models.ShopRepo.Save(ctx, db, shop); err != nil {
				logger.Errorf("ShopLogic.Refresh error:%s", err.Error())
				return err
			}

			return nil
		})

		if err == nil {
			resp.Shop = new(schema.ShopDataResp)
			utils.Copy(&resp.Shop, shop)
		}
		return resp, err
	}

	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *shopLogic) Data(ctx context.Context, db *gorm.DB, roleId uint64) (*schema.ShopDataResp, error) {
	resp := new(schema.ShopDataResp)
	shop, err := models.ShopRepo.Get(ctx, db, roleId)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if shop.List == nil {
		shopRefreshResp, err := s.Refresh(ctx, db, roleId, true)
		if err != nil {
			return nil, err
		}
		resp = shopRefreshResp.Shop
	} else {
		utils.Copy(resp, shop)
	}

	return resp, nil
}

func (s *shopLogic) dealBuy(ctx context.Context, db *gorm.DB, roleId uint64, price *cfg.GlobalItemData, item *cfg.GlobalItemData) (*schema.ShopBuyResp, error) {
	logger := contextx.FromLogger(ctx)
	resp := new(schema.ShopBuyResp)

	err := db.Transaction(func(db *gorm.DB) error {
		if price.Id > 0 {
			reward, err := UtilsLogic.AddItem(ctx, db, roleId, price.Id, -price.Num, price.Type)
			if err != nil {
				return err
			}
			resp.Cost = reward
		}

		itemConfig := s.tables.ItemTb.Get(item.Id)
		if itemConfig == nil {
			logger.Errorf("ShopLogic.dealBuy itemId=%d item not found", item.Id)
			return errors.New("item not found")
		}

		if itemConfig.UseType == 1 {
			for k := 0; k < int(item.Num); k++ {
				reward, err := UtilsLogic.AddItem(ctx, db, roleId, item.Id, 1, item.Type)
				if err != nil {
					return err
				}
				if reward != nil {
					resp.RewardList = append(resp.RewardList, reward)
				}
			}
		} else {
			reward, err := UtilsLogic.AddItem(ctx, db, roleId, item.Id, item.Num, item.Type)
			if err != nil {
				return err
			}
			if reward != nil {
				resp.RewardList = append(resp.RewardList, reward)
			}
		}

		return nil
	})
	if err != nil {
		logger.Errorf("ShopLogic.dealBuy error: %s", err.Error())
		return nil, err
	}

	return resp, nil
}

func (s *shopLogic) dailyBuy(ctx context.Context, db *gorm.DB, roleId uint64, id int32) (*schema.ShopBuyResp, error) {
	logger := contextx.FromLogger(ctx)
	shopConfig := s.tables.DailyShopTb.Get(id)
	if shopConfig == nil {
		logger.Errorf("ShopLogic.dailyBuy %d id not found", id)
		return nil, errors.NewResponseError(constant.ShopNotFound, nil)
	}

	shop, err := models.ShopRepo.Get(ctx, db, roleId)
	if err != nil {
		return nil, err
	}
	list := shop.List
	index := utils.IndexOf(list, id)
	buy := shop.Buy

	if buy[index] >= int(shopConfig.BuyLimit) {
		return nil, errors.NewResponseError(constant.PurchaseRestrictions, nil)
	}

	buy[index] = buy[index] + 1
	resp := new(schema.ShopBuyResp)
	err = db.Transaction(func(db *gorm.DB) error {
		if shopConfig.Price.Id > 0 {
			price := shopConfig.Price
			reward, err := UtilsLogic.AddItem(ctx, db, roleId, price.Id, -price.Num*shop.Discount[index]/100, price.Type)
			if err != nil {
				return err
			}
			resp.Cost = reward
		}

		// 检查分享获得商品条件
		if shopConfig.Share == 1 {
			shopIdx := utils.IndexOf(shop.List, id)
			if shopIdx == -1 {
				return errors.NewResponseError(constant.ParametersInvalid, nil)
			}
			if !utils.InArray(shop.Share[cfg.ShopType_Daily], shopIdx) {
				return errors.NewResponseError(constant.PurchaseRestrictions, nil)
			}
		}

		item := shopConfig.Item
		itemConfig := s.tables.ItemTb.Get(item.Id)
		if itemConfig == nil {
			logger.Errorf("ShopLogic.dailyBuy itemId=%d item not found", item.Id)
			return errors.New("item not found")
		}

		if itemConfig.UseType == 1 {
			for k := 0; k < int(item.Num); k++ {
				reward, err := UtilsLogic.AddItem(ctx, db, roleId, item.Id, 1, item.Type)
				if err != nil {
					return err
				}
				if reward != nil {
					resp.RewardList = append(resp.RewardList, reward)
				}
			}
		} else {
			reward, err := UtilsLogic.AddItem(ctx, db, roleId, item.Id, item.Num, item.Type)
			if err != nil {
				return err
			}
			if reward != nil {
				resp.RewardList = append(resp.RewardList, reward)
			}
		}

		if err := models.ShopRepo.Save(ctx, db, shop); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		logger.Errorf("ShopLogic.dailyBuy error: %s", err.Error())
		return nil, err
	}

	return resp, nil
}

func (s *shopLogic) ordinaryBuy(ctx context.Context, db *gorm.DB, roleId uint64, id int32) (*schema.ShopBuyResp, error) {
	logger := contextx.FromLogger(ctx)
	shopConfig := s.tables.OrdinaryShopTb.Get(id)
	if shopConfig == nil {
		logger.Errorf("ShopLogic.ordinaryBuy %d id not found", id)
		return nil, errors.NewResponseError(constant.ShopNotFound, nil)
	}

	return s.dealBuy(ctx, db, roleId, shopConfig.Price, shopConfig.Item)
}

func (s *shopLogic) specialBuy(ctx context.Context, db *gorm.DB, roleId uint64, id int32) (*schema.ShopBuyResp, error) {
	logger := contextx.FromLogger(ctx)
	shopConfig := s.tables.SpecialShopTb.Get(id)
	if shopConfig == nil {
		logger.Errorf("ShopLogic.specialBuy %d id not found", id)
		return nil, errors.NewResponseError(constant.ShopNotFound, nil)
	}

	return s.dealBuy(ctx, db, roleId, shopConfig.Price, shopConfig.Item)
}

func (s *shopLogic) Buy(ctx context.Context, db *gorm.DB, roleId uint64, req *schema.ShopBuyReq) (*schema.ShopBuyResp, error) {
	if req.ShopType == cfg.ShopType_Daily {
		return s.dailyBuy(ctx, db, roleId, req.ID)
	} else if req.ShopType == cfg.ShopType_Ordinary {
		return s.ordinaryBuy(ctx, db, roleId, req.ID)
	} else if req.ShopType == cfg.ShopType_Special {
		return s.specialBuy(ctx, db, roleId, req.ID)
	}
	return nil, errors.NewResponseError(constant.ParametersInvalid, nil)
}

// CreateOrder 创建订单
func (s *shopLogic) CreateOrder(ctx context.Context, db *gorm.DB, roleId uint64, userId uint64, req *schema.ShopCreateOrderReq) (*schema.ShopCreateOrderResp, error) {
	logger := contextx.FromLogger(ctx)

	var price float64
	if req.ShopType == cfg.ShopType_Diamond {
		diamondShopConfig := s.tables.DiamondShopTb.Get(req.ID)
		price = float64(diamondShopConfig.Price.Num) / 100.0
	} else if req.ShopType == cfg.ShopType_Special {
		specialShopConfig := s.tables.SpecialShopTb.Get(req.ID)
		price = float64(specialShopConfig.Price.Num) / 100.0
	} else {
		return nil, errors.NewResponseError(constant.ParametersInvalid, nil)
	}
	order := new(models.Order)

	if req.Currency == constant.TON {
		ton, err := TonapiLogic.Ticker(ctx, "TON-USDT")
		if err != nil {
			return nil, err
		}
		price = cast.ToFloat64(fmt.Sprintf("%.6f", price*(1/ton)))
		order.TotalAmount = int64(price * 1000000000)
	} else {
		order.TotalAmount = int64(price * 1000000)
	}

	user, err := models.UserRepo.FindOneByUserId(ctx, db, userId)
	if err != nil {
		logger.Errorf("ShopLogic.CreateOrder error:%s", err.Error())
		return nil, err
	}

	order.OrderId = utils.GenerateID()
	order.RoleID = roleId
	order.Platform = user.Platform
	order.CreatedAt = time.Now().Unix()
	order.Currency = req.Currency
	order.OrderStatus = 0
	order.ShopId = req.ID
	order.ShopType = req.ShopType

	if err := models.OrderRepo.Create(ctx, db, order); err != nil {
		return nil, err
	}

	resp := new(schema.ShopCreateOrderResp)
	resp.Price = order.TotalAmount
	resp.OrderId = cast.ToString(order.OrderId)
	fmt.Printf("订单id: %s\n", resp.OrderId)

	return resp, nil
}

// Delivery 发货
func (s *shopLogic) Delivery(ctx context.Context, db *gorm.DB, roleId uint64, req *schema.ShopDeliveryReq) (*schema.ShopDeliveryResp, error) {
	logger := contextx.FromLogger(ctx)
	client := TonapiLogic.TonApi()
	transaction, err := client.GetBlockchainTransaction(ctx, tonapi.GetBlockchainTransactionParams{TransactionID: req.TransactionID})
	if err != nil {
		return nil, errors.NewResponseError(constant.UnknownError, err)
	}

	if transaction.Success && transaction.EndStatus == "active" {
		transactionTime := transaction.Utime
		sender := transaction.Account.Address
		receiver := transaction.OutMsgs[0].Destination.Value.Address
		amount := transaction.OutMsgs[0].Value
		success := transaction.Success
		opCode := transaction.OutMsgs[0].OpCode.Value
		var orderId int64

		if opCode == "0x0f8a7ea5" { // jetton转账
			type Body struct {
				Amount              string `json:"amount"`
				Destination         string `json:"destination"`
				ResponseDestination string `json:"response_destination"`
				ForwardPayload      struct {
					Value struct {
						Value struct {
							Text string `json:"text"`
						} `json:"value"`
					} `json:"value"`
				} `json:"forward_payload"`
			}

			var body Body
			if err := json.Unmarshal(transaction.OutMsgs[0].DecodedBody, &body); err != nil {
				logger.Errorf("ShopLogic.Delivery error: %s", err.Error())
				return nil, errors.NewResponseError(constant.UnknownError, err)
			}

			jettonWalletAddress := TonapiLogic.GetJettonWalletAddress(ctx, sender)
			if jettonWalletAddress != receiver { // 检查每个用户存款地址，防止欺诈
				return nil, errors.NewResponseError(constant.UnknownError, errors.New("jettonWalletAddress is not match"))
			}

			orderId = cast.ToInt64(body.ForwardPayload.Value.Value.Text)
			amount = cast.ToInt64(body.Amount)
			receiver = body.Destination
		} else if opCode == "0x00000000" { // ton转账
			type Body struct {
				Text string `json:"text"`
			}

			var body Body
			if err := json.Unmarshal(transaction.OutMsgs[0].DecodedBody, &body); err != nil {
				logger.Errorf("ShopLogic.Delivery error: %s", err.Error())
				return nil, errors.NewResponseError(constant.UnknownError, err)
			}

			orderId = cast.ToInt64(body.Text)
		} else {
			return nil, errors.NewResponseError(constant.UnknownError, errors.New("opCode Invalid"))
		}

		fmt.Printf("交易时间：%d 发送者: %s 接收者: %s 发送了: %d TON备注信息: %s 是否成功: %v\n", transactionTime, sender, receiver, amount, transaction.OutMsgs[0].DecodedBody.String(), success)

		if receiver == config.C.WalletAddress {
			order, err := models.OrderRepo.Get(ctx, db, orderId)
			if err != nil {
				logger.Errorf("ShopLogic.Delivery error: %s", err.Error())
				return nil, err
			}

			fmt.Printf("订单状态：%d 订单金额: %d 实付金额: %d 订单创建时间: %d 交易时间: %d\n", order.OrderStatus, order.TotalAmount, amount, order.CreatedAt, transactionTime)
			if order.OrderStatus == 0 && order.TotalAmount == amount && order.CreatedAt <= transactionTime {
				resp := new(schema.ShopDeliveryResp)

				var item *cfg.GlobalItemData
				if order.ShopType == cfg.ShopType_Diamond {
					diamondShopConfig := s.tables.DiamondShopTb.Get(order.ShopId)
					item = diamondShopConfig.Item
				} else if order.ShopType == cfg.ShopType_Special {
					specialShopConfig := s.tables.SpecialShopTb.Get(order.ShopId)
					item = specialShopConfig.Item
				} else {
					return nil, errors.NewResponseError(constant.ParametersInvalid, nil)
				}

				err = db.Transaction(func(db *gorm.DB) error {
					itemConfig := s.tables.ItemTb.Get(item.Id)
					if itemConfig == nil {
						logger.Errorf("ShopLogic.Delivery itemId=%d item not found", item.Id)
						return errors.New("item not found")
					}

					if itemConfig.UseType == 1 {
						for k := 0; k < int(item.Num); k++ {
							reward, err := UtilsLogic.AddItem(ctx, db, roleId, item.Id, 1, item.Type)
							if err != nil {
								return err
							}
							if reward != nil {
								resp.RewardList = append(resp.RewardList, reward)
							}
						}
					} else {
						reward, err := UtilsLogic.AddItem(ctx, db, roleId, item.Id, item.Num, item.Type)
						if err != nil {
							return err
						}
						if reward != nil {
							resp.RewardList = append(resp.RewardList, reward)
						}
					}

					if err := models.OrderRepo.Updates(ctx, db, orderId, map[string]interface{}{"orderStatus": 1, "updatedAt": time.Now().Unix()}); err != nil {
						return err
					}
					return nil
				})

				if err != nil {
					logger.Errorf("ShopLogic.Delivery error: %s", err.Error())
					return nil, err
				}

				resp.ShopType = order.ShopType
				resp.ShopId = order.ShopId

				return resp, nil
			}
		} else {

		}
	}

	return nil, errors.NewResponseError(constant.UnknownError, err)
}

func (s *shopLogic) Share(ctx context.Context, db *gorm.DB, userId uint64, req *schema.ShopShareReq) (*schema.ShopShareResp, error) {
	logger := contextx.FromLogger(ctx)

	user, err := models.UserRepo.FindOneByUserId(ctx, db, userId)
	if err != nil {
		logger.Errorf("ShopLogic.Share error: %s", err.Error())
		return nil, err
	}

	shareFriendsUrl := s.tables.GlobalTb.GetDataList()[0].ShareFriendsUrl
	now := time.Now()
	startParam := fmt.Sprintf("%s%03d%02d%02d%d%d", req.UserUid, user.Platform, now.Month(), now.Day(), req.ShopType, req.ShopIdx) // 启动参数

	resp := new(schema.ShopShareResp)
	resp.Param = fmt.Sprintf("%s?startapp=%s", shareFriendsUrl, utils.Encrypt(startParam))

	return resp, nil
}

func (s *shopLogic) RecordShare(ctx context.Context, db *gorm.DB, roleId uint64, shopType int, shopIdx int) error {
	logger := contextx.FromLogger(ctx)

	shop, err := models.ShopRepo.Get(ctx, db, roleId)
	if err != nil {
		logger.Errorf("ShopLogic.RecordShare error: %s", err.Error())
		return err
	}

	if shopType != cfg.ShopType_Daily || len(shop.List)-1 < shopIdx {
		return nil
	}

	shopConfig := s.tables.DailyShopTb.Get(shop.List[shopIdx])
	if shopConfig.Share != 1 {
		return nil
	}

	if shop.Share == nil {
		shop.Share = make(map[int][]int)
	}
	if !utils.InArray(shop.Share[shopType], shopIdx) {
		shop.Share[shopType] = append(shop.Share[shopType], shopIdx)
	}

	if err := models.ShopRepo.Save(ctx, db, shop); err != nil {
		logger.Errorf("ShopLogic.RecordShare error: %s", err.Error())
		return err
	}
	return err
}
