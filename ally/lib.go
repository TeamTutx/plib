package ally

import (
	"reflect"
)

//IsEmptyStruct : Check if structure is empty or not
func IsEmptyStruct(object interface{}) (flag bool) {
	flag = true
	refObj := reflect.ValueOf(object)
	if refObj.Kind() == reflect.Struct {
		for i := 0; i < refObj.NumField(); i++ {
			curField := refObj.Field(i)
			refObjField := refObj.Type().Field(i)
			curFieldName := refObjField.Name

			//Ignoring unexported fileds
			if len(curFieldName) > 0 && curFieldName[0] > 'Z' {
				continue
			}

			//check if default tag is there
			_, dExists := refObjField.Tag.Lookup("default")
			if dExists == true {
				flag = false
				break
			}

			if curField.Kind() == reflect.Struct {
				if flag = IsEmptyStruct(curField.Interface()); flag == false {
					break
				}
			} else if curField.Kind() == reflect.Map || curField.Kind() == reflect.Slice {
				if curField.Len() > 0 {
					flag = false
					break
				}
			} else {
				zero := reflect.Zero(curField.Type()).Interface()
				if reflect.DeepEqual(curField.Interface(), zero) == false {
					flag = false
					break
				}
			}
		}
	} else if refObj.Kind() == reflect.Slice {
		if refObj.Len() != 0 {
			flag = IsEmptyStruct(refObj.Index(0).Interface())
		}
	}
	return flag
}

//IsEmptyInterface : Check if interface is empty or not
func IsEmptyInterface(object interface{}) bool {
	return reflect.DeepEqual(object, reflect.Zero(reflect.TypeOf(object)).Interface())
}
