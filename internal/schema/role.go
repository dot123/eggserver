package schema

type RoleDataResp struct {
	Token                string                  `json:"token" msgpack:"token"`
	RoleID               uint64                  `json:"roleId" msgpack:"roleId"`
	Pets                 []*Pet                  `json:"pets" msgpack:"pets"`
	Items                []*Item                 `json:"items" msgpack:"items"`
	Eggs                 []*Egg                  `json:"eggs" msgpack:"eggs"`
	Shop                 *ShopDataResp           `json:"shop" msgpack:"shop"`
	Task                 *TaskDataResp           `json:"task" msgpack:"task"`
	Guide                *GuideDataResp          `json:"guide" msgpack:"guide"`
	Sign                 *SignDataResp           `json:"sign" msgpack:"sign"`
	PassPortReward       *PassPortRewardDataResp `json:"passPortReward" msgpack:"passPortReward"`
	ServerTime           int64                   `json:"serverTime" msgpack:"serverTime"`
	LastLoginTime        int64                   `json:"lastLoginTime" msgpack:"lastLoginTime"`
	LastLayEggTime       int64                   `json:"lastLayEggTime" msgpack:"lastLayEggTime"`
	LastAddVitTime       int64                   `json:"lastAddVitTime" msgpack:"lastAddVitTime"`
	LastClickCount       int32                   `json:"lastClickCount" msgpack:"lastClickCount"`
	LastDeskId           string                  `json:"lastDeskId" msgpack:"lastDeskId"`
	AutoEggCollectRTime  int64                   `json:"autoEggCollectRTime" msgpack:"autoEggCollectRTime"`   // 自动收蛋剩余时间
	AutoEggCollectETime  int64                   `json:"autoEggCollectETime" msgpack:"autoEggCollectETime"`   // 自动收蛋开始时间
	BattleCount          int32                   `json:"battleCount" msgpack:"battleCount"`                   // 战斗次数
	PassPortDeluxeReward byte                    `json:"passPortDeluxeReward" msgpack:"passPortDeluxeReward"` // 是否有通行证豪华奖励
	GM                   byte                    `json:"gm" msgpack:"gm"`
}

type ClickScreenReq struct {
	ClickCount int32 `json:"clickCount" msgpack:"clickCount" binding:"required"`
}

type AutoAddVitResp struct {
	Reward         *RewardData `json:"reward" msgpack:"reward"`
	LastAddVitTime int64       `json:"lastAddVitTime" msgpack:"lastAddVitTime"`
}

type AutoLayEggResp struct {
	Eggs                []*Egg `json:"egg" msgpack:"eggs"`
	LastLayEggTime      int64  `json:"lastLayEggTime" msgpack:"lastLayEggTime"`
	AutoEggCollectRTime int64  `json:"autoEggCollectRTime" msgpack:"autoEggCollectRTime"`  // 自动收蛋剩余时间
	AutoEggCollectETime int64  `json:"autoEggCollectETime;" msgpack:"autoEggCollectETime"` // 自动收蛋开始时间
}

type ClickScreenResp struct {
	Reward         *RewardData `json:"reward" msgpack:"reward"`
	LastClickCount int32       `json:"lastClickCount" msgpack:"lastClickCount"`
}

type GMReq struct {
	Cmd string `json:"cmd" msgpack:"cmd" binding:"required"`
}

type GMResp struct {
	RewardList  []*RewardData `json:"rewards" msgpack:"rewards"`
	BattleCount int32         `json:"battleCount" msgpack:"battleCount"` // 战斗次数
}

type SyncDataResp struct {
	Task       *TaskDataResp `json:"task" msgpack:"task"`
	Shop       *ShopDataResp `json:"shop" msgpack:"shop"`
	ServerTime int64         `json:"serverTime" msgpack:"serverTime"`
}

type DailyResp struct {
	Shop       *ShopDataResp `json:"shop" msgpack:"shop"`
	Sign       *SignDataResp `json:"sign" msgpack:"sign"`
	Reward     *RewardData   `json:"reward" msgpack:"reward"`
	ServerTime int64         `json:"serverTime" msgpack:"serverTime"`
}
