# xsdt
--
    import "github.com/metaleap/go-xsd/types"

A tiny package imported by all "go-xsd"-generated packages.

Maps all XSD built-in simple-types to Go types, which affords us easy mapping of
any XSD type references in the schema to Go imports: every xs:string and
xs:boolean automatically becomes xsdt.String and xsdt.Boolean etc. Types are
mapped to Go types depending on how encoding/xml.Unmarshal() can handle them:
ie. it parses bools and numbers, but dates/durations have too many format
mismatches and thus are just declared string types. Same for base64- and
hex-encoded binary data: since Unmarshal() won't decode them, we leave them as
strings. If you need their binary data, your code needs to import Go's
base64/hex codec packages and use them as necessary.

## Usage

#### func  ListValues

```go
func ListValues(v string) (spl []string)
```
XSD "list" types are always space-separated strings. All generated Go types
based on any XSD's list types get a Values() method, which will always resort to
this function.

#### func  ListValuesBoolean

```go
func ListValuesBoolean(vals []Boolean) (sl []bool)
```

#### func  ListValuesDouble

```go
func ListValuesDouble(vals []Double) (sl []float64)
```

#### func  ListValuesLong

```go
func ListValuesLong(vals []Long) (sl []int64)
```

#### func  OnWalkError

```go
func OnWalkError(err *error, slice *[]error, breakWalk bool, handler func(error)) (ret bool)
```
A helper function for the Walk() functionality of generated wrapper packages.

#### type AnySimpleType

```go
type AnySimpleType string
```

In XSD, the type xsd:anySimpleType is the base type from which all other
built-in types are derived.

#### func (*AnySimpleType) Set

```go
func (me *AnySimpleType) Set(v string)
```
Since this is just a simple String type, this merely sets the current value from
the specified string.

#### func (AnySimpleType) String

```go
func (me AnySimpleType) String() string
```
Since this is just a simple String type, this merely returns its current string
value.

#### type AnyType

```go
type AnyType string
```

In XSD, represents any simple or complex type. In Go, we hope no one schema ever
uses it.

#### func (*AnyType) Set

```go
func (me *AnyType) Set(v string)
```
Since this is just a simple String type, this merely sets the current value from
the specified string.

#### func (AnyType) String

```go
func (me AnyType) String() string
```
Since this is just a simple String type, this merely returns its current string
value.

#### type AnyURI

```go
type AnyURI string
```

Represents a URI as defined by RFC 2396. An anyURI value can be absolute or
relative, and may have an optional fragment identifier.

#### func (*AnyURI) Set

```go
func (me *AnyURI) Set(v string)
```
Since this is just a simple String type, this merely sets the current value from
the specified string.

#### func (AnyURI) String

```go
func (me AnyURI) String() string
```
Since this is just a simple String type, this merely returns its current string
value.

#### type Base64Binary

```go
type Base64Binary string // []byte

```

Represents Base64-encoded arbitrary binary data. A base64Binary is the set of
finite-length sequences of binary octets.

#### func (*Base64Binary) Set

```go
func (me *Base64Binary) Set(v string)
```
Since this is just a simple String type, this merely sets the current value from
the specified string.

#### func (Base64Binary) String

```go
func (me Base64Binary) String() string
```
Since this is just a simple String type, this merely returns its current string
value.

#### type Boolean

```go
type Boolean bool
```

Represents Boolean values, which are either true or false.

#### func (Boolean) B

```go
func (me Boolean) B() bool
```
Because littering your code with type conversions is a hassle...

#### func (*Boolean) Set

```go
func (me *Boolean) Set(v string)
```
Since this is a non-string scalar type, sets its current value obtained from
parsing the specified string.

#### func (Boolean) String

```go
func (me Boolean) String() string
```
Returns a string representation of its current non-string scalar value.

#### type Byte

```go
type Byte int8
```

Represents an integer with a minimum value of -128 and maximum of 127.

#### func (Byte) N

```go
func (me Byte) N() int8
```
Because littering your code with type conversions is a hassle...

#### func (*Byte) Set

```go
func (me *Byte) Set(s string)
```
Since this is a non-string scalar type, sets its current value obtained from
parsing the specified string.

#### func (Byte) String

```go
func (me Byte) String() string
```
Returns a string representation of its current non-string scalar value.

#### type Date

```go
type Date string // time.Time

```

Represents a calendar date. The pattern for date is CCYY-MM-DD with optional
time zone indicator as allowed for dateTime.

#### func (*Date) Set

```go
func (me *Date) Set(v string)
```
Since this is just a simple String type, this merely sets the current value from
the specified string.

#### func (Date) String

```go
func (me Date) String() string
```
Since this is just a simple String type, this merely returns its current string
value.

#### type DateTime

```go
type DateTime string // time.Time

```

Represents a specific instance of time.

#### func (*DateTime) Set

```go
func (me *DateTime) Set(v string)
```
Since this is just a simple String type, this merely sets the current value from
the specified string.

#### func (DateTime) String

```go
func (me DateTime) String() string
```
Since this is just a simple String type, this merely returns its current string
value.

#### type Decimal

```go
type Decimal string // complex128

```

Represents arbitrary precision numbers.

#### func (*Decimal) Set

```go
func (me *Decimal) Set(v string)
```
Since this is just a simple String type, this merely sets the current value from
the specified string.

#### func (Decimal) String

```go
func (me Decimal) String() string
```
Since this is just a simple String type, this merely returns its current string
value.

#### type Double

```go
type Double float64
```

Represents double-precision 64-bit floating-point numbers.

#### func (Double) N

```go
func (me Double) N() float64
```
Because littering your code with type conversions is a hassle...

#### func (*Double) Set

```go
func (me *Double) Set(s string)
```
Since this is a non-string scalar type, sets its current value obtained from
parsing the specified string.

#### func (Double) String

```go
func (me Double) String() string
```
Returns a string representation of its current non-string scalar value.

#### type Duration

```go
type Duration string // time.Duration

```

Represents a duration of time.

#### func (*Duration) Set

```go
func (me *Duration) Set(v string)
```
Since this is just a simple String type, this merely sets the current value from
the specified string.

#### func (Duration) String

```go
func (me Duration) String() string
```
Since this is just a simple String type, this merely returns its current string
value.

#### type Entities

```go
type Entities string
```

Represents the ENTITIES attribute type. Contains a set of values of type ENTITY.

#### func (*Entities) Set

```go
func (me *Entities) Set(v string)
```
Since this is just a simple String type, this merely sets the current value from
the specified string.

#### func (Entities) String

```go
func (me Entities) String() string
```
Since this is just a simple String type, this merely returns its current string
value.

#### func (Entities) Values

```go
func (me Entities) Values() (list []Entity)
```
This type declares a String containing a whitespace-separated list of values.
This Values() method creates and returns a slice of all elements in that list.

#### type Entity

```go
type Entity NCName
```

This is a reference to an unparsed entity with a name that matches the specified
name.

#### func (*Entity) Set

```go
func (me *Entity) Set(v string)
```
Since this is just a simple String type, this merely sets the current value from
the specified string.

#### func (Entity) String

```go
func (me Entity) String() string
```
Since this is just a simple String type, this merely returns its current string
value.

#### type Float

```go
type Float float32
```

Represents single-precision 32-bit floating-point numbers.

#### func (Float) N

```go
func (me Float) N() float32
```
Because littering your code with type conversions is a hassle...

#### func (*Float) Set

```go
func (me *Float) Set(s string)
```
Since this is a non-string scalar type, sets its current value obtained from
parsing the specified string.

#### func (Float) String

```go
func (me Float) String() string
```
Returns a string representation of its current non-string scalar value.

#### type GDay

```go
type GDay string
```

Represents a Gregorian day that recurs, specifically a day of the month such as
the fifth day of the month. A gDay is the space of a set of calendar dates.
Specifically, it is a set of one-day long, monthly periodic instances.

#### func (*GDay) Set

```go
func (me *GDay) Set(v string)
```
Since this is just a simple String type, this merely sets the current value from
the specified string.

#### func (GDay) String

```go
func (me GDay) String() string
```
Since this is just a simple String type, this merely returns its current string
value.

#### type GMonth

```go
type GMonth string
```

Represents a Gregorian month that recurs every year. A gMonth is the space of a
set of calendar months. Specifically, it is a set of one-month long, yearly
periodic instances.

#### func (*GMonth) Set

```go
func (me *GMonth) Set(v string)
```
Since this is just a simple String type, this merely sets the current value from
the specified string.

#### func (GMonth) String

```go
func (me GMonth) String() string
```
Since this is just a simple String type, this merely returns its current string
value.

#### type GMonthDay

```go
type GMonthDay string
```

Represents a specific Gregorian date that recurs, specifically a day of the year
such as the third of May. A gMonthDay is the set of calendar dates.
Specifically, it is a set of one-day long, annually periodic instances.

#### func (*GMonthDay) Set

```go
func (me *GMonthDay) Set(v string)
```
Since this is just a simple String type, this merely sets the current value from
the specified string.

#### func (GMonthDay) String

```go
func (me GMonthDay) String() string
```
Since this is just a simple String type, this merely returns its current string
value.

#### type GYear

```go
type GYear string
```

Represents a Gregorian year. A set of one-year long, nonperiodic instances.

#### func (*GYear) Set

```go
func (me *GYear) Set(v string)
```
Since this is just a simple String type, this merely sets the current value from
the specified string.

#### func (GYear) String

```go
func (me GYear) String() string
```
Since this is just a simple String type, this merely returns its current string
value.

#### type GYearMonth

```go
type GYearMonth string
```

Represents a specific Gregorian month in a specific Gregorian year. A set of
one-month long, nonperiodic instances.

#### func (*GYearMonth) Set

```go
func (me *GYearMonth) Set(v string)
```
Since this is just a simple String type, this merely sets the current value from
the specified string.

#### func (GYearMonth) String

```go
func (me GYearMonth) String() string
```
Since this is just a simple String type, this merely returns its current string
value.

#### type HexBinary

```go
type HexBinary string // []byte

```

Represents arbitrary hex-encoded binary data. A hexBinary is the set of
finite-length sequences of binary octets. Each binary octet is encoded as a
character tuple, consisting of two hexadecimal digits ([0-9a-fA-F]) representing
the octet code.

#### func (*HexBinary) Set

```go
func (me *HexBinary) Set(v string)
```
Since this is just a simple String type, this merely sets the current value from
the specified string.

#### func (HexBinary) String

```go
func (me HexBinary) String() string
```
Since this is just a simple String type, this merely returns its current string
value.

#### type Id

```go
type Id NCName
```

The ID must be a no-colon-name (NCName) and must be unique within an XML
document.

#### func (*Id) Set

```go
func (me *Id) Set(v string)
```
Since this is just a simple String type, this merely sets the current value from
the specified string.

#### func (Id) String

```go
func (me Id) String() string
```
Since this is just a simple String type, this merely returns its current string
value.

#### type Idref

```go
type Idref NCName
```

Represents a reference to an element that has an ID attribute that matches the
specified ID. An IDREF must be an NCName and must be a value of an element or
attribute of type ID within the XML document.

#### func (*Idref) Set

```go
func (me *Idref) Set(v string)
```
Since this is just a simple String type, this merely sets the current value from
the specified string.

#### func (Idref) String

```go
func (me Idref) String() string
```
Since this is just a simple String type, this merely returns its current string
value.

#### type Idrefs

```go
type Idrefs string
```

Contains a set of values of type IDREF.

#### func (*Idrefs) Set

```go
func (me *Idrefs) Set(v string)
```
Since this is just a simple String type, this merely sets the current value from
the specified string.

#### func (Idrefs) String

```go
func (me Idrefs) String() string
```
Since this is just a simple String type, this merely returns its current string
value.

#### func (Idrefs) Values

```go
func (me Idrefs) Values() (list []Idref)
```
This type declares a String containing a whitespace-separated list of values.
This Values() method creates and returns a slice of all elements in that list.

#### type Int

```go
type Int int32
```

Represents an integer with a minimum value of -2147483648 and maximum of
2147483647.

#### func (Int) N

```go
func (me Int) N() int32
```
Because littering your code with type conversions is a hassle...

#### func (*Int) Set

```go
func (me *Int) Set(s string)
```
Since this is a non-string scalar type, sets its current value obtained from
parsing the specified string.

#### func (Int) String

```go
func (me Int) String() string
```
Returns a string representation of its current non-string scalar value.

#### type Integer

```go
type Integer int64
```

Represents a sequence of decimal digits with an optional leading sign (+ or -).

#### func (Integer) N

```go
func (me Integer) N() int64
```
Because littering your code with type conversions is a hassle...

#### func (*Integer) Set

```go
func (me *Integer) Set(s string)
```
Since this is a non-string scalar type, sets its current value obtained from
parsing the specified string.

#### func (Integer) String

```go
func (me Integer) String() string
```
Returns a string representation of its current non-string scalar value.

#### type Language

```go
type Language Token
```

Represents natural language identifiers (defined by RFC 1766).

#### func (*Language) Set

```go
func (me *Language) Set(v string)
```
Since this is just a simple String type, this merely sets the current value from
the specified string.

#### func (Language) String

```go
func (me Language) String() string
```
Since this is just a simple String type, this merely returns its current string
value.

#### type Long

```go
type Long int64
```

Represents an integer with a minimum value of -9223372036854775808 and maximum
of 9223372036854775807.

#### func (Long) N

```go
func (me Long) N() int64
```
Because littering your code with type conversions is a hassle...

#### func (*Long) Set

```go
func (me *Long) Set(s string)
```
Since this is a non-string scalar type, sets its current value obtained from
parsing the specified string.

#### func (Long) String

```go
func (me Long) String() string
```
Returns a string representation of its current non-string scalar value.

#### type NCName

```go
type NCName Name
```

Represents noncolonized names. This data type is the same as Name, except it
cannot begin with a colon.

#### func (*NCName) Set

```go
func (me *NCName) Set(v string)
```
Since this is just a simple String type, this merely sets the current value from
the specified string.

#### func (NCName) String

```go
func (me NCName) String() string
```
Since this is just a simple String type, this merely returns its current string
value.

#### type Name

```go
type Name Token
```

Represents names in XML. A Name is a token that begins with a letter,
underscore, or colon and continues with name characters (letters, digits, and
other characters).

#### func (*Name) Set

```go
func (me *Name) Set(v string)
```
Since this is just a simple String type, this merely sets the current value from
the specified string.

#### func (Name) String

```go
func (me Name) String() string
```
Since this is just a simple String type, this merely returns its current string
value.

#### type NegativeInteger

```go
type NegativeInteger int64
```

Represents an integer that is less than zero. Consists of a negative sign (-)
and sequence of decimal digits.

#### func (NegativeInteger) N

```go
func (me NegativeInteger) N() int64
```
Because littering your code with type conversions is a hassle...

#### func (*NegativeInteger) Set

```go
func (me *NegativeInteger) Set(s string)
```
Since this is a non-string scalar type, sets its current value obtained from
parsing the specified string.

#### func (NegativeInteger) String

```go
func (me NegativeInteger) String() string
```
Returns a string representation of its current non-string scalar value.

#### type Nmtoken

```go
type Nmtoken Token
```

An NMTOKEN is set of name characters (letters, digits, and other characters) in
any combination. Unlike Name and NCName, NMTOKEN has no restrictions on the
starting character.

#### func (*Nmtoken) Set

```go
func (me *Nmtoken) Set(v string)
```
Since this is just a simple String type, this merely sets the current value from
the specified string.

#### func (Nmtoken) String

```go
func (me Nmtoken) String() string
```
Since this is just a simple String type, this merely returns its current string
value.

#### type Nmtokens

```go
type Nmtokens string
```

Contains a set of values of type NMTOKEN.

#### func (*Nmtokens) Set

```go
func (me *Nmtokens) Set(v string)
```
Since this is just a simple String type, this merely sets the current value from
the specified string.

#### func (Nmtokens) String

```go
func (me Nmtokens) String() string
```
Since this is just a simple String type, this merely returns its current string
value.

#### func (Nmtokens) Values

```go
func (me Nmtokens) Values() (list []Nmtoken)
```
This type declares a String containing a whitespace-separated list of values.
This Values() method creates and returns a slice of all elements in that list.

#### type NonNegativeInteger

```go
type NonNegativeInteger uint64
```

Represents an integer that is greater than or equal to zero.

#### func (NonNegativeInteger) N

```go
func (me NonNegativeInteger) N() uint64
```
Because littering your code with type conversions is a hassle...

#### func (*NonNegativeInteger) Set

```go
func (me *NonNegativeInteger) Set(s string)
```
Since this is a non-string scalar type, sets its current value obtained from
parsing the specified string.

#### func (NonNegativeInteger) String

```go
func (me NonNegativeInteger) String() string
```
Returns a string representation of its current non-string scalar value.

#### type NonPositiveInteger

```go
type NonPositiveInteger int64
```

Represents an integer that is less than or equal to zero. A
nonPositiveIntegerconsists of a negative sign (-) and sequence of decimal
digits.

#### func (NonPositiveInteger) N

```go
func (me NonPositiveInteger) N() int64
```
Because littering your code with type conversions is a hassle...

#### func (*NonPositiveInteger) Set

```go
func (me *NonPositiveInteger) Set(s string)
```
Since this is a non-string scalar type, sets its current value obtained from
parsing the specified string.

#### func (NonPositiveInteger) String

```go
func (me NonPositiveInteger) String() string
```
Returns a string representation of its current non-string scalar value.

#### type NormalizedString

```go
type NormalizedString String
```

Represents white space normalized strings.

#### func (*NormalizedString) Set

```go
func (me *NormalizedString) Set(v string)
```
Since this is just a simple String type, this merely sets the current value from
the specified string.

#### func (NormalizedString) String

```go
func (me NormalizedString) String() string
```
Since this is just a simple String type, this merely returns its current string
value.

#### type Notation

```go
type Notation string
```

A set of QNames.

#### func (*Notation) Set

```go
func (me *Notation) Set(v string)
```
Since this is just a simple String type, this merely sets the current value from
the specified string.

#### func (Notation) String

```go
func (me Notation) String() string
```
Since this is just a simple String type, this merely returns its current string
value.

#### func (Notation) Values

```go
func (me Notation) Values() (list []Qname)
```
This type declares a String containing a whitespace-separated list of values.
This Values() method creates and returns a slice of all elements in that list.

#### type Notations

```go
type Notations map[string]*notation
```


#### func (Notations) Add

```go
func (me Notations) Add(id, name, public, system string)
```

#### type PositiveInteger

```go
type PositiveInteger uint64
```

Represents an integer that is greater than zero.

#### func (PositiveInteger) N

```go
func (me PositiveInteger) N() uint64
```
Because littering your code with type conversions is a hassle...

#### func (*PositiveInteger) Set

```go
func (me *PositiveInteger) Set(s string)
```
Since this is a non-string scalar type, sets its current value obtained from
parsing the specified string.

#### func (PositiveInteger) String

```go
func (me PositiveInteger) String() string
```
Returns a string representation of its current non-string scalar value.

#### type Qname

```go
type Qname string
```

Represents a qualified name. A qualified name is composed of a prefix and a
local name separated by a colon. Both the prefix and local names must be an
NCName. The prefix must be associated with a namespace URI reference, using a
namespace declaration.

#### func (*Qname) Set

```go
func (me *Qname) Set(v string)
```
Since this is just a simple String type, this merely sets the current value from
the specified string.

#### func (Qname) String

```go
func (me Qname) String() string
```
Since this is just a simple String type, this merely returns its current string
value.

#### type Short

```go
type Short int16
```

Represents an integer with a minimum value of -32768 and maximum of 32767.

#### func (Short) N

```go
func (me Short) N() int16
```
Because littering your code with type conversions is a hassle...

#### func (*Short) Set

```go
func (me *Short) Set(s string)
```
Since this is a non-string scalar type, sets its current value obtained from
parsing the specified string.

#### func (Short) String

```go
func (me Short) String() string
```
Returns a string representation of its current non-string scalar value.

#### type String

```go
type String string
```

Represents character strings.

#### func (*String) Set

```go
func (me *String) Set(v string)
```
Since this is just a simple String type, this merely sets the current value from
the specified string.

#### func (String) String

```go
func (me String) String() string
```
Since this is just a simple String type, this merely returns its current string
value.

#### type Time

```go
type Time string // time.Time

```

Represents a specific instance of time.

#### func (*Time) Set

```go
func (me *Time) Set(v string)
```
Since this is just a simple String type, this merely sets the current value from
the specified string.

#### func (Time) String

```go
func (me Time) String() string
```
Since this is just a simple String type, this merely returns its current string
value.

#### type ToXsdtAnySimpleType

```go
type ToXsdtAnySimpleType interface {
	ToXsdtAnySimpleType() AnySimpleType
}
```

A convenience interface that declares a type conversion to AnySimpleType.

#### type ToXsdtAnyType

```go
type ToXsdtAnyType interface {
	ToXsdtAnyType() AnyType
}
```

A convenience interface that declares a type conversion to AnyType.

#### type ToXsdtAnyURI

```go
type ToXsdtAnyURI interface {
	ToXsdtAnyURI() AnyURI
}
```

A convenience interface that declares a type conversion to AnyURI.

#### type ToXsdtBase64Binary

```go
type ToXsdtBase64Binary interface {
	ToXsdtBase64Binary() Base64Binary
}
```

A convenience interface that declares a type conversion to Base64Binary.

#### type ToXsdtBoolean

```go
type ToXsdtBoolean interface {
	ToXsdtBoolean() Boolean
}
```

A convenience interface that declares a type conversion to Boolean.

#### type ToXsdtByte

```go
type ToXsdtByte interface {
	ToXsdtByte() Byte
}
```

A convenience interface that declares a type conversion to Byte.

#### type ToXsdtDate

```go
type ToXsdtDate interface {
	ToXsdtDate() Date
}
```

A convenience interface that declares a type conversion to Date.

#### type ToXsdtDateTime

```go
type ToXsdtDateTime interface {
	ToXsdtDateTime() DateTime
}
```

A convenience interface that declares a type conversion to DateTime.

#### type ToXsdtDecimal

```go
type ToXsdtDecimal interface {
	ToXsdtDecimal() Decimal
}
```

A convenience interface that declares a type conversion to Decimal.

#### type ToXsdtDouble

```go
type ToXsdtDouble interface {
	ToXsdtDouble() Double
}
```

A convenience interface that declares a type conversion to Double.

#### type ToXsdtDuration

```go
type ToXsdtDuration interface {
	ToXsdtDuration() Duration
}
```

A convenience interface that declares a type conversion to Duration.

#### type ToXsdtEntities

```go
type ToXsdtEntities interface {
	ToXsdtEntities() Entities
}
```

A convenience interface that declares a type conversion to Entities.

#### type ToXsdtEntity

```go
type ToXsdtEntity interface {
	ToXsdtEntity() Entity
}
```

A convenience interface that declares a type conversion to Entity.

#### type ToXsdtFloat

```go
type ToXsdtFloat interface {
	ToXsdtFloat() Float
}
```

A convenience interface that declares a type conversion to Float.

#### type ToXsdtGDay

```go
type ToXsdtGDay interface {
	ToXsdtGDay() GDay
}
```

A convenience interface that declares a type conversion to GDay.

#### type ToXsdtGMonth

```go
type ToXsdtGMonth interface {
	ToXsdtGMonth() GMonth
}
```

A convenience interface that declares a type conversion to GMonth.

#### type ToXsdtGMonthDay

```go
type ToXsdtGMonthDay interface {
	ToXsdtGMonthDay() GMonthDay
}
```

A convenience interface that declares a type conversion to GMonthDay.

#### type ToXsdtGYear

```go
type ToXsdtGYear interface {
	ToXsdtGYear() GYear
}
```

A convenience interface that declares a type conversion to GYear.

#### type ToXsdtGYearMonth

```go
type ToXsdtGYearMonth interface {
	ToXsdtGYearMonth() GYearMonth
}
```

A convenience interface that declares a type conversion to GYearMonth.

#### type ToXsdtHexBinary

```go
type ToXsdtHexBinary interface {
	ToXsdtHexBinary() HexBinary
}
```

A convenience interface that declares a type conversion to HexBinary.

#### type ToXsdtId

```go
type ToXsdtId interface {
	ToXsdtId() Id
}
```

A convenience interface that declares a type conversion to Id.

#### type ToXsdtIdref

```go
type ToXsdtIdref interface {
	ToXsdtIdref() Idref
}
```

A convenience interface that declares a type conversion to Idref.

#### type ToXsdtIdrefs

```go
type ToXsdtIdrefs interface {
	ToXsdtIdrefs() Idrefs
}
```

A convenience interface that declares a type conversion to Idrefs.

#### type ToXsdtInt

```go
type ToXsdtInt interface {
	ToXsdtInt() Int
}
```

A convenience interface that declares a type conversion to Int.

#### type ToXsdtInteger

```go
type ToXsdtInteger interface {
	ToXsdtInteger() Integer
}
```

A convenience interface that declares a type conversion to Integer.

#### type ToXsdtLanguage

```go
type ToXsdtLanguage interface {
	ToXsdtLanguage() Language
}
```

A convenience interface that declares a type conversion to Language.

#### type ToXsdtLong

```go
type ToXsdtLong interface {
	ToXsdtLong() Long
}
```

A convenience interface that declares a type conversion to Long.

#### type ToXsdtNCName

```go
type ToXsdtNCName interface {
	ToXsdtNCName() NCName
}
```

A convenience interface that declares a type conversion to NCName.

#### type ToXsdtName

```go
type ToXsdtName interface {
	ToXsdtName() Name
}
```

A convenience interface that declares a type conversion to Name.

#### type ToXsdtNegativeInteger

```go
type ToXsdtNegativeInteger interface {
	ToXsdtNegativeInteger() NegativeInteger
}
```

A convenience interface that declares a type conversion to NegativeInteger.

#### type ToXsdtNmtoken

```go
type ToXsdtNmtoken interface {
	ToXsdtNmtoken() Nmtoken
}
```

A convenience interface that declares a type conversion to Nmtoken.

#### type ToXsdtNmtokens

```go
type ToXsdtNmtokens interface {
	ToXsdtNmtokens() Nmtokens
}
```

A convenience interface that declares a type conversion to Nmtokens.

#### type ToXsdtNonNegativeInteger

```go
type ToXsdtNonNegativeInteger interface {
	ToXsdtNonNegativeInteger() NonNegativeInteger
}
```

A convenience interface that declares a type conversion to NonNegativeInteger.

#### type ToXsdtNonPositiveInteger

```go
type ToXsdtNonPositiveInteger interface {
	ToXsdtNonPositiveInteger() NonPositiveInteger
}
```

A convenience interface that declares a type conversion to NonPositiveInteger.

#### type ToXsdtNormalizedString

```go
type ToXsdtNormalizedString interface {
	ToXsdtNormalizedS() NormalizedString
}
```

A convenience interface that declares a type conversion to NormalizedString.

#### type ToXsdtNotation

```go
type ToXsdtNotation interface {
	ToXsdtNotation() Notation
}
```

A convenience interface that declares a type conversion to Notation.

#### type ToXsdtPositiveInteger

```go
type ToXsdtPositiveInteger interface {
	ToXsdtPositiveInteger() PositiveInteger
}
```

A convenience interface that declares a type conversion to PositiveInteger.

#### type ToXsdtQname

```go
type ToXsdtQname interface {
	ToXsdtQname() Qname
}
```

A convenience interface that declares a type conversion to Qname.

#### type ToXsdtShort

```go
type ToXsdtShort interface {
	ToXsdtShort() Short
}
```

A convenience interface that declares a type conversion to Short.

#### type ToXsdtString

```go
type ToXsdtString interface {
	ToXsdtString() String
}
```

A convenience interface that declares a type conversion to String.

#### type ToXsdtTime

```go
type ToXsdtTime interface {
	ToXsdtTime() Time
}
```

A convenience interface that declares a type conversion to Time.

#### type ToXsdtToken

```go
type ToXsdtToken interface {
	ToXsdtToken() Token
}
```

A convenience interface that declares a type conversion to Token.

#### type ToXsdtUnsignedByte

```go
type ToXsdtUnsignedByte interface {
	ToXsdtUnsignedByte() UnsignedByte
}
```

A convenience interface that declares a type conversion to UnsignedByte.

#### type ToXsdtUnsignedInt

```go
type ToXsdtUnsignedInt interface {
	ToXsdtUnsignedInt() UnsignedInt
}
```

A convenience interface that declares a type conversion to UnsignedInt.

#### type ToXsdtUnsignedLong

```go
type ToXsdtUnsignedLong interface {
	ToXsdtUnsignedLong() UnsignedLong
}
```

A convenience interface that declares a type conversion to UnsignedLong.

#### type ToXsdtUnsignedShort

```go
type ToXsdtUnsignedShort interface {
	ToXsdtUnsignedShort() UnsignedShort
}
```

A convenience interface that declares a type conversion to UnsignedShort.

#### type Token

```go
type Token NormalizedString
```

Represents tokenized strings.

#### func (*Token) Set

```go
func (me *Token) Set(v string)
```
Since this is just a simple String type, this merely sets the current value from
the specified string.

#### func (Token) String

```go
func (me Token) String() string
```
Since this is just a simple String type, this merely returns its current string
value.

#### type UnsignedByte

```go
type UnsignedByte uint8
```

Represents an integer with a minimum of zero and maximum of 255.

#### func (UnsignedByte) N

```go
func (me UnsignedByte) N() uint8
```
Because littering your code with type conversions is a hassle...

#### func (*UnsignedByte) Set

```go
func (me *UnsignedByte) Set(s string)
```
Since this is a non-string scalar type, sets its current value obtained from
parsing the specified string.

#### func (UnsignedByte) String

```go
func (me UnsignedByte) String() string
```
Returns a string representation of its current non-string scalar value.

#### type UnsignedInt

```go
type UnsignedInt uint32
```

Represents an integer with a minimum of zero and maximum of 4294967295.

#### func (UnsignedInt) N

```go
func (me UnsignedInt) N() uint32
```
Because littering your code with type conversions is a hassle...

#### func (*UnsignedInt) Set

```go
func (me *UnsignedInt) Set(s string)
```
Since this is a non-string scalar type, sets its current value obtained from
parsing the specified string.

#### func (UnsignedInt) String

```go
func (me UnsignedInt) String() string
```
Returns a string representation of its current non-string scalar value.

#### type UnsignedLong

```go
type UnsignedLong uint64
```

Represents an integer with a minimum of zero and maximum of
18446744073709551615.

#### func (UnsignedLong) N

```go
func (me UnsignedLong) N() uint64
```
Because littering your code with type conversions is a hassle...

#### func (*UnsignedLong) Set

```go
func (me *UnsignedLong) Set(s string)
```
Since this is a non-string scalar type, sets its current value obtained from
parsing the specified string.

#### func (UnsignedLong) String

```go
func (me UnsignedLong) String() string
```
Returns a string representation of its current non-string scalar value.

#### type UnsignedShort

```go
type UnsignedShort uint16
```

Represents an integer with a minimum of zero and maximum of 65535.

#### func (UnsignedShort) N

```go
func (me UnsignedShort) N() uint16
```
Because littering your code with type conversions is a hassle...

#### func (*UnsignedShort) Set

```go
func (me *UnsignedShort) Set(s string)
```
Since this is a non-string scalar type, sets its current value obtained from
parsing the specified string.

#### func (UnsignedShort) String

```go
func (me UnsignedShort) String() string
```
Returns a string representation of its current non-string scalar value.

--
**godocdown** http://github.com/robertkrimen/godocdown
