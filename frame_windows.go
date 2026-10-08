//go:build windows

package main

import (
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32                       = windows.NewLazySystemDLL("user32.dll")
	procSendMessageW             = user32.NewProc("SendMessageW")
	procGetForegroundWindow      = user32.NewProc("GetForegroundWindow")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	procRedrawWindow             = user32.NewProc("RedrawWindow")
)

const (
	wmNCActivate  = 0x0086
	rdwInvalidate = 0x0001
	rdwUpdateNow  = 0x0100
	rdwFrame      = 0x0400
)

// refreshWindowFrame forza il ridisegno della barra del titolo dopo un cambio
// di tema: su Windows 10 DWM non aggiorna la cornice (i pulsanti riduci,
// ingrandisci e chiudi spariscono) finché la finestra non viene riattivata.
// Wails applica il tema in modo asincrono sul thread della UI, per questo
// si attende un attimo prima di ridisegnare.
func refreshWindowFrame() {
	go func() {
		time.Sleep(50 * time.Millisecond)

		// Il tema si cambia dal menu, quindi la nostra finestra è in primo piano.
		hwnd, _, _ := procGetForegroundWindow.Call()
		if hwnd == 0 {
			return
		}
		var pid uint32
		procGetWindowThreadProcessId.Call(hwnd, uintptrOf(&pid))
		if pid != windows.GetCurrentProcessId() {
			return
		}

		// Disattiva e riattiva l'area non client per farla ridisegnare.
		procSendMessageW.Call(hwnd, wmNCActivate, 0, 0)
		procSendMessageW.Call(hwnd, wmNCActivate, 1, 0)
		procRedrawWindow.Call(hwnd, 0, 0, rdwFrame|rdwInvalidate|rdwUpdateNow)
	}()
}

func uintptrOf(p *uint32) uintptr {
	return uintptr(unsafe.Pointer(p))
}
