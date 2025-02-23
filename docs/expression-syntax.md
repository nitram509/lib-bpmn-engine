
## Expression Syntax

Cite from the BPMN 2.0 specification...
*BPMN does not itself provide a built-in model for describing structure of data or an Expression language for querying
that data. Instead, it formalizes hooks that allow for externally defined data structures and Expression languages.*

This lib-bpmn-engine uses the [FEEL - Friendly Enough Expression Language](https://www.omg.org/spec/DMN) for evaluating expressions.
FEEL is part of the Decision Modeling Notation (DMN) and commonly used by BPMN engines.

## Expression in exclusive gateways

Expressions used in exclusive gateways must evaluate to a single boolean value.
The lib-bpmn-engine adheres to FEEL standards, which means expressions are simply writen such as ```price > 10```.
Hint: other engines allow expressions to start with `=` (equal sign) which, as said is not the case here.

## Variables

Variables can be provided to the engine, when a task is executed.
The library is type aware. E.g. in the examples below (boolean expressions),
```owner``` must of type string and ```totalPrice``` of type int or float.

### Boolean expressions

| Operator                 | Description              | Example          |
|--------------------------|--------------------------|------------------|
| = (just one equals sign) | equal to                 | owner = "Paul"   |
| !=                       | not equal to             | owner != "Paul"  |
| <                        | less than                | totalPrice < 25  |
| <=                       | less than or equal to    | totalPrice <= 25 |
| >                        | greater than             | totalPrice > 25  |
| >=                       | greater than or equal to | totalPrice >= 25 |

### Mathematical expressions

Basic mathematical operations are supported and can be used in conditional expressions.
E.g. if you define these variables and provide them to the context of a process instance,
then the expression ```sum >= foo + bar``` will evaluate to ```true```.
```go
    variables := map[string]interface{}{
        "foo": 3,
        "bar": 7,
        "sum": 10,
    }
    bpmnEngine.CreateAndRunInstance(key, variables)
```

## Supported data types

The package supports:

* Numbers - e.g. `103`, `2.5`
* Strings - e.g. `"hello"`
* Boolean - `true` and `false`
* Dates
* Time
* DateTime
* Days and time duration
* Years and months duration
* Ranges (or intervals)
* Lists - e.g. `[1, 2, 3]`

## Supported Operators

The package comes with a lot of operators:

### Arithmetic Operators

* `+` (addition)
* `-` (subtraction)
* `*` (multiplication)
* `/` (division)
* `%` (modulus)
* `**` (pow)

Example:

```FEEL
life + 42 // assuming life is of type number as well
``` 

### Comparison Operators

* `=` (equal)
* `!=` (not equal)
* `<` (less than)
* `>` (greater than)
* `<=` (less than or equal to)
* `>=` (greater than or equal to)

## Builtin string functions

* `"foo" + "bar"` -> result: "foobar"
* `string length(“foobar”)` -> result: 23
* `upper case(“foobar”)` -> result: “FOOBAR”
* `lower case(“FOOBAR”)` -> result: “foobar”
* `substring(“foobar”, 2, 2)` -> result: “oo”
* `contains(“foobar”, “foo”)` -> result: true
* `contains(“foobar”, “???”)` -> result: false
* `starts with ("foobar", "foo")` -> result: true
* `ends with ("foobar", "bar")` -> result: true
* `string join(["foo","bar"], "-")` -> result: "foo-bar"
* `string join(["foo","bar"])` -> result: "foobar"

### Ternary Operators

There is no operator but rather expression support, like so...

* `if true then “YES” else “NO”`

Example:

```FEEL
if user.Age > 30 then "mature" else "immature"
```

### Logical Operators

* `not`

Example:

```
not (true)
not (price > 20)
```

## Accessing structs (public properties)

Expression language can be used to access public properties in structs.
Given a `TestItem`, provided before the process instance runs, via variables...

```Go
	type TestItem struct {
		Key string
	}
	scope := Scope{
		"data": TestItem{Key: "foobar"},
	}
```

... then the following expression can be used to retrieve values from the struct:

```FEEL
get value( data, "Key" ) // returns "foobar"
```
