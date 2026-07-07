package widgets

import (
	"github.com/jupiterrider/purego-sdl3/sdl"
)

// Table arranges widgets in a grid: each row is a slice of widgets, one
// per column. Column widths are each the widest widget in that column
// across all rows (so e.g. a button column lines up at a uniform width),
// and row heights are each the tallest widget in that row; every cell's
// widget is vertically centered within its row. Call Layout after adding
// all rows to compute column widths and position everything — this can't
// happen incrementally since a column's width depends on every row.
type Table struct {
	X, Y                   float32
	ColSpacing, RowSpacing float32
	Rows                   [][]Widget

	colWidths     []float32
	contentWidth  float32
	contentHeight float32
}

// NewTable creates an empty Table anchored at (x, y).
func NewTable(x, y, colSpacing, rowSpacing float32) *Table {
	return &Table{X: x, Y: y, ColSpacing: colSpacing, RowSpacing: rowSpacing}
}

// AddRow appends a row of widgets, one per column. Rows may have fewer
// columns than others; missing trailing columns are simply left empty.
func (t *Table) AddRow(cells ...Widget) {
	t.Rows = append(t.Rows, cells)
}

// Layout computes each column's width (and each row's height) from the
// widgets currently in the table, then repositions every widget into its
// cell. Call it once after all rows have been added via AddRow.
func (t *Table) Layout() {
	numCols := 0
	for _, row := range t.Rows {
		if len(row) > numCols {
			numCols = len(row)
		}
	}

	colWidths := make([]float32, numCols)
	rowHeights := make([]float32, len(t.Rows))
	for ri, row := range t.Rows {
		for ci, w := range row {
			bounds := w.GetBounds()
			if bounds.W > colWidths[ci] {
				colWidths[ci] = bounds.W
			}
			if bounds.H > rowHeights[ri] {
				rowHeights[ri] = bounds.H
			}
		}
	}

	y := t.Y
	for ri, row := range t.Rows {
		x := t.X
		for ci, w := range row {
			bounds := w.GetBounds()
			cellW := bounds.W
			// Stretch to fill the column, except Labels: Label.Render
			// scales its text texture to fill Bounds, so widening it
			// would visually stretch/distort the text.
			if _, isLabel := w.(*Label); !isLabel {
				cellW = colWidths[ci]
			}
			cellY := y + (rowHeights[ri]-bounds.H)/2
			w.SetBounds(sdl.FRect{X: x, Y: cellY, W: cellW, H: bounds.H})
			x += colWidths[ci] + t.ColSpacing
		}
		y += rowHeights[ri] + t.RowSpacing
	}

	t.colWidths = colWidths
	t.contentHeight = y - t.RowSpacing - t.Y

	var totalW float32
	for i, w := range colWidths {
		totalW += w
		if i > 0 {
			totalW += t.ColSpacing
		}
	}
	t.contentWidth = totalW
}

// ColWidth returns the computed width of column i (valid after Layout).
func (t *Table) ColWidth(i int) float32 {
	if i < 0 || i >= len(t.colWidths) {
		return 0
	}
	return t.colWidths[i]
}

// ContentWidth returns the table's total width across all columns
// (valid after Layout).
func (t *Table) ContentWidth() float32 {
	return t.contentWidth
}

// ContentHeight returns the table's total height across all rows
// (valid after Layout).
func (t *Table) ContentHeight() float32 {
	return t.contentHeight
}

func (t *Table) Update(event sdl.Event, mx, my float32) bool {
	for _, row := range t.Rows {
		for _, w := range row {
			if w.Update(event, mx, my) {
				return true
			}
		}
	}
	return false
}

func (t *Table) Render(renderer *sdl.Renderer) {
	for _, row := range t.Rows {
		for _, w := range row {
			w.Render(renderer)
		}
	}
}

func (t *Table) Destroy() {
	for _, row := range t.Rows {
		for _, w := range row {
			if btn, ok := w.(*Button); ok {
				btn.Destroy()
			} else if lbl, ok := w.(*Label); ok {
				lbl.Destroy()
			} else if cb, ok := w.(*Checkbox); ok {
				cb.Destroy()
			}
		}
	}
}
