package component

import (
	"fmt"
	"sort"

	"log/slog"

	"github.com/xackery/wlk/walk"
)

var (
	fileView *FileView
)

type FileView struct {
	walk.TableModelBase
	walk.SorterBase
	sortColumn int
	sortOrder  walk.SortOrder
	items      []*FileViewEntry
}

func NewFileView() *FileView {
	m := new(FileView)
	m.ResetRows()
	return m
}

// Called by the TableView from SetModel and every time the model publishes a
// RowsReset event.
func (m *FileView) RowCount() int {
	return len(m.items)
}

// Called by the TableView when it needs the text to display for a given cell.
func (m *FileView) Value(row, col int) interface{} {
	item := m.items[row]

	switch col {
	case -1:
		return nil
	case 0:
		return item.Name
	case 1:
		return item.Ext
	case 2:
		return item.Size
	}

	slog.Info("invalid col: %d\n", col)
	return nil
}

// Called by the TableView to retrieve if a given row is checked.
func (m *FileView) Checked(row int) bool {
	return m.items[row].checked
}

// Called by the TableView when the user toggled the check box of a given row.
func (m *FileView) SetChecked(row int, checked bool) error {
	m.items[row].checked = checked

	return nil
}

// Called by the TableView to sort the model.
func (m *FileView) Sort(col int, order walk.SortOrder) error {
	m.sortColumn, m.sortOrder = col, order

	sort.SliceStable(m.items, func(i, j int) bool {
		a, b := m.items[i], m.items[j]

		c := func(ls bool) bool {
			if m.sortOrder == walk.SortAscending {
				return ls
			}

			return !ls
		}

		switch m.sortColumn {
		case -1:
			return false
		case 0:
			return c(a.Name < b.Name)
		case 1:
			return c(a.Ext < b.Ext)
		case 2:
			return c(a.RawSize < b.RawSize)
		}

		slog.Info("invalid sort col: %d", slog.Attr{Key: "column", Value: slog.IntValue(m.sortColumn)})
		return false
	})

	return m.SorterBase.Sort(col, order)
}

func (m *FileView) ResetRows() {
	m.items = nil

	m.PublishRowsReset()

	m.Sort(m.sortColumn, m.sortOrder)
}

func (m *FileView) SetItems(items []*FileViewEntry) {
	if m == nil {
		fmt.Printf("m is nil\n")
		return
	}
	m.items = items

	m.PublishRowsReset()

	m.Sort(m.sortColumn, m.sortOrder)
}
