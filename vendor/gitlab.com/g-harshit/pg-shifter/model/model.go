package model

//ColSchema : Table Column Schema Model
type ColSchema struct {
	TableName         string `pg:"-"`
	StructColumnName  string `pg:"-"`
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
	SeqName           string `pg:"seq_name"`
	SeqDataType       string `pg:"seq_data_type"`
	Position          int    `pg:"position"`
	IsFkUnique        bool   `pg:"-"`
	FkUniqueName      string `pg:"-"`
	DefaultExists     bool   `pg:"-"`
}

//UKSchema : Unique Schema Model
type UKSchema struct {
	ConstraintName string `pg:"conname"`
	Columns        string `pg:"col"`
}

//Index model
type Index struct {
	IdxName string `pg:"index_name"`
	IType   string `pg:"itype"`
	Columns string `pg:"col"`
}
