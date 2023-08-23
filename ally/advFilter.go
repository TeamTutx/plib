package ally

//SetCustomVal will set custom value in filter from given list if that filter exists
func (f *AdvFilter) SetCustomVal(list map[string]FilterDet, tag string, value interface{}) {
	if filterDet, exists := list[tag]; exists {
		if filterDet.Key != "" {
			tag = filterDet.Key
		}
		f.Filter[tag] = GetFilterValue(filterDet, filterDet.Operator, IfToA(value))
	}
}

//GetFilterValue will return the FilterDet struct
func GetFilterValue(filterDet FilterDet, operator string, value string) FilterDet {

	return FilterDet{
		Operator:     operator,
		Value:        value,
		Alias:        filterDet.Alias,
		Sort:         filterDet.Sort,
		Skip:         filterDet.Skip,
		Or:           filterDet.Or,
		Unescape:     filterDet.Unescape,
		ESFilterType: filterDet.ESFilterType,
		Exclude:      filterDet.Exclude,
	}
}
