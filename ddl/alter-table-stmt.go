package ddl

type AlterTableRenameStmt struct {
	*Builder

	OldNameValue string
	NewNameValue string
}

func (s *AlterTableRenameStmt) IsStmt() {}

type AlterTableAddStmt struct {
	*Builder

	NameValue string // table name

	DataTypeDef DataTypeDef
}

func (s *AlterTableAddStmt) IsStmt() {}

func (s *AlterTableAddStmt) Column(name string, dataType DataType) *AlterTableAddStmt {
	s.DataTypeDef = DataTypeDef{NameValue: name, DataTypeValue: dataType}
	return s
}

func (s *AlterTableAddStmt) ColumnCustom(name, customSQL string) *AlterTableAddStmt {
	s.DataTypeDef = DataTypeDef{NameValue: name, DataTypeValue: Custom, CustomSQLValue: customSQL}
	return s
}

func (s *AlterTableAddStmt) Null() *AlterTableAddStmt {
	s.DataTypeDef.NullValue = true
	return s
}

func (s *AlterTableAddStmt) Default(value interface{}) *AlterTableAddStmt {
	s.DataTypeDef.DefaultValue = value
	return s
}

func (s *AlterTableAddStmt) Length(length int) *AlterTableAddStmt {
	s.DataTypeDef.LengthValue = length
	return s
}

func (s *AlterTableAddStmt) CaseSensitive() *AlterTableAddStmt {
	s.DataTypeDef.CaseSensitiveValue = true
	return s
}
