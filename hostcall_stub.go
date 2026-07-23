//go:build !wasm

package pluginhelper

func hostCall(opPtr, opLen, requestPtr, requestLen, responsePtr, responseCap uint32) uint64 {
	return uint64(7) << 32
}
