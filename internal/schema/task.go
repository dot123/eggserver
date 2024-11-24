package schema

type TaskState struct {
	TaskType byte  `json:"taskType" msgpack:"taskType"`
	TaskId   int32 `json:"taskId" msgpack:"taskId"`
	State    byte  `json:"state" msgpack:"state"`
}

type TaskProgress struct {
	TaskSubId int32 `json:"taskSubId" msgpack:"taskSubId"`
	Progress  int32 `json:"progress" msgpack:"progress"`
}

type GetTaskRewardReq struct {
	TaskId   int32 `json:"taskId" msgpack:"taskId" binding:"required"`
	TaskType int   `json:"taskType" msgpack:"taskType" binding:"required"`
}

type TaskGotoReq struct {
	UserUid   string `json:"userUid" msgpack:"userUid" binding:"required"`
	TaskSubId int32  `json:"taskSubId" msgpack:"taskSubId" binding:"required"`
}

type TaskGotoResp struct {
	Param string `json:"param" msgpack:"param"`
}

type TaskTonAccountReq struct {
	AccountId string `json:"accountId" msgpack:"accountId" binding:"required"`
}

type TaskData struct {
	Progress int32 `json:"progress" msgpack:"progress"`
	State    int32 `json:"state" msgpack:"state"`
	Complete int32 `json:"complete" msgpack:"complete"` // 完成了多少次
}

type TaskDataResp struct {
	Data map[int32]*TaskData `json:"data" msgpack:"data"`
}
