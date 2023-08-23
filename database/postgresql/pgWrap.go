package postgresql

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"gitlab.com/g-harshit/plib/ally"
	"gitlab.com/g-harshit/plib/constant"
	"gitlab.com/g-harshit/plib/perror"
)

//WrapPGConn : wrap pg conn
func WrapPGConn(master bool) (conn *DB, err error) {
	var dbConn *pg.DB
	if dbConn, err = Conn(master); err == nil {
		conn = &DB{
			DB:             dbConn,
			query:          "",
			splitSearchstr: "",
			value:          []interface{}{},
			bindValue:      []interface{}{},
			condition:      map[string][]interface{}{},
			orCondition:    map[string][]interface{}{},
			having:         map[string][]interface{}{},
			group:          []string{},
			order:          []string{},
			limit:          "",
			offset:         "",
		}
	}
	return
}

//WrapGroup : Group by added in query of Wrap Connection
func (db *DB) WrapGroup(groupValue ...string) *DB {
	db.group = append(db.group, groupValue...)
	return db
}

//Group : Group by added in query of Wrap Connection
func (db *DB) Group(groupValue ...string) *DB {
	db.group = append(db.group, groupValue...)
	return db
}

//WrapOrder : Order by added in query of Wrap Connection
func (db *DB) WrapOrder(orderVal ...string) *DB {
	db.order = append(db.order, orderVal...)
	return db
}

//Order : Order by added in query of Wrap Connection
func (db *DB) Order(orderVal ...string) *DB {
	db.order = append(db.order, orderVal...)
	return db
}

//WrapLimit : Limit added in query of Wrap Connection
func (db *DB) WrapLimit(limitVal int) *DB {
	db.limit = strconv.Itoa(limitVal)
	return db
}

//Limit : Limit added in query of Wrap Connection
func (db *DB) Limit(limitVal int) *DB {
	db.limit = strconv.Itoa(limitVal)
	return db
}

//WrapOffset : Offset added in query of Wrap Connection
func (db *DB) WrapOffset(offsetVal int) *DB {
	db.offset = strconv.Itoa(offsetVal)
	return db
}

//Offset : Offset added in query of Wrap Connection
func (db *DB) Offset(offsetVal int) *DB {
	db.offset = strconv.Itoa(offsetVal)
	return db
}

//WrapWhere : Where condition in query of Wrap Connection
func (db *DB) WrapWhere(cond string, value ...interface{}) *DB {
	db.condition[cond] = append(db.condition[cond], value...)
	return db
}

//Where : Where condition in query of Wrap Connection
func (db *DB) Where(cond string, value ...interface{}) *DB {
	db.condition[cond] = append(db.condition[cond], value...)
	return db
}

//WrapWhereOr : Where or condition in query of Wrap Connection
func (db *DB) WrapWhereOr(cond string, value ...interface{}) *DB {
	db.orCondition[cond] = append(db.orCondition[cond], value...)
	return db
}

//WhereOr : Where or condition in query of Wrap Connection
func (db *DB) WhereOr(cond string, value ...interface{}) *DB {
	db.orCondition[cond] = append(db.orCondition[cond], value...)
	return db
}

//WrapHaving : Having condition in query of Wrap Connection
func (db *DB) WrapHaving(cond string, value ...interface{}) *DB {
	db.having[cond] = append(db.having[cond], value...)
	return db
}

//Having : Having condition in query of Wrap Connection
func (db *DB) Having(cond string, value ...interface{}) *DB {
	db.having[cond] = append(db.having[cond], value...)
	return db
}

//buildQuery : build the final query from execution
func (db *DB) buildQuery() (query string) {
	query = db.query
	db.bindValue = []interface{}{}
	if len(db.value) > 0 {
		db.bindValue = append(db.bindValue, db.value...)
	}
	if len(db.condition) > 0 {
		query += ` 
		WHERE `
		for cond, condVal := range db.condition {
			query += cond + ` 
		AND `
			db.bindValue = append(db.bindValue, condVal...)
		}
		query = strings.TrimSuffix(query, `
		AND `)
	}
	if len(db.orCondition) > 0 {
		if len(db.condition) == 0 {
			query += `
		WHERE ( `
		} else {
			query += `
		AND ( `
		}
		for cond, condVal := range db.orCondition {
			query += cond + ` 
		OR `
			db.bindValue = append(db.bindValue, condVal...)
		}
		query = strings.TrimSuffix(query, `
		OR `)
		query += `)`
	}
	if len(db.group) > 0 {
		query += ` 
		GROUP BY ` + strings.Join(db.group, ",")
	}
	if len(db.having) > 0 {
		query += ` 
		HAVING `
		for cond, condVal := range db.having {
			query += cond + ` 
		, `
			db.bindValue = append(db.bindValue, condVal...)
		}
		query = strings.TrimSuffix(query, `
		, `)
	}
	if len(db.order) > 0 {
		query += ` 
		ORDER BY ` + strings.Join(db.order, ",")
	}
	if db.limit != "" {
		query += ` 
		LIMIT ` + db.limit
	}
	if db.offset != "" {
		query += " OFFSET " + db.offset
	}

	db.flush()
	return
}

func (db *DB) flush() {
	db.query = ""
	db.splitSearchstr = ""
	db.value = []interface{}{}
	db.condition = map[string][]interface{}{}
	db.orCondition = map[string][]interface{}{}
	db.having = map[string][]interface{}{}
	db.group = []string{}
	db.order = []string{}
	db.limit = ""
	db.offset = ""
}

//RawQuery execute the raw query
func (db *DB) RawQuery(dbModel interface{}, selectSQL string, value ...interface{}) (orm.Result, error) {
	db.query = selectSQL
	db.value = value
	res, err := db.Query(dbModel, db.buildQuery(), db.bindValue...)
	if err != nil {
		err = perror.SelectError(err)
	}
	return res, err
}

//AdvFilter : Select query with AdvFilter
func (db *DB) AdvFilter(param ally.AdvFilter) *DB {
	for colName, colValue := range param.Filter {
		// //Unescape the query value
		// if colValue.Unescape {
		// 	colValue.Value, _ = url.QueryUnescape(colValue.Value)
		// }

		if colValue.Skip {
			continue
		}

		var field = param.Alias
		if colValue.Alias != "" {
			field = colValue.Alias
		}
		if field == "" {
			field = colName
		} else {
			field = field + "." + colName
		}
		if len(colValue.Exclude) > 0 {
			field = colNameFilter(colValue.Exclude, field)
		}

		if colValue.Sort {
			field = strings.TrimSuffix(field, ".sort")
			db = db.WrapOrder(field + " " + strings.ToUpper(colValue.Value))
			continue
		}

		if operator, fieldVal := getOptVal(colValue.Operator, colValue.Value, colValue.Exclude...); operator != "" {
			cond := ""
			if colValue.Operator == constant.LikeOrOP {
				operator = strings.Replace(operator, "$ColumnName$", field, -1)
				cond = field + operator
			} else if colValue.Operator == constant.LikeAllSearchOP {
				var newFieldVal []interface{}
				for _, curfield := range strings.Split(field, ",") {
					newFieldVal = append(newFieldVal, fieldVal...)

					cond += curfield + strings.Replace(operator, "$ColumnName$", curfield, -1) + " OR "
				}
				fieldVal = newFieldVal
				cond = strings.TrimSuffix(cond, " OR ")
				cond = "(" + cond + ")"
			} else {
				cond = field + operator
			}
			if colValue.Or {
				db.orCondition[cond] = append(db.orCondition[cond], fieldVal...)
			} else {
				db.condition[cond] = append(db.condition[cond], fieldVal...)
			}
		}
	}
	db.splitSearchstr = param.SplitSearchStr
	if param.Limit != 0 {
		db = db.Limit(param.Limit).Offset(param.Offset)
	}
	return db
}

//getOptVal : get the value from operator
func getOptVal(operator string, value string, exclude ...string) (cond string, fieldVal []interface{}) {
	switch strings.ToLower(operator) {
	case constant.LikeOP:
		cond += " ILIKE ? "
		fieldVal = append(fieldVal, "%"+value+"%")
	case constant.NotEqualOP:
		cond += " <> ? "
		fieldVal = append(fieldVal, value)
	case constant.IsNull:
		cond += " IS NULL "
	case constant.IsNotNull:
		cond += " IS NOT NULL "
	case constant.InOP:
		cond += " IN ('" + strings.Replace(value, ",", "','", -1) + "') "
	case constant.NotInOP:
		cond += " NOT IN ('" + strings.Replace(value, ",", "','", -1) + "') "
	case constant.RangeOP:
		rangeVal := strings.Split(value, ",")
		if len(rangeVal) == 2 {
			cond += " BETWEEN ? AND ? "
			fieldVal = append(fieldVal, rangeVal[0], rangeVal[1])
		}
	case constant.LessThan:
		cond += " < ? "
		fieldVal = append(fieldVal, value)
	case constant.LessThanEq:
		cond += " <= ? "
		fieldVal = append(fieldVal, value)
	case constant.GreaterThan:
		cond += " > ? "
		fieldVal = append(fieldVal, value)
	case constant.GreaterThanEq:
		cond += " >= ? "
		fieldVal = append(fieldVal, value)
	case constant.LikeOrOP:
		var re = regexp.MustCompile(`\*|\+|\?|\$|\^|\.|\[|\]|\(|\)|\%|\|`)
		allEscpChars := re.FindAllSubmatch([]byte(value), -1)
		uniqEscpChars := make(map[string]struct{})
		for _, curChar := range allEscpChars {
			if _, exists := uniqEscpChars[string(curChar[0])]; !exists {
				escapeChar := []byte("\\" + string(curChar[0]))
				value = string(bytes.Replace([]byte(value), curChar[0], escapeChar, -1))
				uniqEscpChars[string(curChar[0])] = struct{}{}
			}
		}
		for _, curVal := range strings.Split(value, " ") {
			if curVal != "" {
				cond += " ILIKE ? AND $ColumnName$ "
				if len(exclude) > 0 {
					curVal = colValueFilter(exclude, curVal)
				}
				fieldVal = append(fieldVal, "%"+curVal+"%")
			}
		}
		cond = strings.TrimSuffix(cond, " AND $ColumnName$ ")
	case constant.LikeAllSearchOP:
		values := strings.Split(value, ",")
		for _, curVal := range values {
			if curVal != "" {
				cond += " ILIKE ? OR $ColumnName$ "
				if len(exclude) > 0 {
					curVal = colValueFilter(exclude, curVal)
				}
				fieldVal = append(fieldVal, "%"+curVal+"%")
			}
		}
		cond = strings.TrimSuffix(cond, " OR $ColumnName$ ")

	default:
		cond += " = ? "
		fieldVal = append(fieldVal, value)
	}
	return
}

//colNameFilter : Add filter on column name to ignore excluded character
func colNameFilter(exclude []string, qryColName string) (colFilterCond string) {
	colFilterCond = "REGEXP_REPLACE(" + qryColName + ",'(["
	for _, v := range exclude {
		colFilterCond += v
	}
	colFilterCond += "])','','g')"

	return
}

//colValueFilter : Will filter column value to ignore excluded character
func colValueFilter(exclude []string, qryColValue string) (colFilterValue string) {
	var regex string
	for _, v := range exclude {
		regex += `\` + v + `|`
	}
	regex += `\\`
	reg := regexp.MustCompile(regex)
	colFilterValue = reg.ReplaceAllLiteralString(qryColValue, "")

	return
}

//RankOrder : return rank ordering string with split search value
func (db *DB) RankOrder(searchFields []string, position int) *DB {

	oLen := len(db.order)
	if db.splitSearchstr != "" && position < oLen && position > -1 {
		fields := strings.Join(searchFields, ",")

		strQuery := ` TS_RANK_CD (
		TO_TSVECTOR ('english', CONCAT_WS(' ', ` + fields + `)) , 
		PLAINTO_TSQUERY ('english', '` + db.splitSearchstr + `')
	) DESC `

		var orderFields []string
		//appending the rank query in the specified position
		orderFields = append(orderFields, db.order[:position]...)
		orderFields = append(orderFields, strQuery)
		orderFields = append(orderFields, db.order[position:]...)
		db.order = orderFields
	}
	return db
}

//ILikeOrder : return ilike order with split search value
func (db *DB) ILikeOrder(searchFields []string, position int) *DB {

	oLen := len(db.order)
	if db.splitSearchstr != "" && len(searchFields) > 0 &&
		position < oLen && position > -1 {
		val := strings.Trim(db.splitSearchstr, " ")
		val = strings.Replace(val, "'", "''", -1)

		likeOrder := " ("
		for _, curName := range searchFields {
			likeOrder += fmt.Sprintf(" %v ILIKE '%%%v%%' OR ", curName, val)
		}
		likeOrder = strings.TrimRight(likeOrder, "OR ")
		likeOrder += `
	) DESC `

		var orderFields []string
		//appending the rank query in the specified position
		orderFields = append(orderFields, db.order[:position]...)
		orderFields = append(orderFields, likeOrder)
		orderFields = append(orderFields, db.order[position:]...)
		db.order = orderFields
	}
	return db
}
