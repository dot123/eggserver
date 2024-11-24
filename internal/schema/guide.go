package schema

type GuideStepReq struct {
	GuideId int32 `json:"guideId" msgpack:"guideId"`
}

type GuideDataResp struct {
	Step []int32 `json:"step" msgpack:"step"`
}
