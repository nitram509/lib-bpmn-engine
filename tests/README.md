
### Tests Folder

This folder contains mainly integration tests.
Hint: unit tests should remain in their respective packages (as usual).

Reasoning:
- avoid accidental use of private variables

### How to generate reference files?

See variable `enableJsonDataDump` in file `marshalling_test.go` and enable it,
to generate new JSON files. After that, you need to copy them manually into the desired folder.
