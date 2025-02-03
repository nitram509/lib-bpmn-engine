# ZenBPM

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

### Docker

You can run this in a docker container with the following command:

```bash
docker build -t gobpms-engine .
docker run -p 8080:8080 -p 4001:4001 gobpms-engine
```

## Links to used dependencies

- [https://github.com/pbinitiative/zenbpm](https://github.com/pbinitiative/zenbpm)
- [https://github.com/rqlite/rqlite](https://github.com/rqlite/rqlite)

## Local development

Setting up the cluster locally

```bash
go run cmd/main.go  -port 8090 -node-id 0 -join localhost:4000,localhost:4001,localhost:4002 -http-addr=localhost:8080 -raft-addr=localhost:4000 -bootstrap-expect 3 -path /tmp/bpmn_engine/data-0
go run cmd/main.go  -port 8091 -node-id 1 -join localhost:4000,localhost:4001,localhost:4002 -http-addr=localhost:8081 -raft-addr=localhost:4001 -bootstrap-expect 3 -path /tmp/bpmn_engine/data-1
go run cmd/main.go  -port 8092 -node-id 2 -join localhost:4000,localhost:4001,localhost:4002 -http-addr=localhost:8082 -raft-addr=localhost:4002 -bootstrap-expect 3 -path /tmp/bpmn_engine/data-2
```
