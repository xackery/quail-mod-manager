package handler

var (
	importModURLHandler []func()
	importModZipHandler []func()
)

// SubscribeNewArchive allows subscribing to new archve creation events
func ImportModURLSubscribe(fn func()) {
	importModURLHandler = append(importModURLHandler, fn)
}

// ImportModURLInvoke invokes new archive creation events
func OnImportModURL() {
	for _, fn := range importModURLHandler {
		fn()
	}
}

// SubscribeNewArchive allows subscribing to new archve creation events
func ImportModZipSubscribe(fn func()) {
	importModZipHandler = append(importModZipHandler, fn)
}

// ImportModZipInvoke invokes new archive creation events
func OnImportModZip() {
	for _, fn := range importModZipHandler {
		fn()
	}
}
