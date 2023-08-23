package postgresql

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/TeamTutx/plib/ally"
	"github.com/TeamTutx/plib/constant"
	"github.com/TeamTutx/plib/perror"
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)

//WrapPGTx : wrap pg tx
func WrapPGTx() (txConn *TX, err error) {
	var tx *pg.Tx
	if tx, err = Tx(); err == nil {
		txConn = &TX{
			Tx:             tx,
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

//InitTx : Initialize transaction paramters
func InitTx() *TX {
	return &TX{
		Tx:             new(pg.Tx),
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

//WrapGroup : Group by added in query of Wrap Connection
func (tx *TX) WrapGroup(groupValue ...string) *TX {
	tx.group = append(tx.group, groupValue...)
	return tx
}

//Group : Group by added in query of Wrap Connection
func (tx *TX) Group(groupValue ...string) *TX {
	tx.group = append(tx.group, groupValue...)
	return tx
}

//WrapOrder : Order by added in query of Wrap Connection
func (tx *TX) WrapOrder(orderVal ...string) *TX {
	tx.order = append(tx.order, orderVal...)
	return tx
}

//Order : Order by added in query of Wrap Connection
func (tx *TX) Order(orderVal ...string) *TX {
	tx.order = append(tx.order, orderVal...)
	return tx
}

//WrapLimit : Limit added in query of Wrap Connection
func (tx *TX) WrapLimit(limitVal int) *TX {
	tx.limit = strconv.Itoa(limitVal)
	return tx
}

//Limit : Limit added in query of Wrap Connection
func (tx *TX) Limit(limitVal int) *TX {
	tx.limit = strconv.Itoa(limitVal)
	return tx
}

//WrapOffset : Offset added in query of Wrap Connection
func (tx *TX) WrapOffset(offsetVal int) *TX {
	tx.offset = strconv.Itoa(offsetVal)
	return tx
}

//Offset : Offset added in query of Wrap Connection
func (tx *TX) Offset(offsetVal int) *TX {
	tx.offset = strconv.Itoa(offsetVal)
	return tx
}

//WrapWhere : Where condition in query of Wrap Connection
func (tx *TX) WrapWhere(cond string, value ...interface{}) *TX {
	tx.condition[cond] = append(tx.condition[cond], value...)
	return tx
}

//Where : Where condition in query of Wrap Connection
func (tx *TX) Where(cond string, value ...interface{}) *TX {
	tx.condition[cond] = append(tx.condition[cond], value...)
	return tx
}

//WrapWhereOr : Where or condition in query of Wrap Connection
func (tx *TX) WrapWhereOr(cond string, value ...interface{}) *TX {
	tx.orCondition[cond] = append(tx.orCondition[cond], value...)
	return tx
}

//WhereOr : Where or condition in query of Wrap Connection
func (tx *TX) WhereOr(cond string, value ...interface{}) *TX {
	tx.orCondition[cond] = append(tx.orCondition[cond], value...)
	return tx
}

//WrapHaving : Having condition in query of Wrap Connection
func (tx *TX) WrapHaving(cond string, value ...interface{}) *TX {
	tx.having[cond] = append(tx.having[cond], value...)
	return tx
}

//Having : Having condition in query of Wrap Connection
func (tx *TX) Having(cond string, value ...interface{}) *TX {
	tx.having[cond] = append(tx.having[cond], value...)
	return tx
}

//buildQuery : build the final query from execution
func (tx *TX) buildQuery() (query string) {
	query = tx.query
	tx.bindValue = []interface{}{}
	if len(tx.value) > 0 {
		tx.bindValue = append(tx.bindValue, tx.value...)
	}
	if len(tx.condition) > 0 {
		query += ` 
		WHERE `
		for cond, condVal := range tx.condition {
			query += cond + ` 
		AND `
			tx.bindValue = append(tx.bindValue, condVal...)
		}
		query = strings.TrimSuffix(query, `
		AND `)
	}
	if len(tx.orCondition) > 0 {
		if len(tx.condition) == 0 {
			query += `
		WHERE ( `
		} else {
			query += `
		AND ( `
		}
		for cond, condVal := range tx.orCondition {
			query += cond + ` 
		OR `
			tx.bindValue = append(tx.bindValue, condVal...)
		}
		query = strings.TrimSuffix(query, `
		OR `)
		query += `)`
	}
	if len(tx.group) > 0 {
		query += ` 
		GROUP BY ` + strings.Join(tx.group, ",")
	}
	if len(tx.having) > 0 {
		query += ` 
		HAVING `
		for cond, condVal := range tx.having {
			query += cond + ` 
		, `
			tx.bindValue = append(tx.bindValue, condVal...)
		}
		query = strings.TrimSuffix(query, `
		, `)
	}
	if len(tx.order) > 0 {
		query += ` 
		ORDER BY ` + strings.Join(tx.order, ",")
	}
	if tx.limit != "" {
		query += ` 
		LIMIT ` + tx.limit
	}
	if tx.offset != "" {
		query += " OFFSET " + tx.offset
	}

	tx.flush()
	return
}

func (tx *TX) flush() {
	tx.query = ""
	tx.splitSearchstr = ""
	tx.value = []interface{}{}
	tx.condition = map[string][]interface{}{}
	tx.orCondition = map[string][]interface{}{}
	tx.having = map[string][]interface{}{}
	tx.group = []string{}
	tx.order = []string{}
	tx.limit = ""
	tx.offset = ""
}

//RawQuery execute the raw query
func (tx *TX) RawQuery(txModel interface{}, selectSQL string, value ...interface{}) (orm.Result, error) {
	tx.query = selectSQL
	tx.value = value
	res, err := tx.Query(txModel, tx.buildQuery(), tx.bindValue...)
	if err != nil {
		err = perror.SelectError(err)
	}
	return res, err
}

//AdvFilter : Select query with AdvFilter
func (tx *TX) AdvFilter(param ally.AdvFilter) *TX {
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
			tx = tx.WrapOrder(field + " " + strings.ToUpper(colValue.Value))
			continue
		}

		if operator, fieldVal := getOptVal(colValue.Operator, colValue.Value, colValue.Exclude...); operator != "" {
			if colValue.Operator == constant.LikeOrOP {
				operator = strings.Replace(operator, "$ColumnName$", field, -1)
			}
			cond := field + operator
			if colValue.Or {
				tx.orCondition[cond] = append(tx.orCondition[cond], fieldVal...)
			} else {
				tx.condition[cond] = append(tx.condition[cond], fieldVal...)
			}
		}
	}
	tx.splitSearchstr = param.SplitSearchStr
	if param.Limit != 0 {
		tx = tx.Limit(param.Limit).Offset(param.Offset)
	}
	return tx
}

//RankOrder : return rank ordering string with split search value
func (tx *TX) RankOrder(searchFields []string, position int) *TX {

	oLen := len(tx.order)
	if tx.splitSearchstr != "" && position < oLen && position > -1 {
		fields := strings.Join(searchFields, ",")

		strQuery := ` TS_RANK_CD (
		TO_TSVECTOR ('english', CONCAT_WS(' ', ` + fields + `)) , 
		PLAINTO_TSQUERY ('english', '` + tx.splitSearchstr + `')
	) DESC `

		var orderFields []string
		//appending the rank query in the specified position
		orderFields = append(orderFields, tx.order[:position]...)
		orderFields = append(orderFields, strQuery)
		orderFields = append(orderFields, tx.order[position:]...)
		tx.order = orderFields
	}
	return tx
}

//ILikeOrder : return ilike order with split search value
func (tx *TX) ILikeOrder(searchFields []string, position int) *TX {

	oLen := len(tx.order)
	if tx.splitSearchstr != "" && len(searchFields) > 0 &&
		position < oLen && position > -1 {
		val := strings.Trim(tx.splitSearchstr, " ")

		likeOrder := " ("
		for _, curName := range searchFields {
			likeOrder += fmt.Sprintf(" %v ILIKE '%%%v%%' OR ", curName, val)
		}
		likeOrder = strings.TrimRight(likeOrder, "OR ")
		likeOrder += `
	) DESC `

		var orderFields []string
		//appending the rank query in the specified position
		orderFields = append(orderFields, tx.order[:position]...)
		orderFields = append(orderFields, likeOrder)
		orderFields = append(orderFields, tx.order[position:]...)
		tx.order = orderFields
	}
	return tx
}
