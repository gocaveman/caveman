package ddl

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"

	_ "github.com/mattn/go-sqlite3"
)

// TestSQLite3 tests each feature against a SQLite3 database to ensure syntax is correct.
func TestSQLite3(t *testing.T) {

	assert := assert.New(t)

	f := NewSQLite3Formatter()

	b := New()
	b.SetCategory("test")

	db, err := sql.Open("sqlite3", `file:TestSQLite3?mode=memory&cache=shared`)
	if err != nil {
		t.Fatal(err)
	}

	runSQL := func(up, _ []string, err error) {
		if err != nil {
			assert.NoError(err)
			return
		}
		for _, s := range up {
			t.Logf("Running SQL: %s", s)
			_, err = db.Exec(s)
			assert.NoError(err)
		}
	}

	// -- create table

	// one of each type on it (except the integer pk)
	runSQL(b.Reset().
		CreateTable("table_types").
		Column("table_types_id", VarCharPK).PrimaryKey().
		ColumnCustom("test_custom", "TEXT NOT NULL").
		Column("test_varcharfk", VarCharFK).Length(255). // NOTE: lengths are ignored by SQLite, they don't get output
		Column("test_bigintfk", BigIntFK).
		Column("test_int", Int).
		Column("test_intu", IntU).
		Column("test_bigint", BigInt).
		Column("test_bigintu", BigIntU).
		Column("test_double", Double).
		Column("test_datetime", DateTime).
		Column("test_varchar", VarChar).Length(255).
		Column("test_bool", Bool).
		Column("test_text", Text).
		Column("test_blob", Blob).
		MakeSQL(f))

	// integer autoinc pk
	runSQL(b.Reset().
		CreateTable("table_autoinc").
		Column("table_autoinc_id", BigIntAutoPK).PrimaryKey().
		Column("test_varchar", VarChar).
		MakeSQL(f))

	// mulitple pks
	runSQL(b.Reset().
		CreateTable("table_join").
		Column("table_join_a_id", VarCharPK).PrimaryKey().
		Column("table_join_b_id", VarCharPK).PrimaryKey().
		MakeSQL(f))

	// if not exists
	runSQL(b.Reset().
		CreateTable("table_existential").IfNotExists().
		Column("table_existential_id", VarCharPK).PrimaryKey().
		MakeSQL(f))

	// null
	runSQL(b.Reset().
		CreateTable("table_null").
		Column("table_null_id", VarCharPK).PrimaryKey().
		Column("test_int", Int).Null().
		Column("test_intu", IntU).Null().
		Column("test_bigint", BigInt).Null().
		Column("test_bigintu", BigIntU).Null().
		Column("test_double", Double).Null().
		Column("test_datetime", DateTime).Null().
		Column("test_varchar", VarChar).Length(255).Null().
		Column("test_bool", Bool).Null().
		Column("test_text", Text).Null().
		Column("test_blob", Blob).Null().
		MakeSQL(f))

	// case sensitive
	runSQL(b.Reset().
		CreateTable("table_cs").
		Column("table_cs_id", VarCharPK).PrimaryKey().
		Column("test_varchar_cs", VarChar).CaseSensitive().
		Column("test_text_cs", Text).CaseSensitive().
		MakeSQL(f))

	// -- drop table
	runSQL(b.Reset().
		DropTable("table_cs").
		MakeSQL(f))

	// -- rename table
	runSQL(b.Reset().
		AlterTableRename("table_null", "table_null2").
		MakeSQL(f))

	// -- add column
	runSQL(b.Reset().
		AlterTableAdd("table_existential").
		Column("other_cool_field", VarChar).Null().Default("mozdef").CaseSensitive().
		MakeSQL(f))

	// -- create index
	runSQL(b.Reset().
		CreateIndex("table_existential_other", "table_existential").Columns("other_cool_field").
		MakeSQL(f))

	// -- drop index
	runSQL(b.Reset().
		DropIndex("table_existential_other", "table_existential").
		MakeSQL(f))

	// -- more create index

	// unique
	runSQL(b.Reset().
		CreateIndex("table_existential_other", "table_existential").
		Unique().
		Columns("other_cool_field").
		MakeSQL(f))

	// if not exists
	runSQL(b.Reset().
		CreateIndex("table_existential_other", "table_existential").
		IfNotExists().
		Columns("other_cool_field").
		MakeSQL(f))

}
