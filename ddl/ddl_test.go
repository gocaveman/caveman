package ddl

// func TestDDL1(t *testing.T) {

// 	assert := assert.New(t)

// 	// formatters := NewFormatterList(NewSQLite3Formatter())
// 	f := NewSQLite3Formatter()

// 	b := New()
// 	b.SetCategory("test")

// 	up, down, err := b.SetVersion("0001").
// 		Up().
// 		CreateTable("widget").
// 		Column("widget_id", VarcharPK).
// 		Column("name", Varchar).Length(255).PrimaryKey().
// 		Down().
// 		DropTable("widget").
// 		MakeSQL(f)

// 	assert.NoError(err)
// 	assert.Len(up, 1)
// 	assert.Len(down, 1)

// 	t.Logf("up: %v", up)
// 	t.Logf("down: %v", down)

// }
