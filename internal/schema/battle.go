package schema

type Battle struct {
	BattleId  int32  `json:"battleId" msgpack:"battleId"`
	CreatedAt int64  `json:"createdAt" msgpack:"createdAt"`
	PlayerNum int    `json:"playerNum" msgpack:"playerNum"`
	DeskId    string `json:"deskId" msgpack:"deskId"`
}

type BattleMatchReq struct {
	BattleId int32 `json:"battleId" msgpack:"battleId" binding:"required"`
	PetId    int32 `json:"petId" msgpack:"petId" binding:"required"`
}

type BattleMatchResp struct {
	MatchState *BattleMatchStateResp `json:"matchState" msgpack:"matchState"`
	Reward     *RewardData           `json:"reward" msgpack:"reward"`
}

type BattleMatchStateReq struct {
	DeskId string `json:"deskId" msgpack:"deskId" binding:"required"`
}

type BattleMatchStateResp struct {
	BattleId  int32  `json:"battleId" msgpack:"battleId"`
	CreatedAt int64  `json:"createdAt" msgpack:"createdAt"`
	StartAt   int64  `json:"startAt" msgpack:"startAt"`
	JoinAt    int64  `json:"joinAt" msgpack:"joinAt"`
	PlayerNum int    `json:"playerNum" msgpack:"playerNum"`
	DeskId    string `json:"deskId" msgpack:"deskId"`
}

type BattleLeaveReq struct {
	DeskId string `json:"deskId" msgpack:"deskId" binding:"required"`
}

type BattleLeaveResp struct {
	Reward *RewardData `json:"reward" msgpack:"reward"`
}

type BattleExitReq struct {
	DeskId string `json:"deskId" msgpack:"deskId" binding:"required"`
}

type BattleExitResp struct {
	State byte `json:"state" msgpack:"state"` // 0立即退出 1退出无效
}

type BattleBetReq struct {
	DeskId string `json:"deskId" msgpack:"deskId" binding:"required"`
	Grid   int32  `json:"grid" msgpack:"grid" binding:"required"`
}

type BattleSyncScoreResp struct {
	ScoreList   map[int32]int32   `json:"scoreList,omitempty" msgpack:"scoreList"`
	PetList     map[int32][]int32 `json:"petList,omitempty" msgpack:"petList"`
	PlayerBonus int32             `json:"playerBonus" msgpack:"playerBonus"`
	BaseBonus   int32             `json:"baseBonus" msgpack:"baseBonus"`
	Grid        int32             `json:"grid" msgpack:"grid"`
}

type BattleSyncScoreReq struct {
	DeskId string `json:"deskId" msgpack:"deskId" binding:"required"`
}

type BattleRoundResultReq struct {
	DeskId string `json:"deskId" msgpack:"deskId" binding:"required"`
}

type BattleRoundResultResp struct {
	SyncScore      *BattleSyncScoreResp `json:"syncScore" msgpack:"syncScore"`
	PetId          int32                `json:"petId" msgpack:"petId"`
	BattleId       int32                `json:"battleId" msgpack:"battleId"`
	RoundStartTime int64                `json:"roundStartTime" msgpack:"roundStartTime"` // 回合时间
	Round          int                  `json:"round" msgpack:"round"`
	Result         []int32              `json:"result,omitempty" msgpack:"result"`
	LeftPlayerNum  int                  `json:"leftPlayerNum,omitempty" msgpack:"leftPlayerNum"`
	State          byte                 `json:"state" msgpack:"state"` // 0正在进行游戏，1战斗已结束，2游戏还没开始, 3可以进行结算
}

type BattleSettlementReq struct {
	DeskId string `json:"deskId" msgpack:"deskId" binding:"required"`
}

type BattleSettlementResp struct {
	Win    byte        `json:"win" msgpack:"win"`
	Bonus  int32       `json:"bonus" msgpack:"bonus"`
	Reward *RewardData `json:"reward" msgpack:"reward"`
}
