
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

package cfg;


type DiamondShopTb struct {
    _dataMap map[int32]*IDiamondShopTb
    _dataList []*IDiamondShopTb
}

func NewDiamondShopTb(_buf []map[string]interface{}) (*DiamondShopTb, error) {
    _dataList := make([]*IDiamondShopTb, 0, len(_buf))
    dataMap := make(map[int32]*IDiamondShopTb)

    for _, _ele_ := range _buf {
        if _v, err2 := NewIDiamondShopTb(_ele_); err2 != nil {
            return nil, err2
        } else {
            _dataList = append(_dataList, _v)
            dataMap[_v.Id] = _v
        }
    }
    return &DiamondShopTb{_dataList:_dataList, _dataMap:dataMap}, nil
}

func (table *DiamondShopTb) GetDataMap() map[int32]*IDiamondShopTb {
    return table._dataMap
}

func (table *DiamondShopTb) GetDataList() []*IDiamondShopTb {
    return table._dataList
}

func (table *DiamondShopTb) Get(key int32) *IDiamondShopTb {
    return table._dataMap[key]
}


