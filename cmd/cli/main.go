package main

import (
	"log"
	"syscall"
	"unsafe"
)

func main() {
	var mod = syscall.NewLazyDLL("WinRing0x64.dll")
	var procInit = mod.NewProc("InitializeOls")
	var proc = mod.NewProc("Rdmsr")
	var low, high uint32
	var regAddr uint32 = 0x610
	procInit.Call()
	res, _, er := proc.Call(uintptr(regAddr), uintptr(unsafe.Pointer(&low)), uintptr(unsafe.Pointer(&high)))
	log.Print(res, er, low, high)
}
