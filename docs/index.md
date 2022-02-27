# lib-bpmn-engine

## Motivation

A BPMN engine, meant to be embedded in Go applications with minimum hurdles,
and a pleasant developer experience using it.
This approach can increase transparency of code/implementation for non-developers.

This library is meant to be embedded in your application and should not introduce more runtime-dependencies.
Hence, there's no database support built nor planned.
Also, the engine is not agnostic to any high availability approaches, like multiple instances or similar.

See [Getting Started](./getting-started.md)


## Philosophies around BPMN

The BPMN specification in its core is a set of graphical symbols (rectangles, arrows, etc.)
and a standard definition about how to read/interpret them.
With this foundation, it's an excellent opportunity to enrich transparency or communication or discussions
about implementation details. So BPMN has a great potential to support me as a developer to not write
documentation into a wiki but rather expose the business process via well known symbols/graphics.

There's a conceptual similarity in usage between BPMN and OpenAPI/Swagger.
As developers, on the one hand side we often use OpenAPI/Swagger to document our endpoints, HTTP methods, and purpose
of the (HTTP) interface, our services offer. Hence, we enable others to use and integrate them.
With BPMN on the other hand it can be conceptual similar, when it comes to share internal behaviour of our services.
I see even larger similarity, when it comes to the question: *How do I maintain the documentation?*
Again, on the one hand side with OpenAPI/Swagger, we tend to either use reflection and code generators
or we follow the API spec first approach.
The later one is addressed by this library in the BPMN context: **Business Process spec first approach**


## Roadmap

#### ✔ v0.1.0

[progress milestone v0.1.0](///github.com/nitram509/lib-bpmn-engine/issues?q=is%3Aopen+is%3Aissue+milestone%3Av0.1.0)

For the first release I would like to have service tasks and events fully supported.


#### v0.2.0

[progress milestone v0.2.0](///github.com/nitram509/lib-bpmn-engine/issues?q=is%3Aopen+is%3Aissue+milestone%3Av0.2.0)

With basic element support, I would like to add visualization/monitoring capabilities.
If the idea of using Zeebe's exporter protocol is not too complex, that would be ideal.
If not, a simple console logger might do the job as well.


#### v0.3.0

[progress milestone v0.3.0](///github.com/nitram509/lib-bpmn-engine/issues?q=is%3Aopen+is%3Aissue+milestone%3Av0.3.0)

With basic element and visualization support, I would like to add expression language support as well as support for correlation keys
