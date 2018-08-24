package ddl

//go:generate stringer -type=DataType -output=data-types_string.go

type DataType int

// enum so we can refer to data types with a simple value
const (
	Invalid DataType = iota
	Custom
	VarCharPK
	BigIntAutoPK
	VarCharFK
	BigIntFK
	Int
	IntU
	BigInt
	BigIntU
	Double
	DateTime
	VarChar
	Bool
	Text
	Blob
	// TODO: We actually need to add a DECIMAL type - in MySQL in can be DECIMAL/NUMERIC
	// and in SQLite3 it can be just a string.  This is vital for financial and other
	// calculations where rounding errors and the general imprecision of floating points
	// have big bad consequences.  It will be important to also think about what the recommended
	// corresponding type(s) are in Go, e.g. big.Rat or something else.
)

// DataTypeDef describes a column.  It includes the name and various common options.
type DataTypeDef struct {
	DataTypeValue      DataType    // which basic data type
	NameValue          string      // SQL name
	CustomSQLValue     string      // if DataType is Custom this is the full SQL of the type (not including the name)
	NullValue          bool        // NULL (or default of NOT NULL)
	DefaultValue       interface{} // if you want a DEFAULT included
	LengthValue        int         // for types that support a length part
	CaseSensitiveValue bool        // for cases where a string should be forced to be case sensitive
}

type DataTypeDefPtrList []*DataTypeDef
