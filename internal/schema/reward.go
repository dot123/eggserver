package schema

type ItemData struct {
	ID   int32 `json:"id" msgpack:"id"`
	Num  int32 `json:"num" msgpack:"num"`
	Type int32 `json:"type" msgpack:"type"`
}

type RewardData struct {
	Item *ItemData `json:"item" msgpack:"item"`
	Egg  *Egg      `json:"egg" msgpack:"egg"`
}
