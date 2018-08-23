package ddl

type CreateTableStmt struct {
	*Builder

	// By convention, the struct members are public and end with "Value", so we
	// can use the name without any suffix as a builder method, e.g.
	// IfNotExistsValue is set with the IfNotExists() method.  We make an exception
	// for lists of things like PrimaryKeys, since the method name is PrimaryKey()

	NameValue string

	IfNotExistsValue bool

	Columns DataTypeDefPtrList

	PrimaryKeys []string

	ForeignKeys []*CreateTableFKDef
}

type CreateTableFKDef struct {
	ColumnValue      string
	OtherTableValue  string
	OtherColumnValue string
}

type CreateTableColDef struct {
	*CreateTableStmt
	*DataTypeDef

	// TODO: specific context and funcs so they can do Null() and Default()
	// but then embedding *CreateTableStmt means they can also do the next
	// column, e.g.
	// builder.CreateTable("blah").
	// 	ColVarchar("something").Default("empty_something").
	//  ColVarchar("something").Null().Default(nil).
	// ... etc.

	// we do need some way to describe these columns though that is database agnostic...
	// and store them in CreateTableStmt...  Probably a type for each column type and
	// embed things like Nullable, Defaultable, etc. that has the field and the method(s)
	// formatters will end up switch()ing over these types anyway so it's really just
	// to organize the code and avoid duplication - but basically each type is just different
}

func (s *CreateTableStmt) IsStmt() {}

func (s *CreateTableStmt) IfNotExists() *CreateTableStmt {
	s.IfNotExistsValue = true
	return s
}

func (s *CreateTableStmt) Column(name string, dataType DataType) *CreateTableColDef {
	dtd := &DataTypeDef{NameValue: name, DataTypeValue: dataType}
	s.Columns = append(s.Columns, dtd)
	return &CreateTableColDef{
		CreateTableStmt: s,
		DataTypeDef:     dtd,
	}
}

func (s *CreateTableStmt) ColumnCustom(name, customSQL string) *CreateTableColDef {
	dtd := &DataTypeDef{NameValue: name, DataTypeValue: Custom, CustomSQLValue: customSQL}
	s.Columns = append(s.Columns, dtd)
	return &CreateTableColDef{
		CreateTableStmt: s,
		DataTypeDef:     dtd,
	}
}

func (def *CreateTableColDef) Null() *CreateTableColDef {
	def.DataTypeDef.NullValue = true
	return def
}

func (def *CreateTableColDef) Default(value interface{}) *CreateTableColDef {
	def.DataTypeDef.DefaultValue = value
	return def
}

func (def *CreateTableColDef) Length(length int) *CreateTableColDef {
	def.DataTypeDef.LengthValue = length
	return def
}

func (def *CreateTableColDef) CaseSensitive() *CreateTableColDef {
	def.DataTypeDef.CaseSensitiveValue = true
	return def
}

func (def *CreateTableColDef) ForiegnKey(otherTable, otherColumn string) *CreateTableColDef {
	def.CreateTableStmt.ForeignKeys = append(def.CreateTableStmt.ForeignKeys, &CreateTableFKDef{
		ColumnValue:      def.CreateTableStmt.NameValue,
		OtherTableValue:  otherTable,
		OtherColumnValue: otherColumn,
	})
	return def
}

func (def *CreateTableColDef) PrimaryKey() *CreateTableColDef {
	def.CreateTableStmt.PrimaryKeys = append(def.CreateTableStmt.PrimaryKeys, def.DataTypeDef.NameValue)
	return def
}

type DropTableStmt struct {
	*Builder

	NameValue string
}

func (s *DropTableStmt) IsStmt() {}
