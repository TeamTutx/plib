package ally

import (
	"reflect"
	"strings"

	"gitlab.com/g-harshit/plib/perror"
)

//MergeMap will merge two maps
func MergeMap(a, b map[string]interface{}) {
	for k, v := range b {
		a[k] = v
	}
}

//StructTagValue will return struct model tag field and value from specific tag
func StructTagValue(input interface{}, tag string) (fields map[string]interface{}, err error) {
	if tag == "" {
		tag = "sql"
	}
	refObj := reflect.ValueOf(input)
	fields = make(map[string]interface{})
	if refObj.Kind() == reflect.Ptr {
		refObj = refObj.Elem()
	}
	if refObj.IsValid() {
		for i := 0; i < refObj.NumField(); i++ {
			refField := refObj.Field(i)
			refType := refObj.Type().Field(i)
			if refType.Name[0] > 'Z' {
				continue
			}
			if refType.Anonymous && refField.Kind() == reflect.Struct {
				var embdFields map[string]interface{}
				if embdFields, err = StructTagValue(refField.Interface(), tag); err != nil {
					break
				}
				MergeMap(fields, embdFields)
			} else {
				if col, exists := refType.Tag.Lookup(tag); exists {
					isDef := IsDefaultVal(refField)
					if col == "-" ||
						(strings.Contains(col, ",omitempty") && isDef) {
						continue
					}
					if tag == "gorm" {
						if sqlCol := refType.Tag.Get("sql"); sqlCol == "-" {
							continue
						}
						col = strings.TrimPrefix(col, "column:")
					}
					col = strings.Split(col, ",")[0]
					dVal := refType.Tag.Get("default")
					if dVal == "null" && isDef {
						fields[col] = nil
					} else if fields[col], err = GetFieldVal(refField); err != nil {
						break
					}
				}
			}
		}
	}
	return
}

//MySQLFieldVal will reutrn mysql fields value
func MySQLFieldVal(model interface{}, cols []string) (
	field string, value []interface{}, err error) {

	refObj := reflect.ValueOf(model)

	if refObj.Kind() == reflect.Struct {
		field, value, err = getUpdateField(model, cols)
	} else if refObj.Kind() == reflect.Slice {
		for i := 0; i < refObj.Len(); i++ {
			var (
				curField string
				curValue []interface{}
			)
			if curField, curValue, err = getUpdateField(refObj.Index(i).Interface(), cols); err != nil {
				break
			}
			field += curField + ","
			value = append(value, curValue...)
		}
		if err == nil {
			field = strings.TrimSuffix(field, ",")
		}
	}
	return
}

//MySQLInsertSQL will return mysql insert query
func MySQLInsertSQL(tableName string, model interface{}, cols []string) (
	insertSQL string, params []interface{}, err error) {

	var field string
	if field, params, err = MySQLFieldVal(model, cols); err == nil {
		insertSQL = "INSERT INTO " + tableName + " (`" + strings.Join(cols, "`,`") + "`) VALUES " + field
	}
	return
}

//MySQLInsertIgnoreSQL will return mysql insert ignore query
func MySQLInsertIgnoreSQL(tableName string, model interface{}, cols []string) (
	insertSQL string, params []interface{}, err error) {

	var field string
	if field, params, err = MySQLFieldVal(model, cols); err == nil {
		insertSQL = "INSERT IGNORE INTO " + tableName + " (`" + strings.Join(cols, "`,`") + "`) VALUES " + field
	}
	return
}

//MySQLUpdateVal will return mysql update fields value
func MySQLUpdateVal(model interface{}, cols []string) (
	fieldVal map[string]interface{}, err error) {

	var allField map[string]interface{}
	fieldVal = make(map[string]interface{})
	if allField, err = StructTagValue(model, "gorm"); err == nil {
		for _, col := range cols {
			if colVal, exists := allField[col]; exists {
				fieldVal[col] = colVal
			}
		}
		if len(fieldVal) == 0 {
			err = perror.CustomError("No Field to update")
		}
	}
	return
}

//getUpdateField will return updated field tag and params
func getUpdateField(model interface{}, cols []string) (
	field string, value []interface{}, err error) {

	var allField map[string]interface{}
	if allField, err = StructTagValue(model, "gorm"); err == nil {
		field = "("
		for _, col := range cols {
			col = strings.Trim(col, " ")
			if colVal, exists := allField[col]; exists {
				field += "?,"
				value = append(value, colVal)
			}
		}
		field = strings.TrimSuffix(field, ",")
		if len(value) > 0 {
			field += ")"
		} else {
			err = perror.CustomError("No Field to update")
		}
	}
	return
}

//GetBulkUpdateSQL will create bulk update pg sql
func GetBulkUpdateSQL(model []interface{}) (sql string, tags []string, params []interface{}, err error) {
	sql = `
	FROM (VALUES `
	var (
		ques string
	)
	for _, curModel := range model {
		var updateValue []interface{}
		if tags, updateValue, err = GetUpdateTagVal(curModel); err != nil {
			break
		}
		ques = strings.TrimSuffix(strings.Repeat("?, ", len(updateValue)), ", ")
		sql += `(` + ques + `), `
		params = append(params, updateValue...)
	}
	sql = strings.TrimSuffix(sql, ", ")
	sql += `) AS
	_data("` + strings.Join(tags, `", "`) + `")`
	return
}
