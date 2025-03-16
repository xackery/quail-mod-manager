package main

import (
	"context"
	"fmt"
	"os"
	"quail-mod-manager/dialog"
	"quail-mod-manager/mw"
	"quail-mod-manager/qmm"
)

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
