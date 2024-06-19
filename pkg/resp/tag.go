package resp

// Tag represents a RESP value type, as a single prefix byte
type Tag byte

//go:generate go run golang.org/x/tools/cmd/stringer -type=Tag -linecomment
const (
	SimpleStringTag   Tag = '+' // simple string
	SimpleErrorTag    Tag = '-' // simple error
	IntegerTag        Tag = ':' // integer
	BulkStringTag     Tag = '$' // bulk string
	ArrayTag          Tag = '*' // array
	NullTag           Tag = '_' // null
	BooleanTag        Tag = '#' // boolean
	DoubleTag         Tag = ',' // double
	BigNumberTag      Tag = '(' // big number
	BulkErrorTag      Tag = '!' // bulk error
	VerbatimStringTag Tag = '=' // verbose string
	MapTag            Tag = '%' // map
	AttributeTag      Tag = '|' // attribute
	SetTag            Tag = '~' // data
	PushTag           Tag = '>' // push
)
