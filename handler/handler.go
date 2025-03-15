package handler

var (
	importModHandler []func()
)

// SubscribeNewArchive allows subscribing to new archve creation events
func ImportModSubscribe(fn func()) {
	importModHandler = append(importModHandler, fn)
}

// ImportModInvoke invokes new archive creation events
func ImportModInvoke() {
	for _, fn := range importModHandler {
		fn()
	}
}
