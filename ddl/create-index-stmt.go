package ddl

type CreateIndexStmt struct {
	*Builder

	NameValue        string // index name
	TableNameValue   string
	UniqueValue      bool
	IfNotExistsValue bool // TODO: consider removing this, not supported by MySQL and not really needed

	ColumnNames []string
}

func (s *CreateIndexStmt) IsStmt() {}

func (s *CreateIndexStmt) IfNotExists() *CreateIndexStmt {
	s.IfNotExistsValue = true
	return s
}

func (s *CreateIndexStmt) Unique() *CreateIndexStmt {
	s.UniqueValue = true
	return s
}

func (s *CreateIndexStmt) Columns(name ...string) *CreateIndexStmt {
	s.ColumnNames = append(s.ColumnNames, name...)
	return s
}

type DropIndexStmt struct {
	*Builder

	NameValue      string // index name
	TableNameValue string
}

func (s *DropIndexStmt) IsStmt() {}
