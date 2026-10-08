//go:build windows

package main

import (
	"sync"
	"sync/atomic"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// Windows non offre una barra dei menu scura: in tema scuro la disegniamo noi
// intercettando i messaggi non documentati WM_UAHDRAWMENU/WM_UAHDRAWMENUITEM
// (la stessa tecnica usata da Notepad++). I menu a tendina diventano scuri
// tramite le API non documentate di uxtheme (SetPreferredAppMode).

var (
	gdi32 = windows.NewLazySystemDLL("gdi32.dll")

	procCallWindowProcW   = user32.NewProc("CallWindowProcW")
	procGetWindowLongPtrW = user32.NewProc("GetWindowLongPtrW")
	procSetWindowLongPtrW = user32.NewProc("SetWindowLongPtrW")
	procEnumWindows       = user32.NewProc("EnumWindows")
	procGetClassNameW     = user32.NewProc("GetClassNameW")
	procGetMenuBarInfo    = user32.NewProc("GetMenuBarInfo")
	procGetMenuItemInfoW  = user32.NewProc("GetMenuItemInfoW")
	procGetWindowRect     = user32.NewProc("GetWindowRect")
	procGetClientRect     = user32.NewProc("GetClientRect")
	procMapWindowPoints   = user32.NewProc("MapWindowPoints")
	procGetWindowDC       = user32.NewProc("GetWindowDC")
	procReleaseDC         = user32.NewProc("ReleaseDC")
	procFillRect          = user32.NewProc("FillRect")
	procDrawTextW         = user32.NewProc("DrawTextW")
	procDrawMenuBar       = user32.NewProc("DrawMenuBar")

	procCreateSolidBrush = gdi32.NewProc("CreateSolidBrush")
	procSetTextColor     = gdi32.NewProc("SetTextColor")
	procSetBkMode        = gdi32.NewProc("SetBkMode")
)

const (
	wmSettingChange    = 0x001A
	wmNCPaint          = 0x0085
	wmUAHDrawMenu      = 0x0091
	wmUAHDrawMenuItem  = 0x0092
	gwlpWndProc        = ^uintptr(3) // -4
	objIDMenu          = ^uintptr(2) // -3
	miimString         = 0x0040
	odsSelected        = 0x0001
	odsGrayed          = 0x0002
	odsDisabled        = 0x0004
	odsHotlight        = 0x0040
	odsInactive        = 0x0080
	odsNoAccel         = 0x0100
	dtCenter           = 0x0001
	dtVCenter          = 0x0004
	dtSingleLine       = 0x0020
	dtHidePrefix       = 0x00100000
	bkTransparent      = 1
	appModeAllowDark   = 1
	appModeForceDark   = 2
	appModeForceLight  = 3
	wailsWindowClass   = "wailsWindow"
	immersiveColorSet  = "ImmersiveColorSet"
	personalizeKeyPath = `Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`
)

// Colori del tema scuro (COLORREF 0x00BBGGRR), allineati a style.css.
const (
	colorBar      = 0x00262525 // --bg-chrome #252526
	colorHover    = 0x003a3a3a // --bg-hover
	colorText     = 0x00e0e0e0 // --fg
	colorTextGray = 0x00aea59d // --fg-muted #9da5ae
)

type rect struct{ left, top, right, bottom int32 }

type menuBarInfo struct {
	cbSize   uint32
	rcBar    rect
	hMenu    uintptr
	hwndMenu uintptr
	flags    uint32
}

type menuItemInfo struct {
	cbSize, fMask, fType, fState, wID                          uint32
	hSubMenu, hbmpChecked, hbmpUnchecked, dwItemData, typeData uintptr
	cch                                                        uint32
	hbmpItem                                                   uintptr
}

type drawItemStruct struct {
	ctlType, ctlID, itemID, itemAction, itemState uint32
	hwndItem, hDC                                 uintptr
	rcItem                                        rect
	itemData                                      uintptr
}

type uahMenu struct {
	hmenu, hdc uintptr
	flags      uint32
}

type uahDrawMenuItem struct {
	dis      drawItemStruct
	um       uahMenu
	position int32
	_        [13]uint32 // UAHMENUITEMMETRICS + UAHMENUPOPUPMETRICS
}

var (
	menuHwnd    uintptr
	oldWndProc  uintptr
	menuTheme   atomic.Value // "system", "light" o "dark"
	menuDark    atomic.Bool
	installOnce sync.Once

	brushBar, brushHover uintptr

	allowDarkModeForWindow, setPreferredAppMode, flushMenuThemes uintptr
)

// applyMenuTheme rende la barra e i menu a tendina chiari o scuri secondo il
// tema scelto ("system" segue l'impostazione di Windows).
func applyMenuTheme(theme string) {
	installOnce.Do(installMenuBarHook)
	if menuHwnd == 0 {
		return
	}
	menuTheme.Store(theme)
	updateMenuDark()
}

func updateMenuDark() {
	theme, _ := menuTheme.Load().(string)
	dark := theme == "dark" || (theme == "system" && systemUsesDarkTheme())
	menuDark.Store(dark)

	mode := uintptr(appModeAllowDark)
	switch theme {
	case "dark":
		mode = appModeForceDark
	case "light":
		mode = appModeForceLight
	}
	if setPreferredAppMode != 0 {
		syscall.SyscallN(setPreferredAppMode, mode)
	}
	if allowDarkModeForWindow != 0 {
		syscall.SyscallN(allowDarkModeForWindow, menuHwnd, boolArg(dark))
	}
	if flushMenuThemes != 0 {
		syscall.SyscallN(flushMenuThemes)
	}
	procDrawMenuBar.Call(menuHwnd)
	procRedrawWindow.Call(menuHwnd, 0, 0, rdwFrame|rdwInvalidate|rdwUpdateNow)
}

func installMenuBarHook() {
	if uxtheme, err := windows.LoadLibrary("uxtheme.dll"); err == nil {
		allowDarkModeForWindow, _ = windows.GetProcAddressByOrdinal(uxtheme, 133)
		setPreferredAppMode, _ = windows.GetProcAddressByOrdinal(uxtheme, 135)
		flushMenuThemes, _ = windows.GetProcAddressByOrdinal(uxtheme, 136)
	}

	hwnd := findMainWindow()
	if hwnd == 0 {
		return
	}
	brushBar, _, _ = procCreateSolidBrush.Call(colorBar)
	brushHover, _, _ = procCreateSolidBrush.Call(colorHover)
	// La procedura originale va salvata prima di installare la nostra, che
	// può ricevere messaggi subito dopo SetWindowLongPtrW.
	oldWndProc, _, _ = procGetWindowLongPtrW.Call(hwnd, gwlpWndProc)
	if oldWndProc == 0 {
		return
	}
	procSetWindowLongPtrW.Call(hwnd, gwlpWndProc, syscall.NewCallback(menuWndProc))
	menuHwnd = hwnd
}

// findMainWindow cerca la finestra Wails di questo processo.
func findMainWindow() uintptr {
	pid := windows.GetCurrentProcessId()
	var found uintptr
	cb := syscall.NewCallback(func(hwnd, _ uintptr) uintptr {
		var owner uint32
		procGetWindowThreadProcessId.Call(hwnd, uintptrOf(&owner))
		if owner != pid {
			return 1
		}
		var buf [64]uint16
		n, _, _ := procGetClassNameW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
		if windows.UTF16ToString(buf[:n]) == wailsWindowClass {
			found = hwnd
			return 0
		}
		return 1
	})
	procEnumWindows.Call(cb, 0)
	return found
}

func menuWndProc(hwnd, msg, wparam, lparam uintptr) uintptr {
	if menuDark.Load() {
		switch msg {
		case wmUAHDrawMenu:
			drawMenuBarBackground(hwnd, *(**uahMenu)(unsafe.Pointer(&lparam)))
			return 1
		case wmUAHDrawMenuItem:
			drawMenuBarItem(*(**uahDrawMenuItem)(unsafe.Pointer(&lparam)))
			return 1
		case wmNCPaint, wmNCActivate:
			ret, _, _ := procCallWindowProcW.Call(oldWndProc, hwnd, msg, wparam, lparam)
			drawMenuBarBottomLine(hwnd)
			return ret
		}
	}
	ret, _, _ := procCallWindowProcW.Call(oldWndProc, hwnd, msg, wparam, lparam)
	if msg == wmSettingChange && lparam != 0 &&
		windows.UTF16PtrToString(*(**uint16)(unsafe.Pointer(&lparam))) == immersiveColorSet {
		if theme, _ := menuTheme.Load().(string); theme == "system" {
			updateMenuDark()
		}
	}
	return ret
}

// menuBarRect restituisce il rettangolo della barra in coordinate della finestra.
func menuBarRect(hwnd uintptr) (rect, bool) {
	mbi := menuBarInfo{cbSize: uint32(unsafe.Sizeof(menuBarInfo{}))}
	ok, _, _ := procGetMenuBarInfo.Call(hwnd, objIDMenu, 0, uintptr(unsafe.Pointer(&mbi)))
	if ok == 0 {
		return rect{}, false
	}
	var win rect
	procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&win)))
	r := mbi.rcBar
	r.left -= win.left
	r.right -= win.left
	r.top -= win.top
	r.bottom -= win.top
	return r, true
}

func drawMenuBarBackground(hwnd uintptr, um *uahMenu) {
	if r, ok := menuBarRect(hwnd); ok {
		procFillRect.Call(um.hdc, uintptr(unsafe.Pointer(&r)), brushBar)
	}
}

func drawMenuBarItem(item *uahDrawMenuItem) {
	var text [256]uint16
	mii := menuItemInfo{
		fMask:    miimString,
		typeData: uintptr(unsafe.Pointer(&text[0])),
		cch:      uint32(len(text) - 1),
	}
	mii.cbSize = uint32(unsafe.Sizeof(mii))
	procGetMenuItemInfoW.Call(item.um.hmenu, uintptr(item.position), 1, uintptr(unsafe.Pointer(&mii)))

	state := item.dis.itemState
	brush := brushBar
	if state&(odsHotlight|odsSelected) != 0 {
		brush = brushHover
	}
	color := uintptr(colorText)
	if state&(odsInactive|odsGrayed|odsDisabled) != 0 {
		color = colorTextGray
	}
	flags := uintptr(dtCenter | dtSingleLine | dtVCenter)
	if state&odsNoAccel != 0 {
		flags |= dtHidePrefix
	}

	hdc := item.um.hdc
	r := item.dis.rcItem
	procFillRect.Call(hdc, uintptr(unsafe.Pointer(&r)), brush)
	procSetBkMode.Call(hdc, bkTransparent)
	procSetTextColor.Call(hdc, color)
	procDrawTextW.Call(hdc, uintptr(unsafe.Pointer(&text[0])), uintptr(mii.cch), uintptr(unsafe.Pointer(&r)), flags)
}

// drawMenuBarBottomLine copre la riga chiara che Windows disegna sotto la barra.
func drawMenuBarBottomLine(hwnd uintptr) {
	var client, win rect
	procGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&client)))
	procMapWindowPoints.Call(hwnd, 0, uintptr(unsafe.Pointer(&client)), 2)
	procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&win)))
	line := rect{
		left:   client.left - win.left,
		right:  client.right - win.left,
		top:    client.top - win.top - 1,
		bottom: client.top - win.top,
	}
	hdc, _, _ := procGetWindowDC.Call(hwnd)
	if hdc == 0 {
		return
	}
	procFillRect.Call(hdc, uintptr(unsafe.Pointer(&line)), brushBar)
	procReleaseDC.Call(hwnd, hdc)
}

func systemUsesDarkTheme() bool {
	key, err := registry.OpenKey(registry.CURRENT_USER, personalizeKeyPath, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer key.Close()
	light, _, err := key.GetIntegerValue("AppsUseLightTheme")
	return err == nil && light == 0
}

func boolArg(b bool) uintptr {
	if b {
		return 1
	}
	return 0
}
