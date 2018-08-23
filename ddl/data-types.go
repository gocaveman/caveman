package ddl

// types:
// VarCharPK (mysql: use "ascii" for charset and "binary" for collation)
// BigIntAutoPK
// VarCharFK (mysql: use "ascii" for charset and "binary" for collation)
// BigIntFK
// Int
// IntU
// BigInt
// BigIntU
// Double
// DateTime
// VarChar (needs length option, sensible default) (also CaseSensitive() - sets collation on mysql, nop on others)
// Bool
// Text
// Blob
// variations: null, default value, foreign key
// Custom (can just provide whatever they want)

// type Namer interface {
// 	SetName(name string)
// }

// type namer string

// func (n *namer) SetName(name string) {
// 	*n = name
// }
// func (n *namer) Name() string {
// 	return string(*n)
// }

// // Nuller allows you to specify the "NULL" option on a SQL field (otherwise is NOT NULL).
// type Nuller interface {
// 	Null()
// }

// type nuller bool

// func (n *nuller) Null() {
// 	*n = true
// }
// func (n *nuller) IsNull() bool {
// 	return *n
// }

// // Defaulter allows you to specify a default value as part of a data type.
// type Defaulter interface {
// 	Default(value interface{})
// }

// type defaulter interface{}

// func (d *defaulter) Default(value interface{}) {
// 	*d = value
// }
// func (d *defaulter) DefaultValue() interface{} {
// 	return *d
// }

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
)

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
