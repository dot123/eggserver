
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

package cfg;


type QualityTb struct {
    _dataMap map[int32]*IQuality
    _dataList []*IQuality
}

func NewQualityTb(_buf []map[string]interface{}) (*QualityTb, error) {
    _dataList := make([]*IQuality, 0, len(_buf))
    dataMap := make(map[int32]*IQuality)

    for _, _ele_ := range _buf {
        if _v, err2 := NewIQuality(_ele_); err2 != nil {
            return nil, err2
        } else {
            _dataList = append(_dataList, _v)
            dataMap[_v.Id] = _v
        }
    }
    return &QualityTb{_dataList:_dataList, _dataMap:dataMap}, nil
}

func (table *QualityTb) GetDataMap() map[int32]*IQuality {
    return table._dataMap
}

func (table *QualityTb) GetDataList() []*IQuality {
    return table._dataList
}

func (table *QualityTb) Get(key int32) *IQuality {
    return table._dataMap[key]
}


