
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

package cfg;


import "errors"

type IRedPoint struct {
    Id string
}

const TypeId_IRedPoint = 465136360

func (*IRedPoint) GetTypeId() int32 {
    return 465136360
}

func NewIRedPoint(_buf map[string]interface{}) (_v *IRedPoint, err error) {
    _v = &IRedPoint{}
    { var _ok_ bool; if _v.Id, _ok_ = _buf["id"].(string); !_ok_ { err = errors.New("id error"); return } }
    return
}
