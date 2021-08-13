//	A tiny package imported by all "go-xsd"-generated packages.
//
//	Maps all XSD built-in simple-types to Go types, which affords us easy mapping of any XSD type references in the schema to Go imports: every xs:string and xs:boolean automatically becomes xsdt.String and xsdt.Boolean etc.
//	Types are mapped to Go types depending on how encoding/xml.Unmarshal() can handle them: ie. it parses bools and numbers, but dates/durations have too many format mismatches and thus are just declared string types.
//	Same for base64- and hex-encoded binary data: since Unmarshal() won't decode them, we leave them as strings. If you need their binary data, your code needs to import Go's base64/hex codec packages and use them as necessary.
package xsdtypes
