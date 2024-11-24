package schema

type PassPortRewardGetReq struct {
	Id     int32 `json:"id" msgpack:"id"`
	Deluxe byte  `json:"deluxe" msgpack:"deluxe"`
}

type PassPortRewardGetResp struct {
	Reward *RewardData `json:"reward" msgpack:"reward"`
}

type PassPortRewardDataResp struct {
	Data map[int32]byte `json:"data" msgpack:"data"`
}
