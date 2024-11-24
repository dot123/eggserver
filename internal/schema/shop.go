package schema

type ShopDataResp struct {
	Discount     []int32       `json:"discount" msgpack:"discount"`
	List         []int32       `json:"list" msgpack:"list"`
	Buy          []int         `json:"buy" msgpack:"buy"`
	Share        map[int][]int `json:"share" msgpack:"share"`
	RefreshTimes byte          `json:"refreshTimes" msgpack:"refreshTimes"`
}

type ShopRefreshResp struct {
	Shop *ShopDataResp `json:"shop" msgpack:"shop"`
	Cost *RewardData   `json:"cost" msgpack:"cost"`
}

type ShopBuyReq struct {
	ID       int32 `json:"id" msgpack:"id" binding:"required"`
	ShopType byte  `json:"shopType" msgpack:"shopType" binding:"required"`
}

type ShopBuyResp struct {
	RewardList []*RewardData `json:"rewards" msgpack:"rewards"`
	Cost       *RewardData   `json:"cost" msgpack:"cost"`
}

type ShopCreateOrderReq struct {
	ID       int32 `json:"id" msgpack:"id" binding:"required"`
	ShopType byte  `json:"shopType" msgpack:"shopType" binding:"required"`
	Currency int   `json:"currency" msgpack:"currency" binding:"required"`
}

type ShopCreateOrderResp struct {
	OrderId string `json:"orderId" msgpack:"orderId"`
	Price   int64  `json:"price" msgpack:"price"`
}

type ShopDeliveryReq struct {
	TransactionID string `json:"transactionId" msgpack:"transactionId"`
}

type ShopDeliveryResp struct {
	ShopType   byte          `json:"shopType" msgpack:"shopType"`
	ShopId     int32         `json:"shopId" msgpack:"shopId"`
	RewardList []*RewardData `json:"rewards" msgpack:"rewards"`
}

type ShopShareReq struct {
	UserUid  string `json:"userUid" msgpack:"userUid" binding:"required"`
	ShopType byte   `json:"shopType" msgpack:"shopType" binding:"required"`
	ShopIdx  byte   `json:"shopIdx" msgpack:"shopIdx" binding:"required"`
}

type ShopShareResp struct {
	Param string `json:"param" msgpack:"param"`
}
