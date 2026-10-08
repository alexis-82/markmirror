//go:build windows

package main

import (
	"testing"
	"unsafe"
)

// Le strutture passate a Windows devono avere la stessa dimensione di quelle C (x64).
func TestMenuBarStructLayout(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("layout verificato solo a 64 bit")
	}
	cases := []struct {
		name      string
		got, want uintptr
	}{
		{"MENUBARINFO", unsafe.Sizeof(menuBarInfo{}), 48},
		{"MENUITEMINFOW", unsafe.Sizeof(menuItemInfo{}), 80},
		{"DRAWITEMSTRUCT", unsafe.Sizeof(drawItemStruct{}), 64},
		{"UAHDRAWMENUITEM", unsafe.Sizeof(uahDrawMenuItem{}), 144},
		{"WINDOWPLACEMENT", unsafe.Sizeof(windowPlacementInfo{}), 44},
		{"UAHDRAWMENUITEM.umi", unsafe.Offsetof(uahDrawMenuItem{}.position), 88},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = %d, want %d", c.name, c.got, c.want)
		}
	}
}
