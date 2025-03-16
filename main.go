package main

import (
	"context"
	"fmt"
	"os"
	"quail-mod-manager/dialog"
	"quail-mod-manager/handler"
	"quail-mod-manager/mw"
	"quail-mod-manager/qmm"
	"runtime/debug"
	"time"
)

var Version string

func main() {

	err := run()
	if err != nil {
		dialog.ShowMessageBox("Error", "Failed to run: "+err.Error(), true)
		os.Exit(1)
	}

}

func run() error {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)

	handler.AboutSubscribe(onAbout)

	cmw, err := mw.New(ctx, cancel)
	if err != nil {
		return fmt.Errorf("mw new: %w", err)
	}
	_, err = qmm.New()
	if err != nil {
		return fmt.Errorf("qmm new: %w", err)
	}
	errCode := cmw.MainWindowWlk.Run()
	if errCode != 0 {
		return fmt.Errorf("run: %d", errCode)
	}
	return nil
}

func onAbout() {
	info, _ := debug.ReadBuildInfo()
	Version = info.Main.Version
	if Version == "" {
		Version = "dev-" + time.Now().Format("20060102")
	}
	dialog.ShowMessageBox("About", "Quail Mod Manager Version "+Version, false)
}
