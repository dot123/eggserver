
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

package cfg;


type DailyShopTb struct {
    _dataMap map[int32]*IDailyShop
    _dataList []*IDailyShop
}

func NewDailyShopTb(_buf []map[string]interface{}) (*DailyShopTb, error) {
    _dataList := make([]*IDailyShop, 0, len(_buf))
    dataMap := make(map[int32]*IDailyShop)

    for _, _ele_ := range _buf {
        if _v, err2 := NewIDailyShop(_ele_); err2 != nil {
            return nil, err2
        } else {
            _dataList = append(_dataList, _v)
            dataMap[_v.Id] = _v
        }
    }
    return &DailyShopTb{_dataList:_dataList, _dataMap:dataMap}, nil
}

func (table *DailyShopTb) GetDataMap() map[int32]*IDailyShop {
    return table._dataMap
}

func (table *DailyShopTb) GetDataList() []*IDailyShop {
    return table._dataList
}

func (table *DailyShopTb) Get(key int32) *IDailyShop {
    return table._dataMap[key]
}


