package handler

var (
	importModURLHandler []func()
	importModZipHandler []func()
	removeModHandler    []func(modID string)
	generateModHandler  []func()
	enableModHandler    []func(modID string, state bool)
	aboutHandler        []func()
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

// SubscribeNewArchive allows subscribing to new archve creation events
func RemoveModSubscribe(fn func(modID string)) {
	removeModHandler = append(removeModHandler, fn)
}

// ImportModZipInvoke invokes new archive creation events
func OnRemoveMod(modID string) {
	for _, fn := range removeModHandler {
		fn(modID)
	}
}

func GenerateModSubscribe(fn func()) {
	generateModHandler = append(generateModHandler, fn)
}

func OnGenerateMod() {
	for _, fn := range generateModHandler {
		fn()
	}
}

func EnableModSubscribe(fn func(modID string, state bool)) {
	enableModHandler = append(enableModHandler, fn)
}

func OnModEnabled(modID string, state bool) {
	for _, fn := range enableModHandler {
		fn(modID, state)
	}
}

func AboutSubscribe(fn func()) {
	aboutHandler = append(aboutHandler, fn)
}

func OnAbout() {
	for _, fn := range aboutHandler {
		fn()
	}
}
