
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

package cfg;


type PetTb struct {
    _dataMap map[int32]*IPet
    _dataList []*IPet
}

func NewPetTb(_buf []map[string]interface{}) (*PetTb, error) {
    _dataList := make([]*IPet, 0, len(_buf))
    dataMap := make(map[int32]*IPet)

    for _, _ele_ := range _buf {
        if _v, err2 := NewIPet(_ele_); err2 != nil {
            return nil, err2
        } else {
            _dataList = append(_dataList, _v)
            dataMap[_v.Id] = _v
        }
    }
    return &PetTb{_dataList:_dataList, _dataMap:dataMap}, nil
}

func (table *PetTb) GetDataMap() map[int32]*IPet {
    return table._dataMap
}

func (table *PetTb) GetDataList() []*IPet {
    return table._dataList
}

func (table *PetTb) Get(key int32) *IPet {
    return table._dataMap[key]
}


