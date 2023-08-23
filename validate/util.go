package validate

import (
	"math"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/TeamTutx/plib/ally"
	"github.com/TeamTutx/plib/constant"
	"github.com/TeamTutx/plib/empty"
	"github.com/TeamTutx/plib/perror"

	validator "gopkg.in/go-playground/validator.v9"
)

var (
	//V is global validator struct
	V *validator.Validate
)

//Struct will validate the struct based of validate tag
func Struct(structModel interface{}) (err error) {
	if empty.IsEmptyStruct(structModel) == true {
		err = perror.NoUpdateError()
	} else {
		refObj := reflect.ValueOf(structModel)
		if refObj.Kind() == reflect.Struct {
			if err = V.Struct(structModel); err != nil {
				err = perror.InvalidParamError(err)
			}
		} else if refObj.Kind() == reflect.Slice {
			for i := 0; i < refObj.Len(); i++ {
				if err = V.Struct(refObj.Index(i)); err != nil {
					err = perror.InvalidParamError(err)
					break
				}
			}
		}
	}
	return
}

//UnmarshalV : unmarshal of request body and validate the data
func UnmarshalV(reqBody string, structModel interface{}) (err error) {
	if err = ally.Unmarshal([]byte(reqBody), structModel); err == nil {
		err = Struct(reflect.ValueOf(structModel).Elem().Interface())
	}
	return
}

//VAdvFilter : validate advance filter query parameters value
// paging is true if limit and offset required
func VAdvFilter(qParams map[string]string, filterList map[string]ally.FilterDet, paging bool) (
	filter ally.AdvFilter, err error) {
	filterOpt := make(map[string]ally.FilterDet)
	skipUnescape := qParams["unescape"]
	for key, filterDet := range filterList {
		if value, ok := qParams[key]; ok == true {

			if key == ally.SkipLimit && value == "true" {
				paging = false
			}

			//Unescape the query value
			if filterDet.Unescape && skipUnescape != "false" {
				value, _ = url.QueryUnescape(value)
			}
			field := key
			if filterDet.Display != "" {
				field = filterDet.Display
			}
			msg := "Invalid " + ally.SnakeName(field)
			if err = V.Var(value, filterDet.ValidTag); err != nil {
				err = perror.VError(msg + " value:" + value + " tag:" + filterDet.ValidTag).
					SetMsg(msg)
				break
			}
			if filterDet.Key != "" {
				key = filterDet.Key
			}
			if filterDet.Operator == constant.SearchOP {
				keys := strings.Split(key, ",")
				for _, curKey := range keys {
					// filterDet.Unescape = true
					filterOpt[curKey] = ally.GetFilterValue(filterDet, constant.LikeOP, value)
				}
			} else if filterDet.Operator == constant.SplitSearchOP {
				keys := strings.Split(key, ",")
				for _, curKey := range keys {
					// filterDet.Unescape = true
					filterOpt[curKey] = ally.GetFilterValue(filterDet, constant.LikeOrOP, value)
				}
				filter.SplitSearchStr = value
			} else if filterDet.Operator == constant.LikeAllSearchOP {
				filterOpt[key] = ally.GetFilterValue(filterDet, constant.LikeAllSearchOP, value)
			} else {
				filterOpt[key] = ally.GetFilterValue(filterDet, filterDet.Operator, value)
			}
		} else if filterDet.DefaultValue != "" {
			operator := filterDet.Operator
			if filterDet.DefaultOperator != "" {
				operator = filterDet.DefaultOperator
			}
			if filterDet.Key != "" {
				key = filterDet.Key
			}
			filterOpt[key] = ally.GetFilterValue(filterDet, operator, filterDet.DefaultValue)
		}
	}

	filter.Filter = filterOpt
	if err == nil && paging == true {
		if filter.Offset, filter.Limit, err =
			getPaginationParam(qParams, constant.PageDefaultLimit); err == nil {
		}
	}
	if err == nil {
		filter.IncludeFields = qParams[constant.IncludeOP]
		filter.ExcludeFields = qParams[constant.ExcludeOP]
	}
	return
}

//getPaginationParam : get limit & offset to paginate search result
func getPaginationParam(qParams map[string]string, defLimit int) (
	offset int, limit int, err error) {
	if value, ok := qParams["offset"]; err == nil && ok == true {
		if offset, err = strconv.Atoi(value); err == nil {
			if offset <= 0 {
				offset = 0
			}
		} else {
			err = perror.VError("invalid offset", value)
		}
	}

	if value, ok := qParams["limit"]; err == nil && ok == true {
		if limit, err = strconv.Atoi(value); err != nil {
			err = perror.VError("invalid limit", value)
		}
	}

	//set default limit in case of not passed it as request parameter
	if limit <= 0 {
		limit = defLimit
	}
	return
}

//VOverlapping : validate the range overlapping in slice
func VOverlapping(slice interface{}, minTag, maxTag string) (err error) {
	refObj := reflect.ValueOf(slice)
	sliceLen := refObj.Len()
	val := refObj.Slice(0, sliceLen)
	var diff int

	if sliceLen > 1 {
		if refObj.Index(0).FieldByName(minTag).IsValid() == false {
			err = perror.CustomError("Invalid minTag")
		}
		if err == nil && refObj.Index(0).FieldByName(maxTag).IsValid() == false {
			err = perror.CustomError("Invalid maxTag")
		}
		if err == nil {
			sort.Slice(slice, func(i, j int) bool {
				diff, _ := Compare(val.Index(i).FieldByName(minTag).Interface(),
					val.Index(j).FieldByName(minTag).Interface())
				return diff < 0
			})
			for i := 0; i < sliceLen-1; i++ {
				diff, err = Compare(val.Index(i).FieldByName(maxTag).Interface(),
					val.Index(i+1).FieldByName(minTag).Interface())
				if err == nil && diff >= 0 {
					err = perror.VError("Range Overlapped")
					break
				}
			}
		}
	}
	return
}

//Compare : returns difference of values
func Compare(val1, val2 interface{}) (diff int, err error) {
	reflObj1 := reflect.ValueOf(val1)
	reflObj2 := reflect.ValueOf(val2)
	if reflObj1.Kind() != reflObj2.Kind() {
		err = perror.CustomError("Value Type Mismatch")
	} else {
		switch reflObj1.Kind() {
		case reflect.Int:
			diff = (int)(reflObj1.Int() - reflObj2.Int())
		case reflect.Float64, reflect.Float32:
			diff = (int)(math.Ceil(reflObj1.Float() - reflObj2.Float()))
		default:
			err = perror.CustomError("Invalid type")
		}
	}
	return
}
