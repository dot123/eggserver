package schema

type RankData struct {
	RoleId    uint64 `json:"roleId" msgpack:"roleId"`
	Rank      int64  `json:"rank" msgpack:"rank"`
	FirstName string `json:"firstName" msgpack:"firstName"`
	LastName  string `json:"lastName" msgpack:"lastName"`
	GoldNum   int32  `json:"goldNum" msgpack:"goldNum"`
}

type LeaderboardResp struct {
	List       []*RankData `json:"list" msgpack:"list"`
	MyRankData *RankData   `json:"myRankData" msgpack:"myRankData"`
}
