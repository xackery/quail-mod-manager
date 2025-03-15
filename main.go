package main

import (
	"context"
	"fmt"
	"os"
	"quail-mod-manager/component"
	"quail-mod-manager/ico"
	"quail-mod-manager/mw"
	"quail-mod-manager/qmm"
)

func main() {

	err := run()
	if err != nil {
		fmt.Println("Failed to run: ", err)
		os.Exit(1)
	}

}

func run() error {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)

	err := qmm.New()
	if err != nil {
		return fmt.Errorf("qmm new: %w", err)
	}

	cmw, err := mw.New(ctx, cancel)
	if err != nil {
		return fmt.Errorf("mw new: %w", err)
	}

	mw.SetModEntries([]*component.ModViewEntry{
		{Icon: ico.Grab("header"), Name: "Quail", URL: "https://github.com/xackery/quail"},
		{Icon: ico.Grab("bon"), Name: "eqgzi", URL: "https://github.com/xackery/eqgzi"},
	})
	errCode := cmw.MainWindowWlk.Run()
	if errCode != 0 {
		return fmt.Errorf("run: %d", errCode)
	}
	return nil
}
