package postgresql

import (
	"github.com/go-pg/pg/v10"
)

//DB is a wrap struct of go-pg
type DB struct {
	*pg.DB
	query          string
	splitSearchstr string
	value          []interface{}
	bindValue      []interface{}
	condition      map[string][]interface{}
	orCondition    map[string][]interface{}
	having         map[string][]interface{}
	group          []string
	order          []string
	limit          string
	offset         string
}

//TX is a wrap struct of go-pg tx
type TX struct {
	*pg.Tx
	query          string
	splitSearchstr string
	value          []interface{}
	bindValue      []interface{}
	condition      map[string][]interface{}
	orCondition    map[string][]interface{}
	having         map[string][]interface{}
	group          []string
	order          []string
	limit          string
	offset         string
}

type Hook struct {
	Id    int
	Value string
}
