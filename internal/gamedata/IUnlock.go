
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

package cfg;


import "errors"

type IUnlock struct {
    Id int32
    Condition []int32
    NodePath string
}

const TypeId_IUnlock = -1393315603

func (*IUnlock) GetTypeId() int32 {
    return -1393315603
}

func NewIUnlock(_buf map[string]interface{}) (_v *IUnlock, err error) {
    _v = &IUnlock{}
    { var _ok_ bool; var _tempNum_ float64; if _tempNum_, _ok_ = _buf["id"].(float64); !_ok_ { err = errors.New("id error"); return }; _v.Id = int32(_tempNum_) }
     {
                    var _arr_ []interface{}
                    var _ok_ bool
                    if _arr_, _ok_ = _buf["condition"].([]interface{}); !_ok_ { err = errors.New("condition error"); return }
    
                    _v.Condition = make([]int32, 0, len(_arr_))
                    
                    for _, _e_ := range _arr_ {
                        var _list_v_ int32
                        { var _ok_ bool; var _x_ float64; if _x_, _ok_ = _e_.(float64); !_ok_ { err = errors.New("_list_v_ error"); return }; _list_v_ = int32(_x_) }
                        _v.Condition = append(_v.Condition, _list_v_)
                    }
                }

    { var _ok_ bool; if _v.NodePath, _ok_ = _buf["nodePath"].(string); !_ok_ { err = errors.New("nodePath error"); return } }
    return
}
