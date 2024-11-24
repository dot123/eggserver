package schema

type TickerReq struct {
	InstId string `json:"instId" msgpack:"instId"`
}

type TickerResp struct {
	Last float64 `json:"last" msgpack:"last"`
}
