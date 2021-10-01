# lib-bpmn-engine

## Motivation

A BPMN engine, meant to be embedded in Go applications with minimum hurdles,
and a pleasant developer experience using it.
This approach can increase transparency for non-developers.

This library is meant to be embedded in your application and should not introduce more runtime-dependencies.
Hence, there's not DB-support built nor planned.
Also, the engine is not agnostic to any high availability approaches, like multiple instances or similar.

Think of an OpenAPI/Swagger spec can be served with your service - such could be done with a BPMN file.

### Philosophies around BPMN

The BPMN specification in its core is just about the rectangles, arrows, and how to interpret them.
With this foundation, it's an excellent opportunity to enrich transparency or communication or discussions 
about implementation details. So BPMN has a great potential to support me as a developer to not write
documentation into a wiki but rather expose the business process via well known symbols/graphics.

## Usage Example

See [example_test.go](./pkg/bpmn_engine/example_test.go)

## Current Implementation State

This is very early development.
A simple 'hello world' task can be executed.

Plenty of other BPMN elements left to be supported.
