
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

package cfg;


import "errors"

type IEggPart3 struct {
    Id int32
    Quality int32
    Weight int32
}

const TypeId_IEggPart3 = 1143924996

func (*IEggPart3) GetTypeId() int32 {
    return 1143924996
}

func NewIEggPart3(_buf map[string]interface{}) (_v *IEggPart3, err error) {
    _v = &IEggPart3{}
    { var _ok_ bool; var _tempNum_ float64; if _tempNum_, _ok_ = _buf["id"].(float64); !_ok_ { err = errors.New("id error"); return }; _v.Id = int32(_tempNum_) }
    { var _ok_ bool; var _tempNum_ float64; if _tempNum_, _ok_ = _buf["quality"].(float64); !_ok_ { err = errors.New("quality error"); return }; _v.Quality = int32(_tempNum_) }
    { var _ok_ bool; var _tempNum_ float64; if _tempNum_, _ok_ = _buf["weight"].(float64); !_ok_ { err = errors.New("weight error"); return }; _v.Weight = int32(_tempNum_) }
    return
}

