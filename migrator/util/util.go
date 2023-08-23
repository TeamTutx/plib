package util

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/go-pg/pg/v10"
	"gitlab.com/g-harshit/plib/conf"
	"gitlab.com/g-harshit/plib/migrator/model"
)

//variables
var (
	TableMap map[string]interface{}
	EnumList map[string][]string
	QueryFp  *os.File
)

//GetColumnSchema : Get Column Schema of given table
func GetColumnSchema(conn *pg.DB, tableName string) (columnSchema []model.ColumnSchema, err error) {
	schemaQuery := `SELECT column_name,column_default, data_type, 
	udt_name, is_nullable,character_maximum_length 
	FROM information_schema.columns WHERE table_name = ?;`
	_, err = conn.Query(&columnSchema, schemaQuery, tableName)
	return
}

//GetConstraint : Get Constraint of table from database
func GetConstraint(conn *pg.DB, tableName string) (constraint []model.ColumnSchema, err error) {
	constraintQuery := `SELECT tc.constraint_type,
    tc.constraint_name, tc.is_deferrable, tc.initially_deferred, 
    kcu.column_name AS column_name, ccu.table_name AS foreign_table_name, 
    ccu.column_name AS foreign_column_name, pgc.confupdtype, pgc.confdeltype  
    FROM 
    information_schema.table_constraints AS tc 
    JOIN information_schema.key_column_usage AS kcu 
    ON tc.constraint_name = kcu.constraint_name 
    JOIN information_schema.constraint_column_usage AS ccu 
    ON ccu.constraint_name = tc.constraint_name 
    JOIN pg_constraint AS pgc ON pgc.conname = tc.constraint_name AND 
    conrelid=?::regclass::oid WHERE tc.constraint_type 
    IN('FOREIGN KEY','PRIMARY KEY','UNIQUE') AND tc.table_name = ?
    AND array_length(pgc.conkey,1) = 1;`
	_, err = conn.Query(&constraint, constraintQuery, tableName, tableName)
	return
}

//GetCompositeUniqueKey : Get composite unique key name and columns
func GetCompositeUniqueKey(conn *pg.DB, tableName string) (uniqueKeySchema []model.UniqueKeySchema, err error) {
	uniqueKeyQuery := `select string_agg(c.column_name,',') as col, pgc.conname 
	from pg_constraint as pgc join
	information_schema.table_constraints tc on pgc.conname = tc.constraint_name, 
	unnest(pgc.conkey::int[]) as colNo join information_schema.columns as c 
	on c.ordinal_position = colNo and c.table_name = ? 
	where array_length(pgc.conkey,1)>1 and pgc.contype='u'
	and pgc.conrelid=c.table_name::regclass::oid group by pgc.conname;`
	_, err = conn.Query(&uniqueKeySchema, uniqueKeyQuery, tableName)
	return
}

//EnumExists : Check if Enum Type Exists in database
func EnumExists(conn *pg.DB, enumName string) (flag bool) {
	var num int
	enumSQL := `SELECT 1 FROM pg_type WHERE typname = ?;`
	if _, err := conn.Query(pg.Scan(&num), enumSQL, enumName); err == nil && num == 1 {
		flag = true
	}
	return
}

//TableExists : Check if Table Exists in database
func TableExists(conn *pg.DB, tableName string) (flag bool) {
	var num int
	enumSQL := `SELECT 1 FROM pg_tables WHERE tablename = ?;`
	if _, err := conn.Query(pg.Scan(&num), enumSQL, tableName); err == nil && num == 1 {
		flag = true
	}
	return
}

//MergeColumnConstraint : Merge Table Schema with Constraint
func MergeColumnConstraint(columnSchema,
	constraint []model.ColumnSchema) map[string]model.ColumnSchema {
	constraintMap := make(map[string]model.ColumnSchema)
	tableSchema := make(map[string]model.ColumnSchema)
	for _, curConstraint := range constraint {
		constraintMap[curConstraint.ColumnName] = curConstraint
	}
	for _, curColumnSchema := range columnSchema {
		if curConstraint, exists :=
			constraintMap[curColumnSchema.ColumnName]; exists == true {
			curColumnSchema.ConstraintType = curConstraint.ConstraintType
			curColumnSchema.ConstraintName = curConstraint.ConstraintName
			curColumnSchema.IsDeferrable = curConstraint.IsDeferrable
			curColumnSchema.InitiallyDeferred = curConstraint.InitiallyDeferred
			curColumnSchema.ForeignTableName = curConstraint.ForeignTableName
			curColumnSchema.ForeignColumnName = curConstraint.ForeignColumnName
			curColumnSchema.UpdateType = curConstraint.UpdateType
			curColumnSchema.DeleteType = curConstraint.DeleteType
		}
		tableSchema[curColumnSchema.ColumnName] = curColumnSchema
	}
	return tableSchema
}

//IsSkip will check table contain skip tags
func IsSkip(tableName string) (flag bool) {
	tableModel, isValid := TableMap[tableName]
	if isValid {
		flag = SkipTag(tableModel)
	}
	return
}

//SkipTag will check skiptag exists in model or not
func SkipTag(object interface{}) (flag bool) {
	refObj := reflect.ValueOf(object).Elem()
	if refObj.Kind() == reflect.Struct {
		if refObj.NumField() > 0 {
			if tag, exists := refObj.Type().Field(0).Tag.Lookup("ty"); exists && tag == "skip" {
				flag = true
			}
		}
	}
	return
}

func mergeMap(a, b map[reflect.Value]reflect.StructField) {
	for k, v := range b {
		a[k] = v
	}
}

//GetStructField will return struct fields
func GetStructField(model interface{}) (fields map[reflect.Value]reflect.StructField) {
	refObj := reflect.ValueOf(model)
	fields = make(map[reflect.Value]reflect.StructField)
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
				embdFields := GetStructField(refField.Interface())
				mergeMap(fields, embdFields)
			} else {
				if _, exists := refType.Tag.Lookup("pg"); exists == false {
					fmt.Println("No SQL tag in", refType.Name)
					panic("sql tag not fround")
				}
				fields[refField] = refType
			}
		}
	}
	return
}

//getSQLTag will return sql tag
func getSQLTag(refField reflect.StructField) (sqlTag string) {
	sqlTag = refField.Tag.Get("pg")
	sqlTag = strings.ToLower(sqlTag)
	return
}

//FieldType will return field type
func FieldType(refField reflect.StructField) (fType string) {
	sqlTag := getSQLTag(refField)
	vals := strings.Split(sqlTag, ",type:")
	if len(vals) > 1 {
		fType = vals[1]
		fType = strings.Trim(strings.Split(fType, " ")[0], " ")
	}
	return
}

//RefTable will reutrn reference table
func RefTable(refField reflect.StructField) (refTable string) {
	sqlTag := getSQLTag(refField)
	refTag := strings.Split(sqlTag, "references")
	if len(refTag) > 1 {
		refTable = strings.Split(refTag[1], "(")[0]
		refTable = strings.Trim(refTable, " ")
	}
	return
}

//GetIndex will return index fields of struct
func GetIndex(tableName string) (idx map[string]string) {
	dbModel := TableMap[tableName]
	refObj := reflect.ValueOf(dbModel)
	m := refObj.MethodByName("Index")
	if m.IsValid() {
		out := m.Call([]reflect.Value{})
		if len(out) > 0 && out[0].Kind() == reflect.Map {
			idx = out[0].Interface().(map[string]string)
		}
	}
	return
}

//GetUniqueKey will return unique key fields of struct
func GetUniqueKey(tableName string) (uk map[string]string) {
	dbModel := TableMap[tableName]
	refObj := reflect.ValueOf(dbModel)
	m := refObj.MethodByName("UniqueKey")
	if m.IsValid() {
		out := m.Call([]reflect.Value{})
		if len(out) > 0 && out[0].Kind() == reflect.Slice {
			val := out[0].Interface().([]string)
			uk = make(map[string]string)
			for i, ukFields := range val {
				ukName := fmt.Sprintf("uk_%v_%d", tableName, i+1)
				uk[ukName] = ukFields
			}
		}
	}
	return
}

//GetChoice will return user Choice
func GetChoice(sql string, skipPrompt string) (choice string) {
	choice = YES_CHOICE
	if skipPrompt == "" {
		fmt.Printf("%v\nWant to continue (y/n): ", sql)
		fmt.Scan(&choice)
	}
	return
}

//SetTableMap will set table map
func SetTableMap(tables []interface{}) {
	TableMap = make(map[string]interface{})
	for _, curModel := range tables {
		refObj := reflect.ValueOf(curModel)
		if refObj.Kind() != reflect.Ptr || refObj.Elem().Kind() != reflect.Struct {
			panic("invalid struct pointer")
		} else {
			refObj = refObj.Elem()
			if field, exists := refObj.Type().FieldByName("tableName"); exists {
				tableName := field.Tag.Get("pg")
				TableMap[tableName] = curModel
			} else {
				panic("tableName field not found")
			}
		}
	}
}

//InitLogger : Init File logger
func InitLogger() (err error) {
	var fp *os.File
	sTime := time.Now()
	path := conf.String(DATABASE_ALTER_LOG, "glacier.alter")
	queryPath := conf.String(DATABASE_ALTER_QUERY_LOG, "glacier.alter")
	path = fmt.Sprintf("%s_%d-%d-%d.log", path, sTime.Day(), sTime.Month(), sTime.Year())
	queryPath = fmt.Sprintf("%s_%d-%d-%d.log", queryPath, sTime.Day(), sTime.Month(), sTime.Year())
	if fp, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666); err != nil {
		fmt.Println("Not Able to init logger", err.Error())
		return
	} else if QueryFp, err = os.OpenFile(queryPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666); err != nil {
		fmt.Println("Not Able to init logger", err.Error())
		return
	}
	log.SetOutput(fp)
	return
}

//GetFileData : will return file data
func GetFileData(files []string) (data string, err error) {
	for _, curFile := range files {
		var (
			bData []byte
			dir   string
		)
		if dir, err = os.Getwd(); err == nil {
			if bData, err = ioutil.ReadFile(dir + "/" + curFile); err == nil {
				data += `
				` + string(bData)
			}
		}
		if err != nil {
			break
		}
	}
	return
}
