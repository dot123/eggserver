
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

package cfg;


import "errors"

type IReSignTb struct {
    Id int32
    Cost *GlobalItemData
}

const TypeId_IReSignTb = 1294471

func (*IReSignTb) GetTypeId() int32 {
    return 1294471
}

func NewIReSignTb(_buf map[string]interface{}) (_v *IReSignTb, err error) {
    _v = &IReSignTb{}
    { var _ok_ bool; var _tempNum_ float64; if _tempNum_, _ok_ = _buf["id"].(float64); !_ok_ { err = errors.New("id error"); return }; _v.Id = int32(_tempNum_) }
    { var _ok_ bool; var _x_ map[string]interface{}; if _x_, _ok_ = _buf["cost"].(map[string]interface{}); !_ok_ { err = errors.New("cost error"); return }; if _v.Cost, err = NewGlobalItemData(_x_); err != nil { return } }
    return
}

