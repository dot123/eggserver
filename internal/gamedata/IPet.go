
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

package cfg;


import "errors"

type IPet struct {
    Id int32
    Name string
    Quality int32
    Score float64
    Desc string
    Head string
    Body string
}

const TypeId_IPet = 2254870

func (*IPet) GetTypeId() int32 {
    return 2254870
}

func NewIPet(_buf map[string]interface{}) (_v *IPet, err error) {
    _v = &IPet{}
    { var _ok_ bool; var _tempNum_ float64; if _tempNum_, _ok_ = _buf["id"].(float64); !_ok_ { err = errors.New("id error"); return }; _v.Id = int32(_tempNum_) }
    { var _ok_ bool; if _v.Name, _ok_ = _buf["name"].(string); !_ok_ { err = errors.New("name error"); return } }
    { var _ok_ bool; var _tempNum_ float64; if _tempNum_, _ok_ = _buf["quality"].(float64); !_ok_ { err = errors.New("quality error"); return }; _v.Quality = int32(_tempNum_) }
    { var _ok_ bool; var _tempNum_ float64; if _tempNum_, _ok_ = _buf["score"].(float64); !_ok_ { err = errors.New("score error"); return }; _v.Score = float64(_tempNum_) }
    { var _ok_ bool; if _v.Desc, _ok_ = _buf["desc"].(string); !_ok_ { err = errors.New("desc error"); return } }
    { var _ok_ bool; if _v.Head, _ok_ = _buf["head"].(string); !_ok_ { err = errors.New("head error"); return } }
    { var _ok_ bool; if _v.Body, _ok_ = _buf["body"].(string); !_ok_ { err = errors.New("body error"); return } }
    return
}
