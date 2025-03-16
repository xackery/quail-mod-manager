package mw

import (
	"context"
	"fmt"
	"os/exec"
	"quail-mod-manager/component"
	"quail-mod-manager/handler"
	"quail-mod-manager/ico"
	"sync"

	"github.com/xackery/wlk/cpl"
	"github.com/xackery/wlk/walk"
)

var (
	instance *Mw
	mux      = &sync.Mutex{}
)

// Mw is short for Main Window
type Mw struct {
	ctx           context.Context
	cancel        context.CancelCauseFunc
	MainWindow    cpl.MainWindow
	MainWindowWlk *walk.MainWindow
	modViewWlk    *walk.TableView
	progressWlk   *walk.ProgressBar
	modView       *component.ModView
	modEntries    []*component.ModViewEntry
}

func Instance() *Mw {
	mux.Lock()
	defer mux.Unlock()
	if instance == nil {
		instance = &Mw{}
	}
	return instance
}

func New(ctx context.Context, cancel context.CancelCauseFunc) (*Mw, error) {
	mux.Lock()
	defer mux.Unlock()

	//walk.SetDarkModeAllowed(true)

	mw := &Mw{
		ctx:    ctx,
		cancel: cancel,
	}
	mw.modView = component.NewModView()
	fvs := component.NewModViewStyler(mw.modView)

	mw.MainWindow = cpl.MainWindow{
		AssignTo: &mw.MainWindowWlk,
		Title:    "Quail Mod Manager",
		MinSize:  cpl.Size{Width: 500, Height: 320},
		Size:     cpl.Size{Width: 500, Height: 400},
		MenuItems: []cpl.MenuItem{
			cpl.Menu{
				Text: "&File",
				Items: []cpl.MenuItem{
					cpl.Action{
						Text: "&Import Mod URL",
						OnTriggered: func() {
							handler.OnImportModURL()
						},
						Shortcut: cpl.Shortcut{
							Key:       walk.KeyN,
							Modifiers: walk.ModControl,
						},
					},
					cpl.Action{
						Text: "&Import Mod via Zip",
						OnTriggered: func() {
							handler.OnImportModZip()
						},
						Shortcut: cpl.Shortcut{
							Key:       walk.KeyN,
							Modifiers: walk.ModControl,
						},
					},
					cpl.Action{
						Text: "E&xit",
						OnTriggered: func() {
							mw.MainWindowWlk.Close()
						},
						Shortcut: cpl.Shortcut{
							Key:       walk.KeyX,
							Modifiers: walk.ModControl,
						},
					},
				},
			},
		},
		ToolBar: cpl.ToolBar{
			Alignment: cpl.AlignHCenterVNear,
			Items: []cpl.MenuItem{
				cpl.Action{
					Image: ico.Grab("new"),
					OnTriggered: func() {
						handler.OnImportModURL()
					},
				},
				cpl.Action{
					Image: ico.Grab("material"),
					OnTriggered: func() {
						handler.OnGenerateMod()
					},
				},
			},
		},
		Layout: cpl.VBox{MarginsZero: true},
		Children: []cpl.Widget{
			cpl.Composite{
				Layout: cpl.Grid{Rows: 1},
				Children: []cpl.Widget{
					cpl.TableView{
						AssignTo:         &mw.modViewWlk,
						Name:             "tableView",
						CheckBoxes:       true,
						AlternatingRowBG: true,
						ColumnsOrderable: true,
						MultiSelection:   false,
						ContextMenuItems: []cpl.MenuItem{
							cpl.Action{
								Image: ico.Grab("obj"),
								Text:  "Open Github Page...",
								OnTriggered: func() {
									index := mw.modViewWlk.CurrentIndex()
									if index < 0 || index >= len(mw.modEntries) {
										return
									}

									entry := mw.modEntries[index]
									if entry == nil || entry.URL == "" {
										return
									}

									// Open the URL in the default browser
									cmd := exec.Command("cmd", "/c", "start", entry.URL)
									err := cmd.Run()
									if err != nil {
										walk.MsgBox(mw.MainWindowWlk, "Error", "Failed to open GitHub page: "+err.Error(), walk.MsgBoxIconError)
									}
								},
							},
							cpl.Action{
								Image: ico.Grab("delete"),
								Text:  "Remove",
								OnTriggered: func() {
									index := mw.modViewWlk.CurrentIndex()
									if index < 0 || index >= len(mw.modEntries) {
										return
									}
									entry := mw.modEntries[index]
									if entry == nil {
										return
									}
									handler.OnRemoveMod(entry.ID)
									fmt.Printf("Remove\n")
								},

								Shortcut: cpl.Shortcut{
									Key: walk.KeyDelete,
								},
							},
						},
						OnCurrentIndexChanged: mw.onModSelect,
						StyleCell:             fvs.StyleCell,
						//MaxSize:               cpl.Size{Width: 300, Height: 0},
						Columns: []cpl.TableViewColumn{
							{Name: "Name", Width: 160},
							{Name: "Ext", Width: 40},
							{Name: "Size", Width: 80},
						},
					},
				},
			},
		},
	}

	err := mw.MainWindow.Create()
	if err != nil {
		return nil, fmt.Errorf("create main window: %w", err)
	}

	mw.modViewWlk.SetModel(mw.modView)

	instance = mw

	return mw, nil
}
