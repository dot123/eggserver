
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

package cfg;


type GlobalTb struct {
    _dataMap map[int32]*IGlobal
    _dataList []*IGlobal
}

func NewGlobalTb(_buf []map[string]interface{}) (*GlobalTb, error) {
    _dataList := make([]*IGlobal, 0, len(_buf))
    dataMap := make(map[int32]*IGlobal)

    for _, _ele_ := range _buf {
        if _v, err2 := NewIGlobal(_ele_); err2 != nil {
            return nil, err2
        } else {
            _dataList = append(_dataList, _v)
            dataMap[_v.LayEggTime] = _v
        }
    }
    return &GlobalTb{_dataList:_dataList, _dataMap:dataMap}, nil
}

func (table *GlobalTb) GetDataMap() map[int32]*IGlobal {
    return table._dataMap
}

func (table *GlobalTb) GetDataList() []*IGlobal {
    return table._dataList
}

func (table *GlobalTb) Get(key int32) *IGlobal {
    return table._dataMap[key]
}

