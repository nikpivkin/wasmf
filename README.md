This repository provides an example of adding [custom functions](https://www.openpolicyagent.org/docs/latest/extensions/#custom-built-in-functions-in-go) to OPA in Go using [exported WASM functions](https://developer.mozilla.org/en-US/docs/WebAssembly/Exported_functions). The goal is to allow Rego policies to use these custom functions without requiring users to update the OPA engine. 

This is accomplished by providing WASM modules containing the required functions. The provider compiles the WASM module with these functions, creates a configuration file with the function declarations and delivers this along with the policies to the users. 

The user only has to implement dynamic function registration based on this configuration file.
