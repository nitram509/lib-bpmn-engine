development hints and notes for lib-bpmn-engine
===============================================

### update Zeebe exporter protobuf

1. get new source file from https://github.com/camunda-community-hub/zeebe-exporter-protobuf/tree/master/src/main/proto
2. ensure you have latest ```protoc``` in your path installed
3. switch to folder pkg/bpmn_engine/exporter/zeebe
4. run ```protoc --go_opt=paths=source_relative --go_out=. --go_opt=Mschema.proto=zeebe/ schema.proto```
