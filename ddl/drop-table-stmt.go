package ddl

type DropTableStmt struct {
	*Builder

	NameValue string
}

func (s *DropTableStmt) IsStmt() {}
