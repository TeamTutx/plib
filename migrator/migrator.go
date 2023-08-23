package migrator

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	shifter "gitlab.com/g-harshit/pg-shifter"
	"gitlab.com/g-harshit/plib/ally"
	"gitlab.com/g-harshit/plib/database/postgresql"
	"gitlab.com/g-harshit/plib/migrator/model"
	"gitlab.com/g-harshit/plib/migrator/util"
)

//variables
var (
	tableCreated         = make(map[interface{}]bool)
	enumCreated          = make(map[interface{}]struct{})
	structToTableDataMap = make(map[string]string)
	addedColumn          = make(map[string]string)
	removedColumn        = make(map[string]string)
	alteredColumn        = make(map[string]string)
	alteredConstraint    = make(map[string]string)
	structFields         = make(map[reflect.Value]reflect.StructField)
	PostExecutionSQL     string
	PostExeFiles         []string
	InitQueryExecuted    bool
	PreExecutionSQL      string
	PreExeFiles          []string
)

//SetMigrator will set tables and enums
func SetMigrator(tables []interface{}, enumList map[string][]string) {
	util.SetTableMap(tables)
	util.EnumList = enumList
}

//AddPostExeFile will add post execution files
func AddPostExeFile(files ...string) {
	PostExeFiles = append(PostExeFiles, files...)
}

//AddPreExeFile will add pre execution files
func AddPreExeFile(files ...string) {
	PreExeFiles = append(PreExeFiles, files...)
}

//Run will execute migrator
func Run() {

	tableName := flag.String("t", "",
		"Table name which you want to create (, seperated if multiple tables and all for all tables)")
	enumType := flag.String("en", "", "Enum Name")
	uniqueKey := flag.String("uk", "", "Unique Key name (, seperated if multiple unique keys/ all for all unique keys)")
	index := flag.String("idx", "", "Index name (, seperated if multiple index/ all for all index)")
	skipPrompt := flag.String("sp", "", "1 to skip [unique key and index create] prompt")
	raw := flag.Int("r", 0, "1 if you want to execute raw query")
	withDependency := flag.Int("dep", 0, "Database creation with depencencies")
	alter := flag.Int("alt", 0, "alter flag 1 to Alter Table")
	old := flag.Int("old", 0, "1 to user old scrpit to alter")
	flag.Parse()
	util.InitLogger()

	util.GetFileData(PostExeFiles)

	var fileSQL string

	if validateInput(*tableName, *enumType, *uniqueKey, *index, *raw) == true {
		if conn, err := postgresql.Conn(true); err == nil {

			//pg-shifter added
			s := shifter.NewShifter()
			for _, curModel := range util.TableMap {
				if err = s.SetTableModel(curModel); err != nil {
					fmt.Println(err)
					return
				}
			}
			s.SetEnum(util.EnumList)
			//pg-shifter added

			if *raw == 0 {
				if *tableName == util.ALL {

					if InitQueryExecuted == false && (PreExecutionSQL != "" || len(PreExeFiles) > 0) {
						if fileSQL, err = util.GetFileData(PreExeFiles); err != nil {
							fmt.Println("Pre Execution File Error", err)
							return
						}
						PreExecutionSQL += fileSQL
						if err = executeRaw(conn, PreExecutionSQL); err != nil {
							return
						}
					}

					*skipPrompt = "1"

					for curTableName := range util.TableMap {
						if err = create(conn, curTableName, 1, *alter, *enumType, util.ALL, util.ALL, *skipPrompt, *old, s); err != nil {
							fmt.Println("Create Error", err)
							return
						}
					}
					if *index == "" && *uniqueKey == "" && *alter == 0 {
						if InitQueryExecuted == false && (PostExecutionSQL != "" || len(PostExeFiles) > 0) {
							if fileSQL, err = util.GetFileData(PostExeFiles); err != nil {
								fmt.Println("Post Execution File Error", err)
								return
							}
							PostExecutionSQL += fileSQL
							executeRaw(conn, PostExecutionSQL)
						}
					}
				} else {
					tableNameSlice := strings.Split(*tableName, ",")
					for _, curTableName := range tableNameSlice {
						if err = create(conn, curTableName, *withDependency, *alter, *enumType, *uniqueKey, *index, *skipPrompt, *old, s); err != nil {
							fmt.Println(err)
						}
					}
				}
				if *enumType == "" && *alter == 0 && *index == "" {
					fmt.Printf("%v tables created\n", len(tableCreated))
				}
			} else {
				var (
					query  string
					choice string
				)
				fmt.Println("Enter the Query between tilt (~): ")
				fmt.Scanf("%q", &query)
				fmt.Printf("Are you sure! You want to execute:\n\n%v\n\n(y/n): ", query)
				if fmt.Scan(&choice); choice == util.YES_CHOICE {
					executeRaw(conn, query)
				}
			}
		} else {
			fmt.Println("Database Connectivity Error: ", err.Error())
		}
	}
}

//Execute raw query
func executeRaw(conn *pg.DB, query string) (err error) {
	if affected, err := conn.Exec(query); err == nil {
		util.QueryFp.WriteString(fmt.Sprintf("-- Raw Execution:\n%v\n", query))
		fmt.Println("Row Affected:", affected.RowsAffected())
	} else {
		fmt.Println("Raw query Execution Error:", err.Error())
	}
	return
}

//Validate input values
func validateInput(tableName, enumType, uniqueKey, index string, raw int) (flag bool) {
	if raw == 1 {
		flag = true
	} else {
		noOfTable := len(strings.Split(tableName, ","))
		if enumType == "" || (tableName != util.ALL && noOfTable == 1) {
			flag = true
		} else {
			fmt.Println("Invalid Input: To create or alter enum, provide one table name only")
		}
		if (uniqueKey == util.ALL && tableName == util.ALL) ||
			(noOfTable == 1 && tableName != util.ALL) ||
			(tableName == util.ALL && uniqueKey == "") {
			flag = true
		} else {
			fmt.Println("Invalid Input: Create all unique key of all table or all unique key of table one by one")
		}
		if (index == util.ALL && tableName == util.ALL) ||
			(noOfTable == 1 && tableName != util.ALL) ||
			(tableName == util.ALL && index == "") {
			flag = true
		} else {
			fmt.Println("Invalid Input: Create all index of all table or all index of table one by one")
		}
	}
	return
}

//Initiallize struct datatype to database datatype mapping
func initStructTableMap() {
	structToTableDataMap = map[string]string{
		"smallint":    "int2",
		"int":         "int4",
		"integer":     "int4",
		"serial":      "int4",
		"bigint":      "int8",
		"numeric":     "numeric",
		"varchar":     "varchar",
		"text":        "text",
		"citext":      "citext",
		"jsonb":       "jsonb",
		"timestamp":   "timestamp",
		"timestamptz": "timestamptz",
		"boolean":     "bool",
		"bool":        "bool",
	}
}

//Create table or alter table based on alter flag
func create(conn *pg.DB, tableName string, withDependency int,
	alter int, enumName string, uniqueKey string, index string, skipPrompt string, old int,
	s *shifter.Shifter) (err error) {

	if model, exists := util.TableMap[tableName]; exists {
		structFields = util.GetStructField(model)
		if enumName != "" {
			err = createEnumByName(conn, tableName, enumName)
		} else if alter == 0 {
			if old == 0 {
				err = createUsingShifter(conn, s, tableName)
			} else {
				err = createTable(conn, tableName, withDependency)
			}
		} else {
			if old == 0 {
				err = alterUsingShifter(conn, s, tableName)
			} else {
				err = alterTable(conn, tableName)
			}
		}
		if err == nil && uniqueKey != "" && old == 1 {
			err = createUniqueKey(conn, tableName, uniqueKey, skipPrompt)
		}
		if err == nil && index != "" {
			err = createIndex(conn, tableName, index, skipPrompt)
		}
		if err == nil {
			logAlterDetails(tableName)
		}
	} else {
		err = errors.New("Invalid Table Name")
	}
	return
}

//alterUsingShifter will use shifter to alter table
func alterUsingShifter(conn *pg.DB, s *shifter.Shifter, tableName string) (err error) {
	err = s.AlterTable(conn, tableName)
	return
}

//createUsingShifter will use shifter to create table
func createUsingShifter(conn *pg.DB, s *shifter.Shifter, tableName string) (err error) {
	err = s.CreateTable(conn, tableName)
	return
}

//Get unique key query by tablename, unique key constraing name and table columns
func getUniqueKeyQuery(tableName string, constraintName string,
	column string) (uniqueKeyQuery string) {
	return fmt.Sprintf("ALTER TABLE %v DROP CONSTRAINT IF EXISTS %v;\nALTER TABLE %v ADD CONSTRAINT %v UNIQUE (%v);\n",
		tableName, constraintName, tableName, constraintName, column)
}

//Create unique key of given table
func createUniqueKey(conn *pg.DB, tableName string, uniqueKey string, skipPrompt string) (err error) {

	var uniqueKeySQL string
	uk := util.GetUniqueKey(tableName)
	for ukName, ukFild := range uk {
		uniqueKeySQL += getUniqueKeyQuery(tableName, ukName, ukFild)
	}
	if uniqueKeySQL != "" {
		choice := util.GetChoice("UNIQUE KEY:\n"+uniqueKeySQL, skipPrompt)
		if choice == util.YES_CHOICE {
			if _, err = conn.Exec(uniqueKeySQL); err != nil {
				fmt.Println("Unique Key Constraint creation error: ", tableName, err.Error())
			} else {
				// util.QueryFp.WriteString(fmt.Sprintf("-- UNIQUE KEY\n%v\n", uniqueKeySQL))
				alteredConstraint[tableName] = uniqueKeySQL
			}
		}
	}
	return
}

//Get index query by tablename and table columns
func getIndexQuery(tableName string, indexDS string, column string) (uniqueKeyQuery string) {
	if strings.HasPrefix(indexDS, util.GIN_INDEX) == true {
		indexDS = util.GIN_INDEX
	} else {
		indexDS = util.B_TREE_INDEX
	}

	constraintName := fmt.Sprintf("idx_%v_%v", tableName, strings.Replace(strings.Replace(column, " ", "", -1), ",", "_", -1))
	return fmt.Sprintf("CREATE INDEX IF NOT EXISTS %v ON %v USING %v (%v);\n",
		constraintName, tableName, indexDS, column)
}

//Create index of given table
func createIndex(conn *pg.DB, tableName string, index string, skipPrompt string) (err error) {
	var indexSQL string
	for index, idxType := range util.GetIndex(tableName) {
		indexSQL += getIndexQuery(tableName, idxType, index)
	}
	if indexSQL != "" {
		choice := util.GetChoice("INDEX:\n"+indexSQL, skipPrompt)
		if choice == util.YES_CHOICE {
			if _, err = conn.Exec(indexSQL); err != nil {
				fmt.Println("Index Constraint creation error: ", tableName, err.Error())
			} else {
				// util.QueryFp.WriteString(fmt.Sprintf("-- INDEX\n%v\n", indexSQL))
			}
		}
	}
	return
}

//Get enum values by enumType
func getEnumValue(conn *pg.DB, enumType string) (enumValue []string, err error) {
	enumSQL := `SELECT e.enumlabel as enum_value
	  FROM pg_enum e
	  JOIN pg_type t ON e.enumtypid = t.oid
	  WHERE t.typname = ?;`
	_, err = conn.Query(&enumValue, enumSQL, enumType)
	return
}

//Create enum alter query by enumType and new values
func getEnumAlterQuery(enumName string, newValue []string) (enumAlterSQL string) {
	for _, curValue := range newValue {
		enumAlterSQL += fmt.Sprintf("ALTER type %v ADD VALUE '%v'; ", enumName, curValue)
	}
	return
}

//Compare enum values
func compareEnumValue(dbEnumVal, structEnumValue []string) (newValue []string) {
	var enumValueMap = make(map[string]struct{})
	for _, curEnumVal := range dbEnumVal {
		enumValueMap[curEnumVal] = struct{}{}
	}
	for _, curEnumVal := range structEnumValue {
		if _, exists := enumValueMap[curEnumVal]; exists == false {
			newValue = append(newValue, curEnumVal)
		}
	}
	return
}

//Create Enum in database
func createEnumByName(conn *pg.DB, tableName, enumName string) (err error) {

	var dbEnumVal []string
	if _, created := enumCreated[enumName]; created == false {
		if enumValue, exists := util.EnumList[enumName]; exists {
			if enumSQL, enumExists := getEnumQuery(conn, enumName, enumValue); enumExists == false {
				if _, err = conn.Exec(enumSQL); err == nil {
					enumCreated[enumName] = struct{}{}
					fmt.Printf("Enum %v created\n", enumName)
				} else {
					fmt.Println("createEnumByName: Enum Creation Error: ", tableName, err.Error())
				}
			} else {

				if dbEnumVal, err = getEnumValue(conn, enumName); err == nil {
					//comparing old and new enum values
					if newValue := compareEnumValue(dbEnumVal, enumValue); len(newValue) > 0 {
						var choice string
						enumAlterSQL := getEnumAlterQuery(enumName, newValue)
						fmt.Printf("%v\nWant to continue (y/n): ", enumAlterSQL)
						fmt.Scan(&choice)
						if choice == util.YES_CHOICE {
							if _, err = conn.Exec(enumAlterSQL); err == nil {
								util.QueryFp.WriteString(fmt.Sprintf("-- ALTER ENUM\n%v\n", enumAlterSQL))
								log.Println(fmt.Sprintf("----ALTER TABLE: %v", tableName))
								log.Println(fmt.Sprintf("ENUM TYPE MODIFIED:\t%v\nPREV VALUE:\t%v\nNEW VALUE:\t%v\n",
									enumName, dbEnumVal, enumValue))
							} else {
								fmt.Println("createEnumByName: Enum Alter Error: ", tableName, err.Error())
							}
						}
					}
				} else {
					fmt.Println("createEnumByName: Enum Value fetch Error: ", tableName, err.Error())
				}
			}
		} else {
			fmt.Printf("createEnumByName: Enum %v not found of table %v\n", enumName, tableName)
		}
	}
	return
}

//Create Enum in database
func createEnum(conn *pg.DB, tableName string) {
	tableModel := util.TableMap[tableName]
	fields := util.GetStructField(tableModel)
	for _, refFeild := range fields {
		fType := util.FieldType(refFeild)
		if _, exists := util.EnumList[fType]; exists {
			createEnumByName(conn, tableName, fType)
		}
	}
}

//Create Enum Query for given table
func getEnumQuery(conn *pg.DB, enumName string, enumValue []string) (query string, enumExists bool) {
	if enumExists = util.EnumExists(conn, enumName); enumExists == false {
		query += fmt.Sprintf("CREATE type %v AS ENUM('%v'); ",
			enumName, strings.Join(enumValue, "','"))
	}
	return
}

//Create trigger
func createTrigger(tx *pg.Tx, tableName string) (err error) {
	if util.IsSkip(tableName) == false {
		trigger := util.GetTrigger(tableName)
		// fmt.Println(trigger)
		if _, err = tx.Exec(trigger); err != nil {
			err = errors.New("createTrigger: trigger creation Error " + err.Error())
		} else {
			util.QueryFp.WriteString(fmt.Sprintf("-- TRIGGER\n%v\n", trigger))
		}
	}
	return
}

//Create history table
func createHistory(conn *pg.DB, tableName string) {
	if util.IsSkip(tableName) == false {
		historyTable := tableName + util.HISTORY_TAG
		if tableExists := util.TableExists(conn, historyTable); tableExists == false {
			if tx, err := conn.Begin(); err == nil {
				if _, err = tx.Exec(fmt.Sprintf("CREATE TABLE %v AS SELECT * FROM %v WHERE 1=2",
					historyTable, tableName)); err == nil {
					if _, err = tx.Exec(fmt.Sprintf(`
				   	ALTER TABLE %v DROP COLUMN IF EXISTS updated_at;
					ALTER TABLE %v ADD COLUMN id BIGSERIAL PRIMARY KEY;
					ALTER TABLE %v ADD COLUMN action VARCHAR(20);`,
						historyTable, historyTable, historyTable)); err == nil {
						if err = createTrigger(tx, tableName); err == nil {
							tx.Commit()
						} else {
							fmt.Println(err.Error())
							tx.Rollback()
						}
					} else {
						fmt.Println("createHistory: history table alter Error", err.Error())
						tx.Rollback()
					}
				} else {
					fmt.Println("createHistory: history table creation Error", err.Error())
				}
			} else {
				fmt.Println("createHistory: transaction start Error", err.Error())
			}
		}
	}
}

//Create Table in database
func createTable(conn *pg.DB, tableName string, withDependency int) (err error) {
	tableModel := util.TableMap[tableName]
	if _, alreadyCreated := tableCreated[tableModel]; alreadyCreated == false {
		tableCreated[tableModel] = true
		createEnum(conn, tableName)
		if withDependency == 1 {
			createTableDependencies(conn, tableModel)
		}
		if err = conn.Model(tableModel).CreateTable(
			&orm.CreateTableOptions{IfNotExists: true}); err == nil {
			fmt.Println("Table Created if not exists: ", tableName)
			createHistory(conn, tableName)
		} else {
			fmt.Println("Table Creation Error: ", tableName, err.Error())
		}
	}
	return
}

//Create all Tables if not exists whose Fk present in table Model
func createTableDependencies(conn *pg.DB, tableModel interface{}) {
	fields := util.GetStructField(tableModel)
	for _, curField := range fields {
		refTable := util.RefTable(curField)
		if len(refTable) > 0 {
			if refTableModel, isValid := util.TableMap[refTable]; isValid == true {
				if _, alreadyCreated := tableCreated[refTableModel]; alreadyCreated == false {
					tableCreated[refTableModel] = true
					createEnum(conn, refTable)
					createTableDependencies(conn, refTableModel)
					if err := conn.Model(tableModel).CreateTable(
						&orm.CreateTableOptions{IfNotExists: true}); err == nil {
						fmt.Println("Table Created if not exists: ", refTable)
						createHistory(conn, refTable)
					} else {
						fmt.Println("Dependency Creation Error: ", refTable, err.Error())
					}
				}
			}
		}
	}
}

//Get Struct Column Default Value from ColumnTag Detail
func getStructColDefaultValue(columnTagDetail string) (defaultValue string) {
	defalutValCheck := strings.Split(columnTagDetail, "default")
	if len(defalutValCheck) > 1 {
		defaultValueSplit := strings.Split(defalutValCheck[1], " ")
		if len(defaultValueSplit) > 1 {
			defaultValue = strings.Trim(defaultValueSplit[1], " ")
		}
	} else {
		defaultValue = "null"
	}
	return
}

//Get Struct Column Data Type form ColumnTag Detail
func getStructColDataType(columnTagDetail string) (dataType string) {
	typeValCheck := strings.Split(columnTagDetail, " ")
	if len(typeValCheck) > 0 {
		if strings.Contains(typeValCheck[0], "(") == true {
			dataType = strings.Split(typeValCheck[0], "(")[0]
		} else {
			dataType = typeValCheck[0]
		}
	}
	return
}

//Get Struct Column Is Nullable or Not form ColumnTag Detail
func getStructColNullable(columnTagDetail string) (isNullable string) {
	if strings.Contains(columnTagDetail, "not null") == true {
		isNullable = util.NO_FLAG
	} else {
		isNullable = util.YES_FLAG
	}
	return
}

//Get Struct Column Char Max Length form ColumnTag Detail
func getStructColMaxCharLen(columnTagDetail string) (charMaxLen string) {
	if strings.Contains(columnTagDetail, "varchar") {
		charCheck := strings.Split(columnTagDetail, "varchar")
		if len(charCheck) > 1 {
			dataType := charCheck[1]
			if len(dataType) > 1 && strings.Contains(dataType, "(") == true {
				dataType = dataType[1:]
				maxLenCheck := strings.Split(dataType, ")")
				if len(maxLenCheck) > 0 {
					charMaxLen = maxLenCheck[0]
				}
			}
		}
	}
	return
}

//Get FK constraint falg
func getConstraintFlag(key string) (flag string) {
	switch key {
	case "noaction":
		flag = "a"
	case "restrict":
		flag = "r"
	case "cascade":
		flag = "c"
	case "setnull":
		flag = "n"
	default:
		flag = "d"
	}
	return
}

//Get FK constraint Flag by key
func getConstraintFlagByKey(refCheck string, key string) (flag string) {
	flag = "d"
	if strings.Contains(refCheck, key) == true {
		keyCheck := strings.Split(refCheck, key)
		if len(keyCheck) > 1 {
			keyDetail := strings.Split(strings.Trim(keyCheck[1], " "), " ")
			keyDetailLen := len(keyDetail)
			if keyDetailLen > 0 {
				key = keyDetail[0]
				if (key == "set" || key == "no") && keyDetailLen > 1 {
					key += keyDetail[1]
				}
				flag = getConstraintFlag(key)
			}
		}
	}
	return
}

//Get FK table and column name
func getFkDetail(refCheck string) (table, column string) {
	refDetail := strings.Split(strings.Trim(refCheck, " "), " ")
	if len(refDetail) > 0 {
		if strings.Contains(refDetail[0], "(") == true {
			tableDetail := strings.Split(refDetail[0], "(")
			if len(tableDetail) > 1 {
				table = tableDetail[0]
				column = strings.Trim(tableDetail[1], ")")
			}
		}
	}
	return
}

//Set Struct Column Constraint
func setStructColConstraint(columnTagDetail string, constraint *model.StructColumnSchema) {
	constraint.IsDeferrable = util.NO_FLAG
	constraint.InitiallyDeferred = util.NO_FLAG
	if strings.Contains(columnTagDetail, strings.ToLower(util.PRIMARY_KEY)) == true {
		constraint.ConstraintType = util.PRIMARY_KEY
	} else if strings.Contains(columnTagDetail, util.REFERENCES) {
		constraint.ConstraintType = util.FOREIGN_KEY
		if strings.Contains(columnTagDetail, "deferrable") == true {
			constraint.IsDeferrable = util.YES_FLAG
		}
		if strings.Contains(columnTagDetail, "initially deferred") == true {
			constraint.InitiallyDeferred = util.YES_FLAG
		}
		referenceCheck := strings.Split(columnTagDetail, util.REFERENCES)
		if len(referenceCheck) > 1 {
			constraint.ForiegnTableName, constraint.ForeignColumnName =
				getFkDetail(referenceCheck[1])
			constraint.DeleteType = getConstraintFlagByKey(referenceCheck[1], "delete")
			constraint.UpdateType = getConstraintFlagByKey(referenceCheck[1], "update")
		}
	} else if strings.Contains(columnTagDetail, strings.ToLower(util.UNIQUE_KEY)) == true {
		constraint.ConstraintType = util.UNIQUE_KEY
	}
	return
}

//Get Column Schema of Structure Model of Database
func getStructColSchema(columnTagDetail string) (structColSchema model.StructColumnSchema) {
	setStructColConstraint(columnTagDetail, &structColSchema)
	structColSchema.ColumnDefault = getStructColDefaultValue(columnTagDetail)
	structColSchema.DataType = getStructColDataType(columnTagDetail)
	if structColSchema.ConstraintType == util.PRIMARY_KEY {
		structColSchema.IsNullable = util.NO_FLAG
	} else {
		structColSchema.IsNullable = getStructColNullable(columnTagDetail)
	}
	structColSchema.CharMaxLen = getStructColMaxCharLen(columnTagDetail)
	return
}

//Get Table column schema by column name
func getTableColSchema(columnSchema []model.ColumnSchema,
	columnName string) (tableSchema model.ColumnSchema) {
	for _, curColSchema := range columnSchema {
		if curColSchema.ColumnName == columnName {
			tableSchema = curColSchema
			break
		}
	}
	return
}

//Alter Add or Drop column in database
func alterColumn(tx *pg.Tx, tableName string, key string,
	columnName string, columnDetail string) (choice string, err error) {
	alterQuery := fmt.Sprintf("ALTER TABLE %v %v %v %v;\n",
		tableName, key, columnName, columnDetail)
	if util.IsSkip(tableName) == false {
		alterQuery += fmt.Sprintf("ALTER TABLE %v %v %v %v;",
			tableName+util.HISTORY_TAG, key, columnName, strings.Split(columnDetail, " ")[0])
	}
	fmt.Printf("%v\nWant to continue (y/n): ", alterQuery)
	fmt.Scan(&choice)
	if choice == util.YES_CHOICE {
		if _, err = tx.Exec(alterQuery); err == nil {
			util.QueryFp.WriteString(fmt.Sprintf("-- ALTER TABLE\n%v\n", alterQuery))
			err = createTrigger(tx, tableName)
		}
	}
	return
}

//Alter Modify column in database
func alterModifyColumn(tx *pg.Tx, tableName string, columnName, dataType string,
	userDefined bool) (choice string, err error) {
	modifyQuery := fmt.Sprintf("ALTER TABLE %v ALTER COLUMN %v TYPE %v",
		tableName, columnName, dataType)
	if userDefined {
		defaultQuery := fmt.Sprintf("ALTER TABLE %v ALTER COLUMN %v DROP DEFAULT;\n",
			tableName, columnName)
		modifyQuery = defaultQuery + modifyQuery
		modifyQuery += fmt.Sprintf(" USING %v::text::%v", columnName, dataType)
	}
	modifyQuery += ";\n"
	if util.IsSkip(tableName) == false {
		modifyQuery += fmt.Sprintf("ALTER TABLE %v ALTER COLUMN %v TYPE %v",
			tableName+util.HISTORY_TAG, columnName, dataType)
		if userDefined {
			modifyQuery += fmt.Sprintf(" USING %v::text::%v;", columnName, dataType)
		}
	}
	fmt.Printf("%v\nWant to continue (y/n): ", modifyQuery)
	fmt.Scan(&choice)
	if choice == util.YES_CHOICE {
		if _, err = tx.Exec(modifyQuery); err == nil {
			util.QueryFp.WriteString(fmt.Sprintf("-- ALTER COLUMN\n%v\n", modifyQuery))
			err = createTrigger(tx, tableName)
		}
	}
	return
}

//Compare DataType of Struct and Table column
func compareDataType(structColSchema model.StructColumnSchema,
	columnSchema model.ColumnSchema) (isEqual bool) {
	isEqual = true
	if tableDataType, exists :=
		structToTableDataMap[structColSchema.DataType]; exists == true {
		if tableDataType != columnSchema.UdtName {
			isEqual = false
		}
	} else if columnSchema.DataType == util.USER_DEFINED {
		if columnSchema.UdtName != structColSchema.DataType {
			isEqual = false
		}
	} else {
		fmt.Println("UNSUPPORTED DATATYPE", structColSchema.ColumnName,
			structColSchema.DataType)
	}
	return
}

//Compare Default Value of Struct and Table column
func compareDefaultValue(structColSchema model.StructColumnSchema,
	columnSchema model.ColumnSchema) (isEqual bool) {
	castRemoved := ""
	isEqual = true
	colDefaultVal := strings.ToLower(columnSchema.ColumnDefault)
	castRemovedSlice := strings.Split(colDefaultVal, "::")
	if len(castRemovedSlice) > 0 {
		castRemoved = castRemovedSlice[0]
	}
	// fmt.Println(structColSchema.ColumnDefault, castRemoved)
	if castRemoved == "" {
		castRemoved = "null"
	}
	// fmt.Println(structColSchema.ColumnName, structColSchema.ColumnDefault, castRemoved, columnSchema.DataType)
	if columnSchema.DataType == util.USER_DEFINED && castRemoved != "null" {
		if columnSchema.ColumnDefault != fmt.Sprintf("%v::%v",
			structColSchema.ColumnDefault, structColSchema.DataType) {
			isEqual = false
		}
	} else if structColSchema.DataType == "serial" && columnSchema.ColumnDefault != "" {
		if strings.Contains(columnSchema.ColumnDefault, fmt.Sprintf("nextval('%v_",
			structColSchema.TableName)) == false {
			// fmt.Println("default", columnSchema.ColumnDefault,
			// 	structColSchema.TableName, structColSchema.ColumnName)
			isEqual = false
		}
	} else if structColSchema.ColumnDefault != castRemoved &&
		strings.Trim(structColSchema.ColumnDefault, "'") != strings.Trim(castRemoved, "'") {
		// fmt.Println("YO", structColSchema.ColumnDefault, castRemoved)
		isEqual = false
	}
	return
}

//Alter Default value in table column
func alterDefaultValue(tx *pg.Tx, tableName string, columnName string,
	defaultValue string) (choice string, err error) {
	option := "SET"
	if defaultValue == "" {
		option = "DROP"
	}
	defaultQuery := fmt.Sprintf("ALTER TABLE %v ALTER COLUMN %v %v DEFAULT %v",
		tableName, columnName, option, defaultValue)
	fmt.Printf("%v\nWant to continue (y/n): ", defaultQuery)
	fmt.Scan(&choice)
	if choice == util.YES_CHOICE {
		if _, err = tx.Exec(defaultQuery); err == nil {
			util.QueryFp.WriteString(fmt.Sprintf("-- ALTER COLUMN\n%v\n", defaultQuery))
		}
	}
	return
}

//Alter Nullable value in table column
func alterNullableValue(tx *pg.Tx, tableName string, columnName string,
	isNullable string) (choice string, err error) {
	option := "SET"
	if isNullable == util.YES_FLAG {
		option = "DROP"
	}
	nullableQuery := fmt.Sprintf("ALTER TABLE %v ALTER COLUMN %v %v NOT NULL",
		tableName, columnName, option)
	fmt.Printf("%v\nWant to continue (y/n): ", nullableQuery)
	fmt.Scan(&choice)
	if choice == util.YES_CHOICE {
		if _, err = tx.Exec(nullableQuery); err == nil {
			util.QueryFp.WriteString(fmt.Sprintf("-- ALTER COLUMN\n%v\n", nullableQuery))
		}
	}
	return
}

//Alter Add Constraint in table column
func alterAddConstraint(tx *pg.Tx, tableName, columnName,
	constraintType, constraintDetail string) (choice string, err error) {
	tag := ""
	switch constraintType {
	case util.PRIMARY_KEY:
		tag = "pkey"
	case util.UNIQUE_KEY:
		tag = "key"
	case util.FOREIGN_KEY:
		tag = "fkey"
	}
	addConstraintQuery := fmt.Sprintf("ALTER TABLE %v ADD CONSTRAINT %v_%v_%v %v %v",
		tableName, tableName, columnName, tag, constraintType, constraintDetail)
	fmt.Printf("%v\nWant to continue (y/n): ", addConstraintQuery)
	fmt.Scan(&choice)
	if choice == util.YES_CHOICE {
		if _, err = tx.Exec(addConstraintQuery); err == nil {
			util.QueryFp.WriteString(fmt.Sprintf("-- ALTER CONSTRAINT\n%v\n", addConstraintQuery))
		}
	}
	return
}

//Compare Constraint of struct and table
func compareConstraint(tx *pg.Tx, structColSchema model.StructColumnSchema,
	columnSchema model.ColumnSchema) (err error) {
	if structColSchema.ConstraintType != columnSchema.ConstraintType ||
		(structColSchema.ConstraintType == util.FOREIGN_KEY &&
			((structColSchema.UpdateType != columnSchema.UpdateType) ||
				(structColSchema.DeleteType != columnSchema.DeleteType))) {
		var choice string
		if choice, err = dropConstraint(tx, structColSchema.TableName,
			columnSchema.ConstraintName); err == nil && choice == util.YES_CHOICE {
			if columnSchema.ConstraintName != "" {
				alteredColumn[structColSchema.ColumnName] +=
					fmt.Sprintf("CONSTRAINT DROP:\t%v\n", columnSchema.ConstraintName)
			}
		}
		if structColSchema.ConstraintType != "" {
			constraintDetail := getStructConstraintDetail(structColSchema)
			if choice, err = alterAddConstraint(tx, structColSchema.TableName, structColSchema.ColumnName,
				structColSchema.ConstraintType, constraintDetail); err == nil && choice == util.YES_CHOICE {
				alteredColumn[structColSchema.ColumnName] +=
					fmt.Sprintf("CONSTRAINT ADDED:\t%v %v\n", structColSchema.ConstraintType, constraintDetail)
			}
		}
	}
	return
}

//Compare Schema of struct and table
func compareSchema(tx *pg.Tx, structColSchema model.StructColumnSchema,
	columnSchema model.ColumnSchema) (err error) {
	if isEqual := compareDataType(structColSchema, columnSchema); isEqual == false ||
		structColSchema.CharMaxLen != columnSchema.CharMaxLen {
		var choice string
		modifyDataType := structColSchema.DataType
		if structColSchema.CharMaxLen != "" {
			modifyDataType = modifyDataType + "(" + structColSchema.CharMaxLen + ")"
		}
		userDefined := false
		if isEqual || columnSchema.DataType == util.USER_DEFINED {
			userDefined = true
		}
		if choice, err = alterModifyColumn(tx, structColSchema.TableName, structColSchema.ColumnName,
			modifyDataType, userDefined); err != nil {
			return
		} else if choice == util.YES_CHOICE {
			alteredColumn[structColSchema.ColumnName] +=
				fmt.Sprintf("PREVIOUS TYPE:\t%v %v\n", columnSchema.DataType, columnSchema.CharMaxLen)
		}
	}
	if isEqual := compareDefaultValue(structColSchema, columnSchema); isEqual == false {
		var choice string
		if choice, err = alterDefaultValue(tx, structColSchema.TableName,
			structColSchema.ColumnName, structColSchema.ColumnDefault); err != nil {
			return
		} else if choice == util.YES_CHOICE {
			alteredColumn[structColSchema.ColumnName] +=
				fmt.Sprintf("PREVIOUS DEFAULT VALUE:\t%v\n", columnSchema.ColumnDefault)
		}
	}
	if structColSchema.IsNullable != columnSchema.IsNullable {
		var choice string
		if choice, err = alterNullableValue(tx, structColSchema.TableName,
			structColSchema.ColumnName, structColSchema.IsNullable); err != nil {
			return
		} else if choice == util.YES_CHOICE {
			alteredColumn[structColSchema.ColumnName] +=
				fmt.Sprintf("PREVIOUS NULLABLE VALUE:\t%v\n", columnSchema.IsNullable)
		}
	}
	err = compareConstraint(tx, structColSchema, columnSchema)
	return
}

//Get Constraint Details from struct schema model
func getStructConstraintDetail(structColSchema model.StructColumnSchema) (constraintDetail string) {
	constraintDetail = fmt.Sprintf("(%v)", structColSchema.ColumnName)
	if structColSchema.ConstraintType == util.FOREIGN_KEY {
		deleteTag := getConstraintTagByFlag(structColSchema.DeleteType)
		updateTag := getConstraintTagByFlag(structColSchema.UpdateType)
		constraintDetail += fmt.Sprintf(" REFERENCES %v(%v) ON DELETE %v ON UPDATE %v",
			structColSchema.ForiegnTableName, structColSchema.ForeignColumnName,
			deleteTag, updateTag)
	}
	return
}

//Get Constraint Tag by Flag
func getConstraintTagByFlag(flag string) (tag string) {
	switch flag {
	case "a":
		tag = "NO ACTION"
	case "r":
		tag = "RESTRICT"
	case "c":
		tag = "CASCADE"
	case "n":
		tag = "SET NULL"
	default:
		tag = "SET DEFAULT"
	}
	return
}

//Check and alter column based on the struct change
func checkAndAlterColumn(tx *pg.Tx, structColSchema model.StructColumnSchema,
	columnSchema model.ColumnSchema, columnDetail string) (err error) {

	// fmt.Println("Struct", structColSchema)
	// fmt.Println("Table", columnSchema)
	if ally.IsEmptyStruct(columnSchema) == true {
		var choice string
		// fmt.Println("Adding New Column", structColSchema.ColumnName,
		// 	"in Table:", structColSchema.TableName)
		if choice, err = alterColumn(tx, structColSchema.TableName, "ADD",
			structColSchema.ColumnName, columnDetail); err == nil && choice == util.YES_CHOICE {
			addedColumn[structColSchema.ColumnName] = columnDetail
		}
	} else {
		err = compareSchema(tx, structColSchema, columnSchema)
	}
	return
}

//Get Column Deteail By Schema of Column
func getColumnDetailBySchema(columnSchema model.ColumnSchema) (columnDetail string) {
	if columnSchema.DataType == util.USER_DEFINED {
		columnDetail = columnSchema.UdtName
	} else {
		columnDetail = columnSchema.DataType
	}
	if columnSchema.CharMaxLen != "" {
		columnDetail += "(" + columnSchema.CharMaxLen + ")"
	}
	if columnSchema.IsNullable == util.NO_FLAG {
		columnDetail += " NOT NULL"
	}
	if columnSchema.ColumnDefault != "" {
		columnDetail += " DEFAULT " + columnSchema.ColumnDefault
	}
	if columnSchema.ForeignTableName != "" {
		columnDetail = fmt.Sprintf("%v REFERENCES %v(%v)",
			columnDetail, columnSchema.ForeignTableName, columnSchema.ForeignColumnName)
	}
	if columnSchema.ConstraintType != "" {
		columnDetail += " " + columnSchema.ConstraintType
	}
	if columnSchema.IsDeferrable == util.YES_FLAG {
		columnDetail += " DEFFERABLE"
	}
	if columnSchema.InitiallyDeferred == util.YES_FLAG {
		columnDetail += " INITIALLY DEFERRED"
	}
	return
}

//Drop Constraint By Constraint Name
func dropConstraint(tx *pg.Tx, tableName string, constraintName string) (choice string, err error) {
	if constraintName == "" {
		choice = util.YES_CHOICE
		return
	}
	dropConstraintQuery := fmt.Sprintf("ALTER TABLE %v DROP CONSTRAINT %v",
		tableName, constraintName)
	fmt.Printf("%v\nWant to continue (y/n): ", dropConstraintQuery)
	fmt.Scan(&choice)
	if choice == util.YES_CHOICE {
		if _, err = tx.Exec(dropConstraintQuery); err == nil {
			util.QueryFp.WriteString(fmt.Sprintf("-- ALTER CONSTRAINT\n%v\n", dropConstraintQuery))
			alteredConstraint[tableName] = dropConstraintQuery
		}
	}
	return
}

//Check and Remove Column and Constraint
func checkAndRemoveColumn(tx *pg.Tx, tableSchema map[string]model.ColumnSchema,
	tableName string, visitedCol map[string]string) (err error) {
	for _, curColumnSchema := range tableSchema {
		if _, exists := visitedCol[curColumnSchema.ColumnName]; exists == false {
			var choice string
			// fmt.Println("Removing Column", curColumnSchema.ColumnName, "in Table:", tableName)
			columnDetail := getColumnDetailBySchema(curColumnSchema)
			if choice, err = dropConstraint(tx, tableName,
				curColumnSchema.ConstraintName); err == nil && choice == util.YES_CHOICE {
				// fmt.Println("Dropped Constraint", curColumnSchema.ConstraintName, "from Table:", tableName)
				if choice, err = alterColumn(tx, tableName, "DROP",
					curColumnSchema.ColumnName, ""); err == nil && choice == util.YES_CHOICE {
					removedColumn[curColumnSchema.ColumnName] = columnDetail
				}
			}
		}
	}
	return
}

//Check Struct Table Model with Database Schema
func checkTableToAlter(tx *pg.Tx, tableSchema map[string]model.ColumnSchema,
	tableModel interface{}, tableName string) (err error) {

	addedColumn = make(map[string]string)
	removedColumn = make(map[string]string)
	alteredColumn = make(map[string]string)
	reflectObj := reflect.ValueOf(tableModel)
	if reflectObj.Kind() == reflect.Ptr {
		reflectObj = reflectObj.Elem()
	}
	visitedStructColumn := make(map[string]string)

	fields := util.GetStructField(tableModel)
	if reflectObj.Kind() == reflect.Struct {
		for _, refType := range fields {

			if sqlTag, exists := refType.Tag.Lookup("pg"); exists == true {
				columnDetail := strings.Split(sqlTag, ",")
				if len(columnDetail) > 1 {
					columnName := columnDetail[0]
					columnTypeDetail := strings.Split(columnDetail[1], ":")
					if len(columnTypeDetail) > 1 {
						visitedStructColumn[columnName] = columnDetail[1]
						structColSchema :=
							getStructColSchema(strings.ToLower(columnTypeDetail[1]))
						structColSchema.TableName = tableName
						structColSchema.ColumnName = columnName
						if err = checkAndAlterColumn(tx, structColSchema,
							tableSchema[columnName], columnTypeDetail[1]); err != nil {
							tx.Rollback()
							fmt.Println("Alter Error: ", tableName, err.Error())
							return
						}
					} else {
						fmt.Println("SQL datatype not defined of: ", tableName, refType.Name)
					}
				} else {
					fmt.Println("SQL Column Name not defined of: ", tableName, refType.Name)
				}
			}
		}
		if err = checkAndRemoveColumn(tx, tableSchema,
			tableName, visitedStructColumn); err != nil {
			tx.Rollback()
			fmt.Println("Alter Remove Error: ", tableName, err.Error())
		}
	} else {
		fmt.Println("TableModel is not Structure ", tableModel)
	}
	return
}

//Log Alter Details table wise
func logAlterDetails(tableName string) {
	inFlag := false
	for columnName, columnDetail := range addedColumn {
		if inFlag == false {
			log.Println(fmt.Sprintf("----ALTER TABLE: %v", tableName))
			inFlag = true
		}
		log.Println(fmt.Sprintf("\nCOLUMN ADDED:\t%v\tTYPE:\t%v",
			columnName, columnDetail))
	}
	for columnName, columnDetail := range removedColumn {
		if inFlag == false {
			log.Println(fmt.Sprintf("----ALTER TABLE: %v", tableName))
			inFlag = true
		}
		log.Println(fmt.Sprintf("\nCOLUMN REMOVED:\t%v\tTYPE:\t%v",
			columnName, columnDetail))
	}
	for columnName, columnDetail := range alteredColumn {
		if inFlag == false {
			log.Println(fmt.Sprintf("----ALTER TABLE: %v", tableName))
			inFlag = true
		}
		log.Println(fmt.Sprintf("\nCOLUMN MODIFIED:\t%v\n%v",
			columnName, columnDetail))
	}
	if sql, exists := alteredConstraint[tableName]; exists {
		if inFlag == false {
			log.Println(fmt.Sprintf("----ALTER TABLE: %v", tableName))
			inFlag = true
		}
		log.Println(fmt.Sprintf("\nCONSTRAINT EXECUTED:\t%v\n", sql))
	}
}

//Check unique key constraint to alter
func checkUniqueKeyToAlter(tx *pg.Tx, uniqueKeySchema []model.UniqueKeySchema, tableName string) (err error) {

	uk := util.GetUniqueKey(tableName)
	for _, curUK := range uniqueKeySchema {
		if _, exists := uk[curUK.ConstraintName]; exists {
			//TODO: check unique key diff
			delete(uk, curUK.ConstraintName)
		} else {
			if _, err = dropConstraint(tx, tableName, curUK.ConstraintName); err != nil {
				fmt.Println("Error in Unique Key drop", err.Error())
				break
			}
		}
	}
	uniqueKeySQL := ""
	for ukName, ukFields := range uk {
		uniqueKeySQL += getUniqueKeyQuery(tableName, ukName, ukFields)
	}
	if uniqueKeySQL != "" {
		choice := ""
		fmt.Printf("%v\nWant to continue (y/n): ", uniqueKeySQL)
		fmt.Scan(&choice)
		if choice == util.YES_CHOICE {
			if _, err = tx.Exec(uniqueKeySQL); err != nil {
				fmt.Println("Unique Key Constraint creation error: ", tableName, err.Error())
			} else {
				// util.QueryFp.WriteString(fmt.Sprintf("-- ALTER CONSTRAINT\n%v\n", uniqueKeySQL))
				alteredConstraint[tableName] = uniqueKeySQL
			}
		}
	}
	return
}

//Alter Table
func alterTable(conn *pg.DB, tableName string) (err error) {
	initStructTableMap()
	var (
		columnSchema    []model.ColumnSchema
		constraint      []model.ColumnSchema
		uniqueKeySchema []model.UniqueKeySchema
		tx              *pg.Tx
	)
	tableModel, isValid := util.TableMap[tableName]
	if isValid == true {
		if columnSchema, err = util.GetColumnSchema(conn, tableName); err == nil {
			if constraint, err = util.GetConstraint(conn, tableName); err == nil {
				tableSchema := util.MergeColumnConstraint(columnSchema, constraint)

				if tx, err = conn.Begin(); err == nil {
					if err = checkTableToAlter(tx, tableSchema, tableModel, tableName); err == nil {
						if uniqueKeySchema, err = util.GetCompositeUniqueKey(conn, tableName); err == nil {
							if ally.IsEmptyInterface(uniqueKeySchema) == false {
								err = checkUniqueKeyToAlter(tx, uniqueKeySchema, tableName)
							}
						} else {
							fmt.Println("Composite unique key Fetch Error: ", tableName, err.Error())
						}
					}
					if err == nil {
						tx.Commit()
					} else {
						tx.Rollback()
					}
				} else {
					fmt.Println("Transation Error: ", tableName, err.Error())
				}
			} else {
				fmt.Println("Constraint Fetch Error: ", tableName, err.Error())
			}
		} else {
			fmt.Println("Column Schema Fetch Error: ", tableName, err.Error())
		}
	} else {
		fmt.Println("Invalid Table Name: ", tableName)
	}
	return
}
