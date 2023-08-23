package ally

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/TeamTutx/plib/perror"
)

//ItoASlice : Change provided integer slice to string slice
func ItoASlice(data []int) (stringData []string) {
	stringData = make([]string, 0)
	for _, value := range data {
		stringData = append(stringData, strconv.Itoa(value))
	}
	return stringData
}

//IfToA converts interface value to string
func IfToA(value interface{}) string {
	if value == nil {
		return ""
	} else if reflect.TypeOf(value).Kind() == reflect.Float64 {
		return strconv.FormatFloat(value.(float64), 'f', -1, 64)
	}
	return fmt.Sprintf("%v", value)
}

//IfToI converts interface value to int
func IfToI(value interface{}) (int, error) {
	var (
		intVal int
		err    error
	)
	if value != nil {
		objKind := reflect.TypeOf(value).Kind()
		switch objKind {
		case reflect.Int:
			intVal = value.(int)
		case reflect.Int64:
			intVal = int(value.(int64))
		case reflect.Float64:
			intVal = int(value.(float64))
		case reflect.Float32:
			intVal = int(value.(float32))
		case reflect.String:
			intVal, err = strconv.Atoi(value.(string))
		default:
			err = perror.CustomError("Not Able to typecast",
				objKind.String(), "to int")
		}
	} else {
		err = perror.CustomError("nil value can't be typecast")
	}
	return intVal, err
}

//IfToForceI will convert intercase to int ignoring error
func IfToForceI(value interface{}) (intVal int) {
	intVal, _ = IfToI(value)
	return
}

//IfToF converts interface value to float
func IfToF(value interface{}) float64 {
	var floatVal float64
	if value != nil {
		objKind := reflect.TypeOf(value).Kind()
		switch objKind {
		case reflect.Int:
			floatVal = float64(value.(int))
		case reflect.Int64:
			floatVal = float64(value.(int64))
		case reflect.Float64:
			floatVal = value.(float64)
		case reflect.Float32:
			floatVal = float64(value.(float32))
		case reflect.String:
			floatVal, _ = strconv.ParseFloat(value.(string), 64)
		}
	}
	return floatVal
}

//AtoISlice : Convert Slice of string to slice of integer
// eIgnore flag is used to ignore the error while converting
func AtoISlice(value []string, eIgnore bool) (intVal []int, err error) {
	intVal = make([]int, 0)
	for _, curVal := range value {
		var curIntVal int
		curIntVal, err = strconv.Atoi(curVal)
		if eIgnore == false && err != nil {
			return
		} else if err != nil {
			continue
		}
		intVal = append(intVal, curIntVal)
	}
	return
}

//IftoASlice will return interface to string slice
func IftoASlice(data interface{}) (stringData []string) {
	stringData = make([]string, 0)
	if data != nil {
		if sData, ok := data.([]interface{}); ok {
			for _, v := range sData {
				stringData = append(stringData, IfToA(v))
			}
		}
	}
	return
}

//StrintToInt64 : Convert string to Int64
func StrintToInt64(data string) (val int64, err error) {
	if val, err = strconv.ParseInt(data, 10, 64); err != nil {
		err = perror.CustomError("Not Able to typecast",
			data, "to int64")
		return
	}
	return
}

//StrintToForceInt64 : Convert string to Int64 witout Error
func StrintToForceInt64(data string) (val int64) {
	val, _ = StrintToInt64(data)
	return
}
