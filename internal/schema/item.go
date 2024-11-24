package schema

type Item struct {
	ItemID  int32 `json:"itemId" msgpack:"itemId"`
	ItemNum int32 `json:"itemNum" msgpack:"itemNum"`
}

type UseItemReq struct {
	ItemID int32  `json:"itemId" msgpack:"itemId"`
	Param  string `json:"param" msgpack:"param"`
}
