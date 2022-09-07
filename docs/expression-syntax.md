
## Expression Syntax

Cite from the BPMN 2.0 specification...
*BPMN does not itself provide a built-in model for describing structure of data or an Expression language for querying
that data. Instead, it formalizes hooks that allow for externally defined data structures and Expression languages.*

This lib-bpmn-engine uses [antonmedv/expr](https://github.com/antonmedv/expr) library for evaluate expression.

#### Noteworthy syntax differences

| Expression  | lib-bpmn-engine        | Camunda v8 (Zeebe)    | Comment                                                                             |
|-------------|------------------------|-----------------------|-------------------------------------------------------------------------------------|
| Comparisons | ```= foobar == true``` | ```= foobar = true``` | In Zeebe just use a single equals (=) sign, which causes trouble in lib-bpmn-engine |


## Expression in exclusive gateways

Expressions used in exclusive gateways must evaluate to a single boolean value.
Examples for such expressions are listed below.

Some other engines use the equal sign (```=```) for these boolean expression.
The lib-bpmn-engine allows both, for compatibility reasons. This means, the result of 
```price > 10``` is equal to ```= price > 10```.

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

## 'Expr' Language Definition

This is the full expr language specification, 
copied from this [antonmedv's source](https://github.com/antonmedv/expr/blob/master/docs/Language-Definition.md)

## Supported Literals

The package supports:

* **strings** - single and double quotes (e.g. `"hello"`, `'hello'`)
* **numbers** - e.g. `103`, `2.5`, `.5`
* **arrays** - e.g. `[1, 2, 3]`
* **maps** - e.g. `{foo: "bar"}`
* **booleans** - `true` and `false`
* **nil** - `nil`

## Digit separators

Integer literals may contain digit separators to allow digit grouping into more legible forms.

Example:

```
10_000_000_000
```

## Accessing Public Properties

Public properties on structs can be accessed by using the `.` syntax.
If you pass an array into an expression, use the `[]` syntax to access array keys.

```js
foo.Array[0].Value
```

## Functions and Methods

Functions may be called using `()` syntax. The `.` syntax can also be used to call methods on an struct.

```js
price.String()
```

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

```js
life + universe + everything
``` 

### Comparison Operators

* `==` (equal)
* `!=` (not equal)
* `<` (less than)
* `>` (greater than)
* `<=` (less than or equal to)
* `>=` (greater than or equal to)

### Logical Operators

* `not` or `!`
* `and` or `&&`
* `or` or `||`

Example:

```
life < universe || life < everything
```

### String Operators

* `+` (concatenation)
* `matches` (regex match)
* `contains` (string contains)
* `startsWith` (has prefix)
* `endsWith` (has suffix)

To test if a string does *not* match a regex, use the logical `not` operator in combination with the `matches` operator:

```js
not ("foo" matches "^b.+")
```

You must use parenthesis because the unary operator `not` has precedence over the binary operator `matches`.

Example:

```js
'Arthur' + ' ' + 'Dent'
```

Result will be set to `Arthur Dent`.

### Membership Operators

* `in` (contain)
* `not in` (does not contain)

Example:

```js
user.Group in ["human_resources", "marketing"]
```

```js
"foo" in {foo: 1, bar: 2}
```

### Numeric Operators

* `..` (range)

Example:

```js
user.Age in 18..45
```

The range is inclusive:

```js
1..3 == [1, 2, 3]
```

### Ternary Operators

* `foo ? 'yes' : 'no'`

Example:

```js
user.Age > 30 ? "mature" : "immature"
```

## Builtin functions

* `len` (length of array, map or string)
* `all` (will return `true` if all element satisfies the predicate)
* `none` (will return `true` if all element does NOT satisfies the predicate)
* `any` (will return `true` if any element satisfies the predicate)
* `one` (will return `true` if exactly ONE element satisfies the predicate)
* `filter` (filter array by the predicate)
* `map` (map all items with the closure)
* `count` (returns number of elements what satisfies the predicate)

Examples:

Ensure all tweets are less than 280 chars.

```js
all(Tweets, {.Size < 280})
```

Ensure there is exactly one winner.

```js
one(Participants, {.Winner})
```

## Closures

* `{...}` (closure)

Closures allowed only with builtin functions. To access current item use `#` symbol.

```js
map(0..9, {# / 2})
```

If the item of array is struct, it's possible to access fields of struct with omitted `#` symbol (`#.Value` becomes `.Value`).

```js
filter(Tweets, {len(.Value) > 280})
```

## Slices

* `array[:]` (slice)

Slices can work with arrays or strings.

Example:

Variable `array` is `[1,2,3,4,5]`.

```js
array[1:5] == [2,3,4] 
array[3:] == [4,5]
array[:4] == [1,2,3]
array[:] == array
```
