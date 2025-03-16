package component

import "github.com/xackery/wlk/walk"

type ModViewEntry struct {
	IsEnabled bool
	ID        string
	Icon      *walk.Bitmap
	Name      string
	Version   string
	URL       string
}
