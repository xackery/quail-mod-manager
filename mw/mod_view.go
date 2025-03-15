package mw

import (
	"fmt"
	"quail-mod-manager/component"
)

func (mw *Mw) onModSelect() {

	fmt.Printf("Mod selected: %d\n", mw.modViewWlk.SelectedIndexes())
}

func SetModEntries(entries []*component.ModViewEntry) {
	mw := Instance()
	if mw == nil {
		fmt.Printf("mw is nil\n")
		return
	}

	mw.modEntries = entries
	mw.modView.SetItems(entries)
	if len(entries) > 0 {
		if mw.modViewWlk == nil {
			fmt.Printf("modViewWlk is nil\n")
			return
		}
		mw.modViewWlk.SetCurrentIndex(0)
		//onModViewSelect()
	}
	//handler.ModViewRefreshInvoke(entries)

}
