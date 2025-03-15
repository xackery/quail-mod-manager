package component

import "github.com/xackery/wlk/walk"

type ModViewEntry struct {
	Icon    *walk.Bitmap
	Name    string
	Ext     string
	Size    string
	RawSize int
	checked bool
	URL     string
}
