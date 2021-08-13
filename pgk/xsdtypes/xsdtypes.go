package xsdtypes

import (
	"strconv"
)

type notation struct {
	Id, Name, Public, System string
}

type Notations map[string]*notation

func (me Notations) Add(id, name, public, system string) {
	me[name] = &notation{Id: id, Name: name, Public: public, System: system}
}

//	In XSD, the type xsd:anySimpleType is the base type from which all other built-in types are derived.
type AnySimpleType string

//	Since this is just a simple String type, this merely sets the current value from the specified string.
func (me *AnySimpleType) Set(v string) {
	*me = AnySimpleType(v)
}

//	Since this is just a simple String type, this merely returns its current string value.
func (me AnySimpleType) String() string {
	return string(me)
}

//	A convenience interface that declares a type conversion to AnySimpleType.
type ToXsdtAnySimpleType interface {
	ToXsdtAnySimpleType() AnySimpleType
}

//	In XSD, represents any simple or complex type. In Go, we hope no one schema ever uses it.
type AnyType string

//	Since this is just a simple String type, this merely sets the current value from the specified string.
func (me *AnyType) Set(v string) {
	*me = AnyType(v)
}

//	Since this is just a simple String type, this merely returns its current string value.
func (me AnyType) String() string {
	return string(me)
}

//	A convenience interface that declares a type conversion to AnyType.
type ToXsdtAnyType interface {
	ToXsdtAnyType() AnyType
}

//	Represents a URI as defined by RFC 2396. An anyURI value can be absolute or relative, and may have an optional fragment identifier.
type AnyURI string

//	Since this is just a simple String type, this merely sets the current value from the specified string.
func (me *AnyURI) Set(v string) {
	*me = AnyURI(v)
}

//	Since this is just a simple String type, this merely returns its current string value.
func (me AnyURI) String() string {
	return string(me)
}

//	A convenience interface that declares a type conversion to AnyURI.
type ToXsdtAnyURI interface {
	ToXsdtAnyURI() AnyURI
}

//	Represents Base64-encoded arbitrary binary data. A base64Binary is the set of finite-length sequences of binary octets.
type Base64Binary string // []byte

//	Since this is just a simple String type, this merely sets the current value from the specified string.
func (me *Base64Binary) Set(v string) {
	*me = Base64Binary(v)
}

//	Since this is just a simple String type, this merely returns its current string value.
func (me Base64Binary) String() string {
	return string(me)
}

//	A convenience interface that declares a type conversion to Base64Binary.
type ToXsdtBase64Binary interface {
	ToXsdtBase64Binary() Base64Binary
}

//	Represents Boolean values, which are either true or false.
type Boolean bool

//	Because littering your code with type conversions is a hassle...
func (me Boolean) B() bool {
	return bool(me)
}

//	Since this is a non-string scalar type, sets its current value obtained from parsing the specified string.
func (me *Boolean) Set(v string) {
	//	most schemas use true and false but sadly, a very few rare ones *do* use "0" and "1"...
	switch v {
	case "0":
		*me = false
	case "1":
		*me = true
	default:
		b, _ := strconv.ParseBool(v)
		*me = Boolean(b)
	}
}

//	Returns a string representation of its current non-string scalar value.
func (me Boolean) String() string {
	return strconv.FormatBool(bool(me))
}

//	A convenience interface that declares a type conversion to Boolean.
type ToXsdtBoolean interface {
	ToXsdtBoolean() Boolean
}

//	Represents an integer with a minimum value of -128 and maximum of 127.
type Byte int8

//	Because littering your code with type conversions is a hassle...
func (me Byte) N() int8 {
	return int8(me)
}

//	Since this is a non-string scalar type, sets its current value obtained from parsing the specified string.
func (me *Byte) Set(s string) {
	v, _ := strconv.ParseInt(s, 0, 8)
	*me = Byte(v)
}

//	Returns a string representation of its current non-string scalar value.
func (me Byte) String() string {
	return strconv.FormatInt(int64(me), 10)
}

//	A convenience interface that declares a type conversion to Byte.
type ToXsdtByte interface {
	ToXsdtByte() Byte
}

//	Represents a calendar date.
//	The pattern for date is CCYY-MM-DD with optional time zone indicator as allowed for dateTime.
type Date string // time.Time

//	Since this is just a simple String type, this merely sets the current value from the specified string.
func (me *Date) Set(v string) {
	*me = Date(v)
}

//	Since this is just a simple String type, this merely returns its current string value.
func (me Date) String() string {
	return string(me)
}

//	A convenience interface that declares a type conversion to Date.
type ToXsdtDate interface {
	ToXsdtDate() Date
}

//	Represents a specific instance of time.
type DateTime string // time.Time

//	Since this is just a simple String type, this merely sets the current value from the specified string.
func (me *DateTime) Set(v string) {
	*me = DateTime(v)
}

//	Since this is just a simple String type, this merely returns its current string value.
func (me DateTime) String() string {
	return string(me)
}

//	A convenience interface that declares a type conversion to DateTime.
type ToXsdtDateTime interface {
	ToXsdtDateTime() DateTime
}

//	Represents a specific instance of time.
type Time string // time.Time

//	Since this is just a simple String type, this merely sets the current value from the specified string.
func (me *Time) Set(v string) {
	*me = Time(v)
}

//	Since this is just a simple String type, this merely returns its current string value.
func (me Time) String() string {
	return string(me)
}

//	A convenience interface that declares a type conversion to Time.
type ToXsdtTime interface {
	ToXsdtTime() Time
}

//	Represents arbitrary precision numbers.
type Decimal string // complex128

//	Since this is just a simple String type, this merely sets the current value from the specified string.
func (me *Decimal) Set(v string) {
	*me = Decimal(v)
}

//	Since this is just a simple String type, this merely returns its current string value.
func (me Decimal) String() string {
	return string(me)
}

//	A convenience interface that declares a type conversion to Decimal.
type ToXsdtDecimal interface {
	ToXsdtDecimal() Decimal
}

//	Represents double-precision 64-bit floating-point numbers.
type Double float64

//	Because littering your code with type conversions is a hassle...
func (me Double) N() float64 {
	return float64(me)
}

//	Since this is a non-string scalar type, sets its current value obtained from parsing the specified string.
func (me *Double) Set(s string) {
	v, _ := strconv.ParseFloat(s, 64)
	*me = Double(v)
}

//	Returns a string representation of its current non-string scalar value.
func (me Double) String() string {
	return strconv.FormatFloat(float64(me), 'f', 8, 64)
}

//	A convenience interface that declares a type conversion to Double.
type ToXsdtDouble interface {
	ToXsdtDouble() Double
}

//	Represents a duration of time.
type Duration string // time.Duration

//	Since this is just a simple String type, this merely sets the current value from the specified string.
func (me *Duration) Set(v string) {
	*me = Duration(v)
}

//	Since this is just a simple String type, this merely returns its current string value.
func (me Duration) String() string {
	return string(me)
}

//	A convenience interface that declares a type conversion to Duration.
type ToXsdtDuration interface {
	ToXsdtDuration() Duration
}

//	Represents the ENTITIES attribute type. Contains a set of values of type ENTITY.
type Entities string

//	Since this is just a simple String type, this merely sets the current value from the specified string.
func (me *Entities) Set(v string) {
	*me = Entities(v)
}

//	Since this is just a simple String type, this merely returns its current string value.
func (me Entities) String() string {
	return string(me)
}

//	This type declares a String containing a whitespace-separated list of values. This Values() method creates and returns a slice of all elements in that list.
func (me Entities) Values() (list []Entity) {
	spl := ListValues(string(me))
	list = make([]Entity, len(spl))
	for i, s := range spl {
		list[i].Set(s)
	}
	return
}

//	A convenience interface that declares a type conversion to Entities.
type ToXsdtEntities interface {
	ToXsdtEntities() Entities
}

//	This is a reference to an unparsed entity with a name that matches the specified name.
type Entity NCName

//	Since this is just a simple String type, this merely sets the current value from the specified string.
func (me *Entity) Set(v string) {
	*me = Entity(v)
}

//	Since this is just a simple String type, this merely returns its current string value.
func (me Entity) String() string {
	return string(me)
}

//	A convenience interface that declares a type conversion to Entity.
type ToXsdtEntity interface {
	ToXsdtEntity() Entity
}

//	Represents single-precision 32-bit floating-point numbers.
type Float float32

//	Because littering your code with type conversions is a hassle...
func (me Float) N() float32 {
	return float32(me)
}

//	Since this is a non-string scalar type, sets its current value obtained from parsing the specified string.
func (me *Float) Set(s string) {
	v, _ := strconv.ParseFloat(s, 32)
	*me = Float(v)
}

//	Returns a string representation of its current non-string scalar value.
func (me Float) String() string {
	return strconv.FormatFloat(float64(me), 'f', 8, 32)
}

//	A convenience interface that declares a type conversion to Float.
type ToXsdtFloat interface {
	ToXsdtFloat() Float
}

//	Represents a Gregorian day that recurs, specifically a day of the month such as the fifth day of the month. A gDay is the space of a set of calendar dates. Specifically, it is a set of one-day long, monthly periodic instances.
type GDay string

//	Since this is just a simple String type, this merely sets the current value from the specified string.
func (me *GDay) Set(v string) {
	*me = GDay(v)
}

//	Since this is just a simple String type, this merely returns its current string value.
func (me GDay) String() string {
	return string(me)
}

//	A convenience interface that declares a type conversion to GDay.
type ToXsdtGDay interface {
	ToXsdtGDay() GDay
}

//	Represents a Gregorian month that recurs every year. A gMonth is the space of a set of calendar months. Specifically, it is a set of one-month long, yearly periodic instances.
type GMonth string

//	Since this is just a simple String type, this merely sets the current value from the specified string.
func (me *GMonth) Set(v string) {
	*me = GMonth(v)
}

//	Since this is just a simple String type, this merely returns its current string value.
func (me GMonth) String() string {
	return string(me)
}

//	A convenience interface that declares a type conversion to GMonth.
type ToXsdtGMonth interface {
	ToXsdtGMonth() GMonth
}

//	Represents a specific Gregorian date that recurs, specifically a day of the year such as the third of May. A gMonthDay is the set of calendar dates. Specifically, it is a set of one-day long, annually periodic instances.
type GMonthDay string

//	Since this is just a simple String type, this merely sets the current value from the specified string.
func (me *GMonthDay) Set(v string) {
	*me = GMonthDay(v)
}

//	Since this is just a simple String type, this merely returns its current string value.
func (me GMonthDay) String() string {
	return string(me)
}

//	A convenience interface that declares a type conversion to GMonthDay.
type ToXsdtGMonthDay interface {
	ToXsdtGMonthDay() GMonthDay
}

//	Represents a Gregorian year. A set of one-year long, nonperiodic instances.
type GYear string

//	Since this is just a simple String type, this merely sets the current value from the specified string.
func (me *GYear) Set(v string) {
	*me = GYear(v)
}

//	Since this is just a simple String type, this merely returns its current string value.
func (me GYear) String() string {
	return string(me)
}

//	A convenience interface that declares a type conversion to GYear.
type ToXsdtGYear interface {
	ToXsdtGYear() GYear
}

//	Represents a specific Gregorian month in a specific Gregorian year. A set of one-month long, nonperiodic instances.
type GYearMonth string

//	Since this is just a simple String type, this merely sets the current value from the specified string.
func (me *GYearMonth) Set(v string) {
	*me = GYearMonth(v)
}

//	Since this is just a simple String type, this merely returns its current string value.
func (me GYearMonth) String() string {
	return string(me)
}

//	A convenience interface that declares a type conversion to GYearMonth.
type ToXsdtGYearMonth interface {
	ToXsdtGYearMonth() GYearMonth
}

//	Represents arbitrary hex-encoded binary data. A hexBinary is the set of finite-length sequences of binary octets. Each binary octet is encoded as a character tuple, consisting of two hexadecimal digits ([0-9a-fA-F]) representing the octet code.
type HexBinary string // []byte

//	Since this is just a simple String type, this merely sets the current value from the specified string.
func (me *HexBinary) Set(v string) {
	*me = HexBinary(v)
}

//	Since this is just a simple String type, this merely returns its current string value.
func (me HexBinary) String() string {
	return string(me)
}

//	A convenience interface that declares a type conversion to HexBinary.
type ToXsdtHexBinary interface {
	ToXsdtHexBinary() HexBinary
}

//	The ID must be a no-colon-name (NCName) and must be unique within an XML document.
type Id NCName

//	Since this is just a simple String type, this merely sets the current value from the specified string.
func (me *Id) Set(v string) {
	*me = Id(v)
}

//	Since this is just a simple String type, this merely returns its current string value.
func (me Id) String() string {
	return string(me)
}

//	A convenience interface that declares a type conversion to Id.
type ToXsdtId interface {
	ToXsdtId() Id
}

//	Represents a reference to an element that has an ID attribute that matches the specified ID. An IDREF must be an NCName and must be a value of an element or attribute of type ID within the XML document.
type Idref NCName

//	Since this is just a simple String type, this merely sets the current value from the specified string.
func (me *Idref) Set(v string) {
	*me = Idref(v)
}

//	Since this is just a simple String type, this merely returns its current string value.
func (me Idref) String() string {
	return string(me)
}

//	A convenience interface that declares a type conversion to Idref.
type ToXsdtIdref interface {
	ToXsdtIdref() Idref
}

//	Contains a set of values of type IDREF.
type Idrefs string

//	Since this is just a simple String type, this merely sets the current value from the specified string.
func (me *Idrefs) Set(v string) {
	*me = Idrefs(v)
}

//	Since this is just a simple String type, this merely returns its current string value.
func (me Idrefs) String() string {
	return string(me)
}

//	This type declares a String containing a whitespace-separated list of values. This Values() method creates and returns a slice of all elements in that list.
func (me Idrefs) Values() (list []Idref) {
	spl := ListValues(string(me))
	list = make([]Idref, len(spl))
	for i, s := range spl {
		list[i].Set(s)
	}
	return
}

//	A convenience interface that declares a type conversion to Idrefs.
type ToXsdtIdrefs interface {
	ToXsdtIdrefs() Idrefs
}

//	Represents an integer with a minimum value of -2147483648 and maximum of 2147483647.
type Int int32

//	Because littering your code with type conversions is a hassle...
func (me Int) N() int32 {
	return int32(me)
}

//	Since this is a non-string scalar type, sets its current value obtained from parsing the specified string.
func (me *Int) Set(s string) {
	v, _ := strconv.ParseInt(s, 0, 32)
	*me = Int(v)
}

//	Returns a string representation of its current non-string scalar value.
func (me Int) String() string {
	return strconv.FormatInt(int64(me), 10)
}

//	A convenience interface that declares a type conversion to Int.
type ToXsdtInt interface {
	ToXsdtInt() Int
}

//	Represents a sequence of decimal digits with an optional leading sign (+ or -).
type Integer int64

//	Because littering your code with type conversions is a hassle...
func (me Integer) N() int64 {
	return int64(me)
}

//	Since this is a non-string scalar type, sets its current value obtained from parsing the specified string.
func (me *Integer) Set(s string) {
	v, _ := strconv.ParseInt(s, 0, 64)
	*me = Integer(v)
}

//	Returns a string representation of its current non-string scalar value.
func (me Integer) String() string {
	return strconv.FormatInt(int64(me), 10)
}

//	A convenience interface that declares a type conversion to Integer.
type ToXsdtInteger interface {
	ToXsdtInteger() Integer
}

//	Represents natural language identifiers (defined by RFC 1766).
type Language Token

//	Since this is just a simple String type, this merely sets the current value from the specified string.
func (me *Language) Set(v string) {
	*me = Language(v)
}

//	Since this is just a simple String type, this merely returns its current string value.
func (me Language) String() string {
	return string(me)
}

//	A convenience interface that declares a type conversion to Language.
type ToXsdtLanguage interface {
	ToXsdtLanguage() Language
}

//	Represents an integer with a minimum value of -9223372036854775808 and maximum of 9223372036854775807.
type Long int64

//	Because littering your code with type conversions is a hassle...
func (me Long) N() int64 {
	return int64(me)
}

//	Since this is a non-string scalar type, sets its current value obtained from parsing the specified string.
func (me *Long) Set(s string) {
	v, _ := strconv.ParseInt(s, 0, 64)
	*me = Long(v)
}

//	Returns a string representation of its current non-string scalar value.
func (me Long) String() string {
	return strconv.FormatInt(int64(me), 10)
}

//	A convenience interface that declares a type conversion to Long.
type ToXsdtLong interface {
	ToXsdtLong() Long
}

//	Represents names in XML. A Name is a token that begins with a letter, underscore, or colon and continues with name characters (letters, digits, and other characters).
type Name Token

//	Since this is just a simple String type, this merely sets the current value from the specified string.
func (me *Name) Set(v string) {
	*me = Name(v)
}

//	Since this is just a simple String type, this merely returns its current string value.
func (me Name) String() string {
	return string(me)
}

//	A convenience interface that declares a type conversion to Name.
type ToXsdtName interface {
	ToXsdtName() Name
}

//	Represents noncolonized names. This data type is the same as Name, except it cannot begin with a colon.
type NCName Name

//	Since this is just a simple String type, this merely sets the current value from the specified string.
func (me *NCName) Set(v string) {
	*me = NCName(v)
}

//	Since this is just a simple String type, this merely returns its current string value.
func (me NCName) String() string {
	return string(me)
}

//	A convenience interface that declares a type conversion to NCName.
type ToXsdtNCName interface {
	ToXsdtNCName() NCName
}

//	Represents an integer that is less than zero. Consists of a negative sign (-) and sequence of decimal digits.
type NegativeInteger int64

//	Because littering your code with type conversions is a hassle...
func (me NegativeInteger) N() int64 {
	return int64(me)
}

//	Since this is a non-string scalar type, sets its current value obtained from parsing the specified string.
func (me *NegativeInteger) Set(s string) {
	v, _ := strconv.ParseInt(s, 0, 64)
	*me = NegativeInteger(v)
}

//	Returns a string representation of its current non-string scalar value.
func (me NegativeInteger) String() string {
	return strconv.FormatInt(int64(me), 10)
}

//	A convenience interface that declares a type conversion to NegativeInteger.
type ToXsdtNegativeInteger interface {
	ToXsdtNegativeInteger() NegativeInteger
}

//	An NMTOKEN is set of name characters (letters, digits, and other characters) in any combination. Unlike Name and NCName, NMTOKEN has no restrictions on the starting character.
type Nmtoken Token

//	Since this is just a simple String type, this merely sets the current value from the specified string.
func (me *Nmtoken) Set(v string) {
	*me = Nmtoken(v)
}

//	Since this is just a simple String type, this merely returns its current string value.
func (me Nmtoken) String() string {
	return string(me)
}

//	A convenience interface that declares a type conversion to Nmtoken.
type ToXsdtNmtoken interface {
	ToXsdtNmtoken() Nmtoken
}

//	Contains a set of values of type NMTOKEN.
type Nmtokens string

//	Since this is just a simple String type, this merely sets the current value from the specified string.
func (me *Nmtokens) Set(v string) {
	*me = Nmtokens(v)
}

//	Since this is just a simple String type, this merely returns its current string value.
func (me Nmtokens) String() string {
	return string(me)
}

//	This type declares a String containing a whitespace-separated list of values. This Values() method creates and returns a slice of all elements in that list.
func (me Nmtokens) Values() (list []Nmtoken) {
	spl := ListValues(string(me))
	list = make([]Nmtoken, len(spl))
	for i, s := range spl {
		list[i].Set(s)
	}
	return
}

//	A convenience interface that declares a type conversion to Nmtokens.
type ToXsdtNmtokens interface {
	ToXsdtNmtokens() Nmtokens
}

//	Represents an integer that is greater than or equal to zero.
type NonNegativeInteger uint64

//	Because littering your code with type conversions is a hassle...
func (me NonNegativeInteger) N() uint64 {
	return uint64(me)
}

//	Since this is a non-string scalar type, sets its current value obtained from parsing the specified string.
func (me *NonNegativeInteger) Set(s string) {
	v, _ := strconv.ParseUint(s, 0, 64)
	*me = NonNegativeInteger(v)
}

//	Returns a string representation of its current non-string scalar value.
func (me NonNegativeInteger) String() string {
	return strconv.FormatUint(uint64(me), 10)
}

//	A convenience interface that declares a type conversion to NonNegativeInteger.
type ToXsdtNonNegativeInteger interface {
	ToXsdtNonNegativeInteger() NonNegativeInteger
}

//	Represents an integer that is less than or equal to zero. A nonPositiveIntegerconsists of a negative sign (-) and sequence of decimal digits.
type NonPositiveInteger int64

//	Because littering your code with type conversions is a hassle...
func (me NonPositiveInteger) N() int64 {
	return int64(me)
}

//	Since this is a non-string scalar type, sets its current value obtained from parsing the specified string.
func (me *NonPositiveInteger) Set(s string) {
	v, _ := strconv.ParseInt(s, 0, 64)
	*me = NonPositiveInteger(v)
}

//	Returns a string representation of its current non-string scalar value.
func (me NonPositiveInteger) String() string {
	return strconv.FormatInt(int64(me), 10)
}

//	A convenience interface that declares a type conversion to NonPositiveInteger.
type ToXsdtNonPositiveInteger interface {
	ToXsdtNonPositiveInteger() NonPositiveInteger
}

//	Represents white space normalized strings.
type NormalizedString String

//	Since this is just a simple String type, this merely sets the current value from the specified string.
func (me *NormalizedString) Set(v string) {
	*me = NormalizedString(v)
}

//	Since this is just a simple String type, this merely returns its current string value.
func (me NormalizedString) String() string {
	return string(me)
}

//	A convenience interface that declares a type conversion to NormalizedString.
type ToXsdtNormalizedString interface {
	ToXsdtNormalizedS() NormalizedString
}

//	A set of QNames.
type Notation string

//	Since this is just a simple String type, this merely sets the current value from the specified string.
func (me *Notation) Set(v string) {
	*me = Notation(v)
}

//	Since this is just a simple String type, this merely returns its current string value.
func (me Notation) String() string {
	return string(me)
}

//	This type declares a String containing a whitespace-separated list of values. This Values() method creates and returns a slice of all elements in that list.
func (me Notation) Values() (list []QName) {
	spl := ListValues(string(me))
	list = make([]QName, len(spl))
	for i, s := range spl {
		list[i].Set(s)
	}
	return
}

//	A convenience interface that declares a type conversion to Notation.
type ToXsdtNotation interface {
	ToXsdtNotation() Notation
}

//	Represents an integer that is greater than zero.
type PositiveInteger uint64

//	Because littering your code with type conversions is a hassle...
func (me PositiveInteger) N() uint64 {
	return uint64(me)
}

//	Since this is a non-string scalar type, sets its current value obtained from parsing the specified string.
func (me *PositiveInteger) Set(s string) {
	v, _ := strconv.ParseUint(s, 0, 64)
	*me = PositiveInteger(v)
}

//	Returns a string representation of its current non-string scalar value.
func (me PositiveInteger) String() string {
	return strconv.FormatUint(uint64(me), 10)
}

//	A convenience interface that declares a type conversion to PositiveInteger.
type ToXsdtPositiveInteger interface {
	ToXsdtPositiveInteger() PositiveInteger
}

//	Represents a qualified name. A qualified name is composed of a prefix and a local name separated by a colon. Both the prefix and local names must be an NCName. The prefix must be associated with a namespace URI reference, using a namespace declaration.
type QName string

//	Since this is just a simple String type, this merely sets the current value from the specified string.
func (me *QName) Set(v string) {
	*me = QName(v)
}

//	Since this is just a simple String type, this merely returns its current string value.
func (me QName) String() string {
	return string(me)
}

//	A convenience interface that declares a type conversion to QName.
type ToXsdtQname interface {
	ToXsdtQname() QName
}

//	Represents an integer with a minimum value of -32768 and maximum of 32767.
type Short int16

//	Because littering your code with type conversions is a hassle...
func (me Short) N() int16 {
	return int16(me)
}

//	Since this is a non-string scalar type, sets its current value obtained from parsing the specified string.
func (me *Short) Set(s string) {
	v, _ := strconv.ParseInt(s, 0, 16)
	*me = Short(v)
}

//	Returns a string representation of its current non-string scalar value.
func (me Short) String() string {
	return strconv.FormatInt(int64(me), 10)
}

//	A convenience interface that declares a type conversion to Short.
type ToXsdtShort interface {
	ToXsdtShort() Short
}

//	Represents character strings.
type String string

//	Since this is just a simple String type, this merely sets the current value from the specified string.
func (me *String) Set(v string) {
	*me = String(v)
}

//	Since this is just a simple String type, this merely returns its current string value.
func (me String) String() string {
	return string(me)
}

//	A convenience interface that declares a type conversion to String.
type ToXsdtString interface {
	ToXsdtString() String
}

//	Represents tokenized strings.
type Token NormalizedString

//	Since this is just a simple String type, this merely sets the current value from the specified string.
func (me *Token) Set(v string) {
	*me = Token(v)
}

//	Since this is just a simple String type, this merely returns its current string value.
func (me Token) String() string {
	return string(me)
}

//	A convenience interface that declares a type conversion to Token.
type ToXsdtToken interface {
	ToXsdtToken() Token
}

//	Represents an integer with a minimum of zero and maximum of 255.
type UnsignedByte uint8

//	Because littering your code with type conversions is a hassle...
func (me UnsignedByte) N() uint8 {
	return uint8(me)
}

//	Since this is a non-string scalar type, sets its current value obtained from parsing the specified string.
func (me *UnsignedByte) Set(s string) {
	v, _ := strconv.ParseUint(s, 0, 8)
	*me = UnsignedByte(v)
}

//	Returns a string representation of its current non-string scalar value.
func (me UnsignedByte) String() string {
	return strconv.FormatUint(uint64(me), 10)
}

//	A convenience interface that declares a type conversion to UnsignedByte.
type ToXsdtUnsignedByte interface {
	ToXsdtUnsignedByte() UnsignedByte
}

//	Represents an integer with a minimum of zero and maximum of 4294967295.
type UnsignedInt uint32

//	Because littering your code with type conversions is a hassle...
func (me UnsignedInt) N() uint32 {
	return uint32(me)
}

//	Since this is a non-string scalar type, sets its current value obtained from parsing the specified string.
func (me *UnsignedInt) Set(s string) {
	v, _ := strconv.ParseUint(s, 0, 32)
	*me = UnsignedInt(v)
}

//	Returns a string representation of its current non-string scalar value.
func (me UnsignedInt) String() string {
	return strconv.FormatUint(uint64(me), 10)
}

//	A convenience interface that declares a type conversion to UnsignedInt.
type ToXsdtUnsignedInt interface {
	ToXsdtUnsignedInt() UnsignedInt
}

//	Represents an integer with a minimum of zero and maximum of 18446744073709551615.
type UnsignedLong uint64

//	Because littering your code with type conversions is a hassle...
func (me UnsignedLong) N() uint64 {
	return uint64(me)
}

//	Since this is a non-string scalar type, sets its current value obtained from parsing the specified string.
func (me *UnsignedLong) Set(s string) {
	v, _ := strconv.ParseUint(s, 0, 64)
	*me = UnsignedLong(v)
}

//	Returns a string representation of its current non-string scalar value.
func (me UnsignedLong) String() string {
	return strconv.FormatUint(uint64(me), 10)
}

//	A convenience interface that declares a type conversion to UnsignedLong.
type ToXsdtUnsignedLong interface {
	ToXsdtUnsignedLong() UnsignedLong
}

//	Represents an integer with a minimum of zero and maximum of 65535.
type UnsignedShort uint16

//	Because littering your code with type conversions is a hassle...
func (me UnsignedShort) N() uint16 {
	return uint16(me)
}

//	Since this is a non-string scalar type, sets its current value obtained from parsing the specified string.
func (me *UnsignedShort) Set(s string) {
	v, _ := strconv.ParseUint(s, 0, 16)
	*me = UnsignedShort(v)
}

//	Returns a string representation of its current non-string scalar value.
func (me UnsignedShort) String() string {
	return strconv.FormatUint(uint64(me), 10)
}

//	A convenience interface that declares a type conversion to UnsignedShort.
type ToXsdtUnsignedShort interface {
	ToXsdtUnsignedShort() UnsignedShort
}

// XSD "list" types are always space-separated strings. All generated Go types based on any XSD's list types get a Values() method, which will always resort to this function.
func ListValues(v string) (spl []string) {
	if len(v) == 0 {
		return
	}
	lastWs := true
	wsr := func(r rune) bool {
		return (r == ' ') || (r == '\r') || (r == '\n') || (r == '\t')
	}
	wss := func(r string) bool {
		return (r == " ") || (r == "\r") || (r == "\n") || (r == "\t")
	}
	for wss(v[len(v)-1:]) {
		v = v[:len(v)-1]
	}
	for wss(v[:1]) {
		v = v[1:]
	}
	if len(v) > 0 {
		cur, num, i := "", 1, 0
		for _, r := range v {
			if wsr(r) {
				if !lastWs {
					num++
					lastWs = true
				}
			} else {
				lastWs = false
			}
		}
		lastWs, spl = true, make([]string, num)
		for _, r := range v {
			if wsr(r) {
				if !lastWs {
					if len(cur) > 0 {
						spl[i] = cur
						i++
					}
					cur, lastWs = "", true
				}
			} else {
				lastWs = false
				cur += string(r)
			}
		}
		if len(cur) > 0 {
			spl[i] = cur
		}
	}
	return
}

func ListValuesBoolean(vals []Boolean) (sl []bool) {
	sl = make([]bool, len(vals))
	for i, b := range vals {
		sl[i] = b.B()
	}
	return
}

func ListValuesDouble(vals []Double) (sl []float64) {
	sl = make([]float64, len(vals))
	for i, d := range vals {
		sl[i] = d.N()
	}
	return
}

func ListValuesLong(vals []Long) (sl []int64) {
	sl = make([]int64, len(vals))
	for i, l := range vals {
		sl[i] = l.N()
	}
	return
}

//	A helper function for the Walk() functionality of generated wrapper packages.
func OnWalkError(err *error, slice *[]error, breakWalk bool, handler func(error)) (ret bool) {
	if e := *err; e != nil {
		*slice = append(*slice, e)
		ret = breakWalk
		if handler != nil {
			handler(e)
		}
	}
	*err = nil
	return
}
