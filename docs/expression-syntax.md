
## Expression Syntax

Cite from the BPMN 2.0 specification...
*BPMN does not itself provide a built-in model for describing structure of data or an Expression language for querying
that data. Instead, it formalizes hooks that allow for externally defined data structures and Expression languages.*

This lib-bpmn-engine uses [antonmedv/expr](https://github.com/antonmedv/expr) library for evaluate expression.

## Expression in exclusive gateways

Expressions used in exclusive gateways must evaluate to a single boolean value.
Examples for such expressions are listed below.

Some other engines use the equal sign (```=```) for these boolean expression.
The lib-bpmn-engine allows both, for compatibility reasons. This means, the result of 
```price > 10``` is equal to ```= price > 10```.

### Boolean expressions

| Operator                 | Description              | Example          |
|--------------------------|--------------------------|------------------|
| = (only one equals sign) | equal to                 | owner = "Paul"   |
| !=                       | not equal to             | owner != "Paul"  |
| <                        | less than                | totalPrice < 25  |
| <=                       | less than or equal to    | totalPrice <= 25 |
| >                        | greater than             | totalPrice > 25  |
| >=                       | greater than or equal to | totalPrice >= 25 |

## Variables

Variables can be provided to the engine, when a task is executed.
The library is type aware. E.g. in the examples below,
```owner``` must of type string and ```totalPrice``` of type int or float.

