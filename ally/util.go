package ally

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/TeamTutx/plib/conf"
	"github.com/TeamTutx/plib/perror"
	jsoniter "github.com/json-iterator/go"
)

var (
	//SkipLimit : skip limit filter to skip limit and offset
	SkipLimit string
	//FetchFields : fields need to be fetched from api
	FetchFields   string
	services      map[string]interface{}
	serviceMux    sync.RWMutex
	debugServices map[string]interface{}
	apiService    map[string]interface{}
	apiServiceMux sync.RWMutex
	letterRunes   []rune
	numRunes      []rune
	loc           *time.Location
	packageName   string
)

func init() {
	packageName = "github.com/TeamTutx/plib"
	SkipLimit = "skip_limit"
	FetchFields = "fetch_fields"
	rand.Seed(time.Now().UnixNano())
	letterRunes = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	numRunes = []rune("0123456789")
	services = make(map[string]interface{})
	debugServices = make(map[string]interface{})
	apiService = make(map[string]interface{})
	loc, _ = time.LoadLocation("Asia/Kolkata")
}

//AddService will add new service
//If services are added in debug then only those service will execute for debug purpose
func AddService(module interface{}, debug ...bool) {
	refObj := reflect.ValueOf(module)
	if refObj.Kind() == reflect.Ptr {
		moduleName := refObj.Elem().Type().String()
		serviceMux.Lock()
		services[moduleName] = module
		serviceMux.Unlock()
		if len(debug) > 0 && debug[0] {
			debugServices[moduleName] = module
		}
		for i := 1; i < refObj.Elem().NumField(); i++ {
			curFieldName := refObj.Elem().Type().Field(i).Name
			api := moduleName + "." + curFieldName
			apiServiceMux.Lock()
			apiService[api] = module
			apiServiceMux.Unlock()
		}
	}
}

//GetService will return new service
func GetService() []interface{} {
	var modules []interface{}
	if conf.Bool("api_debug", false) && len(debugServices) > 0 {
		for _, curService := range debugServices {
			modules = append(modules, curService)
		}
	} else {
		serviceMux.RLock()
		for _, curService := range services {
			modules = append(modules, curService)
		}
		serviceMux.RUnlock()
	}
	return modules
}

//copyModule will create a copy of module
func copyModule(module interface{}) (copyModule interface{}) {
	if module != nil {
		copyModule = reflect.New(reflect.ValueOf(module).Elem().Type()).Interface()
	}
	return
}

//GetServiceModule will return the pointer of module by name
func GetServiceModule(moduleName string) (module interface{}) {
	serviceMux.RLock()
	module = services[moduleName]
	serviceMux.RUnlock()
	module = copyModule(module)
	return
}

//GetAPIModule will return the pointer of module by api name
func GetAPIModule(apiName string) (module interface{}) {
	apiServiceMux.RLock()
	module = apiService[apiName]
	apiServiceMux.RUnlock()
	module = copyModule(module)
	return
}

//SetModel : set source struct model to target model
// @ty: skip tag is used to skip any field value from binding to target field of same name
func SetModel(source, target interface{}) (err error) {

	// startLow := regexp.MustCompile(`^[a-z]`).MatchString
	if reflect.TypeOf(target).Kind() == reflect.Ptr {
		refObj := reflect.ValueOf(source)
		if refObj.Kind() == reflect.Ptr {
			refObj = refObj.Elem()
		}
		if refObj.Kind() == reflect.Struct {
			for i := 0; i < refObj.NumField(); i++ {
				curFieldName := refObj.Type().Field(i).Name
				curFieldTag := refObj.Type().Field(i).Tag
				curFieldValue := refObj.Field(i)

				//Ignoring unexported fileds
				if len(curFieldName) > 0 && curFieldName[0] > 'Z' {
					continue
				}

				//Skip the request value to map with database model
				if val, exists := curFieldTag.Lookup("ty"); exists && val == "skip" {
					continue
				}

				if curFieldValue.Kind() == reflect.Ptr {
					if curFieldValue.IsNil() {
						continue
					} else {
						curFieldValue = curFieldValue.Elem()
					}
				}
				if curFieldValue.Kind() == reflect.Struct {
					SetModel(curFieldValue.Interface(), target)
				} else if curFieldValue.Kind() == reflect.Slice {
					f := reflect.Indirect(reflect.ValueOf(target)).FieldByName(curFieldName)
					if f.IsValid() {
						n := curFieldValue.Len()
						tObj := curFieldValue
						if n > 0 && curFieldValue.Index(0).Kind() == reflect.Struct {
							tObj = reflect.MakeSlice(f.Type(), n, n)
							for i := 0; i < curFieldValue.Len(); i++ {
								curV := curFieldValue.Index(i)
								if curV.IsValid() {
									tField := reflect.New(tObj.Index(i).Type())
									if err = SetModel(curV.Interface(), tField.Interface()); err != nil {
										break
									}
									tObj.Index(i).Set(tField.Elem())
								}
							}
						}
						f.Set(tObj)
					}
				} else {
					f := reflect.Indirect(reflect.ValueOf(target)).FieldByName(curFieldName)
					if f.IsValid() {
						f.Set(curFieldValue)
					}
				}
			}
		} else {
			err = perror.CustomError("source should be a struct but found", refObj.Kind().String())
		}
	} else {
		err = perror.CustomError("target should be an address of struct")
	}
	return
}

//Unmarshal : unmarshal data and wrap error
func Unmarshal(data []byte, structModel interface{}) (err error) {
	if err = jsoniter.Unmarshal(data, structModel); err != nil {
		err = perror.UnmarshalError(err)
	}
	return
}

//Marshal : marshal model and wrap error
func Marshal(structModel interface{}) (data []byte, err error) {
	if data, err = jsoniter.Marshal(structModel); err != nil {
		err = perror.MarshalError(err)
	}
	return
}

//GetStructName will return sturct name
func GetStructName(model interface{}) (name string) {
	refObj := reflect.ValueOf(model)
	if refObj.Kind() == reflect.Struct {
		name = refObj.Type().Name()
	}
	return
}

//GetUpdateValue : Get Column Values which need to be updated
// --tag: default:true will allow default value update
// --tag: default:null will set null value if default value is given
// --tag: append:true will append map value in place of replacing it
func GetUpdateValue(input interface{}) (
	updateTag string, updateValue []interface{}, err error) {
	var fieldVal interface{}
	refObj := reflect.ValueOf(input)
	if refObj.Kind() == reflect.Struct {
		for i := 0; i < refObj.NumField(); i++ {
			refObjField := refObj.Type().Field(i)
			if refObjField.Name[0] > 'Z' || refObjField.Anonymous {
				continue
			}

			if tagValue, exists := refObjField.Tag.
				Lookup("sql"); exists == true {
				if tagValue == "-" {
					continue
				}
				dVal, dExists := refObjField.Tag.Lookup("default")
				aVal, aExists := refObjField.Tag.Lookup("append")
				curFieldVal := refObj.FieldByName(refObjField.Name)
				isDefaultVal := IsDefaultVal(curFieldVal)
				if isDefaultVal == false || dExists == true {
					tagFields := strings.Split(tagValue, ",")
					if len(tagFields) > 0 {
						if aExists && strings.ToLower(aVal) == "true" && curFieldVal.Kind() == reflect.Map {
							updateTag += tagFields[0] + ` = ` + tagFields[0] + ` || ?,`
						} else {
							updateTag += tagFields[0] + ` = ?,`
						}
						if strings.ToLower(dVal) == "null" && isDefaultVal {
							fieldVal = nil
						} else if fieldVal, err = GetFieldVal(curFieldVal); err != nil {
							break
						}
						updateValue = append(updateValue, fieldVal)
					}
				}
			}
		}
	}
	updateTag = strings.TrimSuffix(updateTag, ",")
	return
}

//GetUpdateValueSpecificUpdate : Update only those columns which needs to be updated
func GetUpdateValueSpecificUpdate(input interface{}, inputModel interface{}, reqBody string) (
	updateTag string, updateValue []interface{}, err error) {
	var (
		fieldVal                  interface{}
		updatedParams, reqBodyMap map[string]interface{}
	)
	refObj := reflect.ValueOf(input)
	if refObj.Kind() == reflect.Struct {
		updatedParams = make(map[string]interface{}, 0)
		if err = Unmarshal([]byte(reqBody), &reqBodyMap); err == nil {
			JsonToFieldName(inputModel, reqBodyMap, updatedParams)
			for i := 0; i < refObj.NumField(); i++ {
				refObjField := refObj.Type().Field(i)
				if refObjField.Name[0] > 'Z' || refObjField.Anonymous {
					continue
				}

				if tagValue, exists := refObjField.Tag.
					Lookup("sql"); exists == true {
					if tagValue == "-" {
						continue
					}
					dVal, dExists := refObjField.Tag.Lookup("default")
					aVal, aExists := refObjField.Tag.Lookup("append")
					curFieldVal := refObj.FieldByName(refObjField.Name)
					isDefaultVal := IsDefaultVal(curFieldVal)
					_, ok := updatedParams[refObjField.Name]
					if isDefaultVal == false || dExists == true || ok == true {
						tagFields := strings.Split(tagValue, ",")
						if len(tagFields) > 0 {
							if aExists && strings.ToLower(aVal) == "true" && curFieldVal.Kind() == reflect.Map {
								updateTag += tagFields[0] + ` = ` + tagFields[0] + ` || ?,`
							} else {
								updateTag += tagFields[0] + ` = ?,`
							}
							if strings.ToLower(dVal) == "null" && isDefaultVal {
								fieldVal = nil
							} else if fieldVal, err = GetFieldVal(curFieldVal); err != nil {
								break
							}
							updateValue = append(updateValue, fieldVal)
						}
					}
				}
			}
		}
	}
	updateTag = strings.TrimSuffix(updateTag, ",")
	return
}

//JsonToFieldName: Compares JSON fields of models and converts it to VariableNames
func JsonToFieldName(inputModel interface{}, reqBody map[string]interface{}, body map[string]interface{}) {
	refObj := reflect.ValueOf(inputModel)
	if refObj.Kind() == reflect.Struct {
		for i := 0; i < refObj.NumField(); i++ {
			refObjField := refObj.Type().Field(i)
			jsonVal, _ := refObjField.Tag.Lookup("json")
			if refObj.Field(i).Kind() == reflect.Struct {
				JsonToFieldName(refObj.Field(i).Interface(), reqBody, body)
			}
			_, ok := reqBody[jsonVal]
			if ok {
				body[refObjField.Name] = struct{}{}
			}
		}
	}
	return
}

//GetUpdateTagVal will return update tags and its values
func GetUpdateTagVal(input interface{}) (tag []string, val []interface{}, err error) {
	var fieldVal interface{}
	refObj := reflect.ValueOf(input)
	if refObj.Kind() == reflect.Struct {
		for i := 0; i < refObj.NumField(); i++ {
			refObjField := refObj.Type().Field(i)
			if refObjField.Name[0] > 'Z' || refObjField.Anonymous {
				continue
			}

			if tagValue, exists := refObjField.Tag.Lookup("sql"); exists == true {
				if tagValue == "-" {
					continue
				}
				dVal, dExists := refObjField.Tag.Lookup("default")
				curFieldVal := refObj.FieldByName(refObjField.Name)
				isDefaultVal := IsDefaultVal(curFieldVal)

				if isDefaultVal == false || dExists == true {
					tagFields := strings.Split(tagValue, ",")
					if len(tagFields) > 0 {
						tag = append(tag, tagFields[0])
						if strings.ToLower(dVal) == "null" && isDefaultVal {
							fieldVal = nil
						} else if fieldVal, err = GetFieldVal(curFieldVal); err != nil {
							break
						}
						val = append(val, fieldVal)
					}
				}
			}
		}
	}
	return
}

//IsDefaultVal Check whether zero value or not
func IsDefaultVal(curFieldVal reflect.Value) (isDefault bool) {
	defaultValue := reflect.Zero(curFieldVal.Type())
	fieldKind := curFieldVal.Kind()
	if curTime, isTimeType := curFieldVal.Interface().(time.Time); isTimeType {
		zeroTime, _ := defaultValue.Interface().(time.Time)
		isDefault = (curTime.After(zeroTime) == false)
	} else if fieldKind == reflect.Struct || fieldKind == reflect.Ptr {
		isDefault = reflect.DeepEqual(curFieldVal, defaultValue)
	} else if fieldKind == reflect.Map || fieldKind == reflect.Slice {
		if curFieldVal.Len() == 0 {
			isDefault = true
		}
	} else if curFieldVal.Interface() == defaultValue.Interface() {
		isDefault = true
	}
	return
}

//GetFieldVal Convert reflect value to its corrosponding data type
func GetFieldVal(val reflect.Value) (castValue interface{}, err error) {
	switch val.Kind() {
	case reflect.String:
		castValue = val.String()
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
		castValue = val.Int()
	case reflect.Float32, reflect.Float64:
		castValue = val.Float()
	case reflect.Map, reflect.Slice, reflect.Struct, reflect.Interface:
		castValue = val.Interface()
	case reflect.Bool:
		castValue = val.Bool()
	default:
		err = perror.CustomError("Invalid datatype:", val.Kind().String())
	}
	return
}

//GetPriorityValue : get the value by priority
func GetPriorityValue(a ...interface{}) (pVal interface{}) {
	for _, curVal := range a {
		if reflect.ValueOf(curVal).IsValid() {
			pVal = reflect.Zero(reflect.TypeOf(curVal)).Interface()
			if curVal != pVal {
				pVal = curVal
				break
			}
		}
	}
	return
}

//DoExistInFilter will check if any one of the input exists in filter
func DoExistInFilter(filter map[string]FilterDet, key ...string) (exists bool) {
	for _, curKey := range key {
		if _, exists = filter[curKey]; exists {
			break
		}
	}
	return
}

//RandString will return random string
func RandString(n int) string {
	b := make([]rune, n)
	max := len(letterRunes)
	for i := range b {
		b[i] = letterRunes[rand.Intn(max)]
	}
	return string(b)
}

//RandNum will return random number
func RandNum(n int) string {
	b := make([]rune, n)
	max := len(numRunes)
	for i := range b {
		b[i] = numRunes[rand.Intn(max)]
	}
	return string(b)
}

//FloatRound : Float round off
func FloatRound(x float64, decimals int) float64 {
	p := 0
	if decimals >= 0 {
		p = decimals
	}
	ipStr := strconv.FormatFloat(x, 'f', -1, 64)
	sl := strings.Split(ipStr, ".")
	if len(sl) == 2 && len(sl[1]) > p {
		newStr := sl[0] + sl[1][:p] + "." + sl[1][p:]
		a, _ := strconv.ParseFloat(newStr, 64)
		_, rem := math.Modf(a)
		if (math.Abs(rem) >= 0.5 && rem > 0) || (math.Abs(rem) < 0.5 && rem < 0) {
			a = math.Ceil(a)
		} else {
			a = math.Floor(a)
		}
		return a / math.Pow(10, float64(p))
	} else {
		return x
	}
}

//FloatFloorRound : Float Floor round off
func FloatFloorRound(x float64, decimals int) float64 {
	p := 0
	if decimals >= 0 {
		p = decimals
	}
	ipStr := strconv.FormatFloat(x, 'f', -1, 64)
	sl := strings.Split(ipStr, ".")
	if len(sl) == 2 && len(sl[1]) > p {
		newStr := sl[0] + sl[1][:p] + "." + sl[1][p:]
		a, _ := strconv.ParseFloat(newStr, 64)
		a = math.Floor(a)
		return a / math.Pow(10, float64(p))
	} else {
		return x
	}
}

//ReverseStrSlice will reverse String Slice
func ReverseStrSlice(s []string) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

//SnakeName will capitalized name by space separated
func SnakeName(name string) (capName string) {
	capName = ""
	for i := range name {
		v := name[i]
		if v <= 'Z' {
			if i != 0 {
				capName += " "
			}
		} else if i == 0 {
			v = v - 32
		}
		capName += string(v)
	}
	return
}

//GetFetchField will return fetch fields if filter exists
func GetFetchField(curFields string, filter map[string]FilterDet) (fFields string) {
	if fVal, exists := filter[FetchFields]; exists && fVal.Value != "" {
		for _, curField := range strings.Split(fVal.Value, ",") {
			fFields += curField + ","
		}
		fFields = strings.TrimSuffix(fFields, ",")
	} else {
		fFields = curFields
	}
	return
}

//GetFetchFieldNParam will return fetch fields and params if filter exists
func GetFetchFieldNParam(curFields string, curParams []interface{}, filter map[string]FilterDet) (fFields string, params []interface{}) {
	if fVal, exists := filter[FetchFields]; exists && fVal.Value != "" {
		for _, curField := range strings.Split(fVal.Value, ",") {
			fFields += curField + ","
		}
		fFields = strings.TrimSuffix(fFields, ",")
		params = make([]interface{}, 0)
	} else {
		fFields = curFields
		params = curParams
	}
	return
}

//DayLogger will create new day wise log file
func DayLogger(file string) (fp *os.File, err error) {
	sTime := time.Now().In(loc)
	file = strings.TrimSuffix(file, ".log")
	path := fmt.Sprintf("%s_%d-%d-%d.log", file, sTime.Day(), sTime.Month(), sTime.Year())
	if fp, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666); err != nil {
		err = perror.MiscError(err, file, "file creation error")
	}
	return
}

//StackTrace : Get function name, file name and line no of the caller function
//Depth is the value from which it will start searching in the stack
func StackTrace(depth int) (funcName string, file string, line int) {
	var (
		ok bool
		pc uintptr
	)
	for i := depth; ; i++ {
		if pc, file, line, ok = runtime.Caller(i); ok {
			if strings.Contains(file, packageName) {
				continue
			}
			fileName := strings.Split(file, "github.com")
			if len(fileName) > 1 {
				file = fileName[1]
			}
			_, funcName = packageFuncName(pc)
			break
		} else {
			break
		}
	}
	return
}

//packageFuncName : Package and function name from package counter
func packageFuncName(pc uintptr) (packageName string, funcName string) {
	if f := runtime.FuncForPC(pc); f != nil {
		funcName = f.Name()
		if ind := strings.LastIndex(funcName, "/"); ind > 0 {
			packageName += funcName[:ind+1]
			funcName = funcName[ind+1:]
		}
		if ind := strings.Index(funcName, "."); ind > 0 {
			packageName += funcName[:ind]
			funcName = funcName[ind+1:]
		}
	}
	return
}

//GetTraceFile will return stack trace file
func GetTraceFile(n int) string {
	fn, file, line := StackTrace(n)
	return fmt.Sprintf("file %v:%v func %v", file, line, fn)
}

//GetTraceFile will return stack trace file without depth
func GetTraceFileWithoutDepth() string {
	fn, file, line := StackTraceWithoutDepth()
	return fmt.Sprintf("file %v:%v func %v", file, line, fn)
}

//StackTraceWithoutDepth : It will trace stack where it comes out from vendor
func StackTraceWithoutDepth() (funcName string, file string, line int) {
	var (
		ok bool
		pc uintptr
		i  int
	)
	i = 1
	for {
		if pc, file, line, ok = runtime.Caller(i); ok {
			i++
			if strings.Contains(file, "vendor") {
				continue
			}
			fileName := strings.Split(file, "github.com")
			if len(fileName) > 1 {
				file = fileName[1]
			}
			_, funcName = packageFuncName(pc)
			break
		} else {
			break
		}
	}
	return
}
