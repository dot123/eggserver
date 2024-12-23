
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

package cfg;


import "errors"

type ITask struct {
    TaskId int32
    TaskSubId int32
    TaskName string
    Desc string
    TaskType int32
    Need int32
    MaxTimes int32
    Reward *GlobalItemData
    IconAvatar string
    Display int32
}

const TypeId_ITask = 70016366

func (*ITask) GetTypeId() int32 {
    return 70016366
}

func NewITask(_buf map[string]interface{}) (_v *ITask, err error) {
    _v = &ITask{}
    { var _ok_ bool; var _tempNum_ float64; if _tempNum_, _ok_ = _buf["taskId"].(float64); !_ok_ { err = errors.New("taskId error"); return }; _v.TaskId = int32(_tempNum_) }
    { var _ok_ bool; var _tempNum_ float64; if _tempNum_, _ok_ = _buf["taskSubId"].(float64); !_ok_ { err = errors.New("taskSubId error"); return }; _v.TaskSubId = int32(_tempNum_) }
    { var _ok_ bool; if _v.TaskName, _ok_ = _buf["taskName"].(string); !_ok_ { err = errors.New("taskName error"); return } }
    { var _ok_ bool; if _v.Desc, _ok_ = _buf["desc"].(string); !_ok_ { err = errors.New("desc error"); return } }
    { var _ok_ bool; var _tempNum_ float64; if _tempNum_, _ok_ = _buf["taskType"].(float64); !_ok_ { err = errors.New("taskType error"); return }; _v.TaskType = int32(_tempNum_) }
    { var _ok_ bool; var _tempNum_ float64; if _tempNum_, _ok_ = _buf["need"].(float64); !_ok_ { err = errors.New("need error"); return }; _v.Need = int32(_tempNum_) }
    { var _ok_ bool; var _tempNum_ float64; if _tempNum_, _ok_ = _buf["maxTimes"].(float64); !_ok_ { err = errors.New("maxTimes error"); return }; _v.MaxTimes = int32(_tempNum_) }
    { var _ok_ bool; var _x_ map[string]interface{}; if _x_, _ok_ = _buf["reward"].(map[string]interface{}); !_ok_ { err = errors.New("reward error"); return }; if _v.Reward, err = NewGlobalItemData(_x_); err != nil { return } }
    { var _ok_ bool; if _v.IconAvatar, _ok_ = _buf["iconAvatar"].(string); !_ok_ { err = errors.New("iconAvatar error"); return } }
    { var _ok_ bool; var _tempNum_ float64; if _tempNum_, _ok_ = _buf["display"].(float64); !_ok_ { err = errors.New("display error"); return }; _v.Display = int32(_tempNum_) }
    return
}

