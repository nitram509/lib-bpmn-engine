development hints and notes for lib-bpmn-engine
===============================================

### update Zeebe exporter protobuf

1. get new source file from https://github.com/camunda-community-hub/zeebe-exporter-protobuf/tree/master/src/main/proto
2. ensure you have latest ```protoc``` in your path installed
3. switch to folder pkg/bpmn_engine/exporter/zeebe
4. run ```protoc --go_opt=paths=source_relative --go_out=. --go_opt=Mschema.proto=zeebe/ schema.proto```

### update documentation

The documentation on Github pages is build via [MkDocs](https://www.mkdocs.org/).

#### local building and testing documentation 

1. ensure you have a Python 3.8+ environment installed
2. install MkDocs, according to their https://www.mkdocs.org/user-guide/installation/
    * shortcut: ```pip3 install -r doc-requirements.txt```
3. within this source repo, run ```mkdocs build``` to get a version of the HTML files

Alternatively, you could use a local test-server,
which eases the manual validation/verification of documentation updates. 

```shell
mkdocs serve
```

#### automated Github Pages update

There's a Github Action [update-gh-pages.yaml](./.github/workflows/update-gh-pages.yml),
which automatically will update the pages on every push to the main branch

### linting

From time to time, do some linting (would be better automatically checked via Github actions)
Using [go-critic](https://github.com/go-critic/go-critic)

```shell
gocritic check ./... 
```
