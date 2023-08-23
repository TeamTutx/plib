package model

//ColumnSchema : Table Column Schema Model
type ColumnSchema struct {
	ColumnName        string `pg:"column_name"`
	ColumnDefault     string `pg:"column_default"`
	DataType          string `pg:"data_type"`
	UdtName           string `pg:"udt_name"`
	IsNullable        string `pg:"is_nullable"`
	CharMaxLen        string `pg:"character_maximum_length"`
	ConstraintType    string `pg:"constraint_type"`
	ConstraintName    string `pg:"constraint_name"`
	IsDeferrable      string `pg:"is_deferrable"`
	InitiallyDeferred string `pg:"initially_deferred"`
	ForeignTableName  string `pg:"foreign_table_name"`
	ForeignColumnName string `pg:"foreign_column_name"`
	UpdateType        string `pg:"confupdtype"`
	DeleteType        string `pg:"confdeltype"`
}

//UniqueKeySchema : Unique Schema Model
type UniqueKeySchema struct {
	ConstraintName string `pg:"conname"`
	Columns        string `pg:"col"`
}

//StructColumnSchema : Struct Column Schema
type StructColumnSchema struct {
	TableName         string
	ColumnName        string
	ColumnDefault     string
	DataType          string
	IsNullable        string
	CharMaxLen        string
	ConstraintType    string
	IsDeferrable      string
	InitiallyDeferred string
	ForiegnTableName  string
	ForeignColumnName string
	UpdateType        string
	DeleteType        string
}
