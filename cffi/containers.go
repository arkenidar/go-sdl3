package main

/*
#include <stdint.h>
*/
import "C"

import (
	"errors"
	"strings"
	"unsafe"

	"arkenidar.com/purego-sdl3/internal/widgets"
)

//export gui_hlayout_new
func gui_hlayout_new(x, y, spacing float32, out *uint64) int32 {
	return guard(func() error {
		layout := widgets.NewLayout(x, y, spacing)
		if out != nil {
			*out = handles.put(layout)
		}
		return nil
	})
}

//export gui_hlayout_add
func gui_hlayout_add(layoutHandle, widgetHandle uint64) int32 {
	return guard(func() error {
		v, ok := handles.get(layoutHandle)
		if !ok {
			return errors.New("unknown layout handle")
		}
		layout, ok := v.(*widgets.Layout)
		if !ok {
			return errors.New("handle is not a Layout")
		}
		w, err := lookupWidget(widgetHandle)
		if err != nil {
			return err
		}
		layout.AddWidget(w)
		return nil
	})
}

//export gui_table_new
func gui_table_new(x, y, colSpacing, rowSpacing float32, out *uint64) int32 {
	return guard(func() error {
		table := widgets.NewTable(x, y, colSpacing, rowSpacing)
		if out != nil {
			*out = handles.put(table)
		}
		return nil
	})
}

// gui_table_add_row appends one row: widgetHandles points to an array of n
// uint64 widget handles, one per column.
//
//export gui_table_add_row
func gui_table_add_row(tableHandle uint64, widgetHandles *uint64, n int32) int32 {
	return guard(func() error {
		v, ok := handles.get(tableHandle)
		if !ok {
			return errors.New("unknown table handle")
		}
		table, ok := v.(*widgets.Table)
		if !ok {
			return errors.New("handle is not a Table")
		}
		if n < 0 {
			return errors.New("negative row length")
		}
		raw := unsafe.Slice(widgetHandles, n)
		cells := make([]widgets.Widget, n)
		for i, h := range raw {
			w, err := lookupWidget(h)
			if err != nil {
				return err
			}
			cells[i] = w
		}
		table.AddRow(cells...)
		return nil
	})
}

//export gui_table_layout
func gui_table_layout(handle uint64) int32 {
	return guard(func() error {
		table, err := lookupTable(handle)
		if err != nil {
			return err
		}
		table.Layout()
		return nil
	})
}

//export gui_table_col_width
func gui_table_col_width(handle uint64, col int32, out *float32) int32 {
	return guard(func() error {
		table, err := lookupTable(handle)
		if err != nil {
			return err
		}
		if out != nil {
			*out = table.ColWidth(int(col))
		}
		return nil
	})
}

//export gui_table_content_size
func gui_table_content_size(handle uint64, outW, outH *float32) int32 {
	return guard(func() error {
		table, err := lookupTable(handle)
		if err != nil {
			return err
		}
		if outW != nil {
			*outW = table.ContentWidth()
		}
		if outH != nil {
			*outH = table.ContentHeight()
		}
		return nil
	})
}

func lookupTable(h uint64) (*widgets.Table, error) {
	v, ok := handles.get(h)
	if !ok {
		return nil, errors.New("unknown table handle")
	}
	table, ok := v.(*widgets.Table)
	if !ok {
		return nil, errors.New("handle is not a Table")
	}
	return table, nil
}

//export gui_textarea_new
func gui_textarea_new(renderer, font, window uint64, x, y, w, h float32, out *uint64) int32 {
	return guard(func() error {
		rend, err := lookupRenderer(renderer)
		if err != nil {
			return err
		}
		f, err := lookupFont(font)
		if err != nil {
			return err
		}
		win, err := lookupWindow(window)
		if err != nil {
			return err
		}
		ta := widgets.NewTextArea(x, y, w, h, f, rend, win)
		if out != nil {
			*out = handles.put(ta)
		}
		return nil
	})
}

//export gui_textarea_get_text
func gui_textarea_get_text(handle uint64) *C.char {
	v, ok := handles.get(handle)
	if !ok {
		return C.CString("")
	}
	ta, ok := v.(*widgets.TextArea)
	if !ok {
		return C.CString("")
	}
	return C.CString(strings.Join(ta.Lines, "\n"))
}

//export gui_textarea_set_text
func gui_textarea_set_text(handle uint64, text *C.char) int32 {
	return guard(func() error {
		v, ok := handles.get(handle)
		if !ok {
			return errors.New("unknown widget handle")
		}
		ta, ok := v.(*widgets.TextArea)
		if !ok {
			return errors.New("handle is not a TextArea")
		}
		ta.Lines = strings.Split(C.GoString(text), "\n")
		if len(ta.Lines) == 0 {
			ta.Lines = []string{""}
		}
		return nil
	})
}

//export gui_textarea_set_wordwrap
func gui_textarea_set_wordwrap(handle uint64, wordWrap int32) int32 {
	return guard(func() error {
		v, ok := handles.get(handle)
		if !ok {
			return errors.New("unknown widget handle")
		}
		ta, ok := v.(*widgets.TextArea)
		if !ok {
			return errors.New("handle is not a TextArea")
		}
		if (wordWrap != 0) != ta.WordWrap {
			ta.ToggleWordWrap()
		}
		return nil
	})
}

//export gui_textarea_focus
func gui_textarea_focus(handle uint64) int32 {
	return guard(func() error {
		v, ok := handles.get(handle)
		if !ok {
			return errors.New("unknown widget handle")
		}
		ta, ok := v.(*widgets.TextArea)
		if !ok {
			return errors.New("handle is not a TextArea")
		}
		ta.Focus()
		return nil
	})
}

//export gui_textarea_blur
func gui_textarea_blur(handle uint64) int32 {
	return guard(func() error {
		v, ok := handles.get(handle)
		if !ok {
			return errors.New("unknown widget handle")
		}
		ta, ok := v.(*widgets.TextArea)
		if !ok {
			return errors.New("handle is not a TextArea")
		}
		ta.Blur()
		return nil
	})
}
