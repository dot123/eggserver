
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

package cfg;


type BattleConfigTb struct {
    _dataMap map[int32]*IBattleConfig
    _dataList []*IBattleConfig
}

func NewBattleConfigTb(_buf []map[string]interface{}) (*BattleConfigTb, error) {
    _dataList := make([]*IBattleConfig, 0, len(_buf))
    dataMap := make(map[int32]*IBattleConfig)

    for _, _ele_ := range _buf {
        if _v, err2 := NewIBattleConfig(_ele_); err2 != nil {
            return nil, err2
        } else {
            _dataList = append(_dataList, _v)
            dataMap[_v.Id] = _v
        }
    }
    return &BattleConfigTb{_dataList:_dataList, _dataMap:dataMap}, nil
}

func (table *BattleConfigTb) GetDataMap() map[int32]*IBattleConfig {
    return table._dataMap
}

func (table *BattleConfigTb) GetDataList() []*IBattleConfig {
    return table._dataList
}

func (table *BattleConfigTb) Get(key int32) *IBattleConfig {
    return table._dataMap[key]
}


