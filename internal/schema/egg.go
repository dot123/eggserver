package schema

type Egg struct {
	ID    uint64 `json:"id" msgpack:"id"`
	Part1 int32  `json:"part1" msgpack:"part1"`
	Part2 int32  `json:"part2" msgpack:"part2"`
	Part3 int32  `json:"part3" msgpack:"part3"`
}

type GetEggOpenReq struct {
	ID      int32 `json:"id" msgpack:"id" binding:"required"`
	GuideId int32 `json:"guideId" msgpack:"guideId"`
}
