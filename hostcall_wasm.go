//go:build wasm

package pluginhelper

//go:wasmimport veloce host_call
func hostCall(opPtr, opLen, requestPtr, requestLen, responsePtr, responseCap uint32) uint64
