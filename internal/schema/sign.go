package schema

type SignInReq struct {
	Id       int32 `json:"id" msgpack:"id"`
	ReSignIn byte  `json:"reSignIn" msgpack:"reSignIn"`
}

type SignInResp struct {
	RewardList []*RewardData `json:"rewards" msgpack:"rewards"`
}

type SignDataResp struct {
	StartTime     int64 `json:"startTime" msgpack:"startTime"`         // 签到开始时间
	LoginNum      int   `json:"loginNum" msgpack:"loginNum"`           // 签到登录天数
	LoginTime     int64 `json:"loginTime" msgpack:"loginTime"`         // 签到登录时间
	SignCardETime int64 `json:"signCardETime" msgpack:"signCardETime"` // 豪华签到奖励结束时间
	Day           int   `json:"day" msgpack:"day"`                     // 签到天数
	ReSign        int   `json:"reSign" msgpack:"reSign"`               // 补签次数
	IsSign        byte  `json:"isSign" msgpack:"isSign"`               // 今日是否签到
}
