
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

package cfg;


type OrdinaryShopTb struct {
    _dataMap map[int32]*IOrdinaryShop
    _dataList []*IOrdinaryShop
}

func NewOrdinaryShopTb(_buf []map[string]interface{}) (*OrdinaryShopTb, error) {
    _dataList := make([]*IOrdinaryShop, 0, len(_buf))
    dataMap := make(map[int32]*IOrdinaryShop)

    for _, _ele_ := range _buf {
        if _v, err2 := NewIOrdinaryShop(_ele_); err2 != nil {
            return nil, err2
        } else {
            _dataList = append(_dataList, _v)
            dataMap[_v.Id] = _v
        }
    }
    return &OrdinaryShopTb{_dataList:_dataList, _dataMap:dataMap}, nil
}

func (table *OrdinaryShopTb) GetDataMap() map[int32]*IOrdinaryShop {
    return table._dataMap
}

func (table *OrdinaryShopTb) GetDataList() []*IOrdinaryShop {
    return table._dataList
}

func (table *OrdinaryShopTb) Get(key int32) *IOrdinaryShop {
    return table._dataMap[key]
}


