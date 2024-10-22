# GoBPM

This is a Go written BPM engine.

> ⚠️ Warning: Right now this project is in early stages and is not suitable for any other than testing and development purposes.

## Getting started

For simple example process you can use [Showcase example process](test-cases/showcase-process.bpmn)

```bash
# 1. Clone the repository
git clone <FILL_REPO_URL>

# 2. Load the dependencies
cd go-bpms-engine
go mod download

# 3. Run the server
cd cmd
go run main.go #plus chosen flags see the usage bellow

# 4. Deploy process definition for example:
curl -X 'POST' \
  'http://localhost:8080/process-definitions' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/xml' \
  -d '<PLACE_PROCESS_DEFINITIONS_XML_HERE>'

# 5. Start a process instance
curl -X 'POST' \
  'http://localhost:8080/process-instances' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "processDefinitionKey": "<PLACE_PROCESS_DEFINITION_KEY>",
  "variables": {"price":100000}
}'
# 6. ...

```

### Usage

```bash
Usage of main:
  -path string
        DB Data path (default "/tmp/bpmn_engine/data")
  -port string
        port where to serve traffic (default "8080")
```

> For more check the [openapi/api.yaml](openapi/api.yaml)

> Or try [https://github.com/pbinitiative/go-bpms-showcase-fe](https://github.com/pbinitiative/go-bpms-showcase-fe)

## Links to used dependencies

- [https://github.com/nitram509/lib-bpmn-engine](https://github.com/nitram509/lib-bpmn-engine)
- [https://github.com/rqlite/rqlite](https://github.com/rqlite/rqlite)
