
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

package cfg;


type EggPart1Tb struct {
    _dataMap map[int32]*IEggPart1
    _dataList []*IEggPart1
}

func NewEggPart1Tb(_buf []map[string]interface{}) (*EggPart1Tb, error) {
    _dataList := make([]*IEggPart1, 0, len(_buf))
    dataMap := make(map[int32]*IEggPart1)

    for _, _ele_ := range _buf {
        if _v, err2 := NewIEggPart1(_ele_); err2 != nil {
            return nil, err2
        } else {
            _dataList = append(_dataList, _v)
            dataMap[_v.Id] = _v
        }
    }
    return &EggPart1Tb{_dataList:_dataList, _dataMap:dataMap}, nil
}

func (table *EggPart1Tb) GetDataMap() map[int32]*IEggPart1 {
    return table._dataMap
}

func (table *EggPart1Tb) GetDataList() []*IEggPart1 {
    return table._dataList
}

func (table *EggPart1Tb) Get(key int32) *IEggPart1 {
    return table._dataMap[key]
}


