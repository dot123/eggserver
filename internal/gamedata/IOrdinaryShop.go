
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

package cfg;


import "errors"

type IOrdinaryShop struct {
    Id int32
    Desc string
    Name string
    Item *GlobalItemData
    Price *GlobalItemData
    Bg string
    Icon string
}

const TypeId_IOrdinaryShop = 1696518625

func (*IOrdinaryShop) GetTypeId() int32 {
    return 1696518625
}

func NewIOrdinaryShop(_buf map[string]interface{}) (_v *IOrdinaryShop, err error) {
    _v = &IOrdinaryShop{}
    { var _ok_ bool; var _tempNum_ float64; if _tempNum_, _ok_ = _buf["id"].(float64); !_ok_ { err = errors.New("id error"); return }; _v.Id = int32(_tempNum_) }
    { var _ok_ bool; if _v.Desc, _ok_ = _buf["desc"].(string); !_ok_ { err = errors.New("desc error"); return } }
    { var _ok_ bool; if _v.Name, _ok_ = _buf["name"].(string); !_ok_ { err = errors.New("name error"); return } }
    { var _ok_ bool; var _x_ map[string]interface{}; if _x_, _ok_ = _buf["item"].(map[string]interface{}); !_ok_ { err = errors.New("item error"); return }; if _v.Item, err = NewGlobalItemData(_x_); err != nil { return } }
    { var _ok_ bool; var _x_ map[string]interface{}; if _x_, _ok_ = _buf["price"].(map[string]interface{}); !_ok_ { err = errors.New("price error"); return }; if _v.Price, err = NewGlobalItemData(_x_); err != nil { return } }
    { var _ok_ bool; if _v.Bg, _ok_ = _buf["bg"].(string); !_ok_ { err = errors.New("bg error"); return } }
    { var _ok_ bool; if _v.Icon, _ok_ = _buf["icon"].(string); !_ok_ { err = errors.New("icon error"); return } }
    return
}

