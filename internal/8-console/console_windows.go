//go:build windows

package console

import (
	"bufio"
	"fmt"
	"os"
	"unsafe"

	ascii "twts/internal/8-result-ascii"

	"golang.org/x/sys/windows"
)

const (
	enableVirtualTerminalProcessing = 0x0004
)

var (
	kernel32           = windows.NewLazySystemDLL("kernel32.dll")
	procGetStdHandle   = kernel32.NewProc("GetStdHandle")
	procGetConsoleMode = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode = kernel32.NewProc("SetConsoleMode")
	procGetch          = windows.NewLazySystemDLL("msvcrt.dll").NewProc("_getch")
)

const stdOutputHandle = ^uintptr(10) // -11 as uintptr
const stdInputHandle = ^uintptr(9)   // -10 as uintptr

func init() {
	enableColors()
}

func enableColors() {
	handle, _, _ := procGetStdHandle.Call(stdOutputHandle)
	if handle == 0 {
		return
	}

	var mode uint32
	r1, _, _ := procGetConsoleMode.Call(handle, uintptr(unsafe.Pointer(&mode)))
	if r1 == 0 {
		return
	}

	procSetConsoleMode.Call(handle, uintptr(mode|enableVirtualTerminalProcessing))
}

func stdinIsConsole() bool {
	handle, _, _ := procGetStdHandle.Call(stdInputHandle)
	if handle == 0 {
		return false
	}
	var mode uint32
	r1, _, _ := procGetConsoleMode.Call(handle, uintptr(unsafe.Pointer(&mode)))
	return r1 != 0
}

// PrintError writes Inspector Gadget art and err in red, then waits for a key press before exiting.
func PrintError(err error) {
	fmt.Print(ColorRed)
	fmt.Print(ascii.InspectorGadget())
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	fmt.Print(ColorReset)
	if stdinIsConsole() {
		fmt.Print("Press any key to close...")
		waitForKey()
	}
	os.Exit(1)
}

// WaitAndExit waits for any key when stdin is a console, then exits with the given code.
func WaitAndExit(code int) {
	if stdinIsConsole() {
		fmt.Fprint(os.Stdout, "\nPress any key to close the window")
		waitForKey()
	}
	os.Exit(code)
}

func finishUsage(code int) {
	if stdinIsConsole() {
		fmt.Print("Press any key to close...")
		waitForKey()
	}
	os.Exit(code)
}

func waitForKey() {
	handle, _, _ := procGetStdHandle.Call(stdInputHandle)
	if handle == 0 {
		waitForKeyFallback()
		return
	}

	var mode uint32
	r1, _, _ := procGetConsoleMode.Call(handle, uintptr(unsafe.Pointer(&mode)))
	if r1 == 0 {
		waitForKeyFallback()
		return
	}

	procGetch.Call()
}

func waitForKeyFallback() {
	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadByte()
}
