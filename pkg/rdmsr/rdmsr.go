package rdmsr

import (
	"errors"
	"strconv"
	"syscall"
	"time"
	"unsafe"
)

var mod = syscall.NewLazyDLL("WinRing0x64.dll")
var procInit = mod.NewProc("InitializeOls")
var proc = mod.NewProc("Rdmsr")

type Msr struct {
	Low  uint32
	High uint32
}

func (msr *Msr) Value() uint64 {
	return (uint64(msr.High) << 32) | uint64(msr.Low)
}

func (msr *Msr) ToHex() (High string, Low string) {
	return strconv.FormatUint(uint64(msr.High), 16), strconv.FormatUint(uint64(msr.Low), 16)
}

func (msr *Msr) HexValue() string {
	return strconv.FormatUint(msr.Value(), 16)
}

func (msr *Msr) ToBinary() (High string, Low string) {
	return strconv.FormatUint(uint64(msr.High), 2), strconv.FormatUint(uint64(msr.Low), 2)
}

func (msr *Msr) BinaryValue() string {
	return strconv.FormatUint(msr.Value(), 2)
}

func (msr *Msr) ToBitfield(beg int, end int) uint64 {

	var val uint64 = msr.Value()
	var res uint64
	if end < beg {
		beg, end = end, beg
	}

	if end > 63 {
		end = 63
	}

	if beg < 0 {
		beg = 0
	}
	// https://go.dev/play/p/he2PjHzCwuR
	mask := val >> (end + 1)
	mask = mask << (end + 1)
	val = mask ^ val
	res = val >> beg

	return res
}

func Monitor(addr uint32, duration time.Duration) (<-chan Msr, <-chan error) {
	res := make(chan Msr, 1)
	err := make(chan error, 1)
	go func() {
		ticker := time.NewTicker(duration)
		for range ticker.C {
			msr, readError := ReadMsr(addr)
			if readError != nil {
				ticker.Stop()
				err <- readError
				return
			}
			res <- msr
		}
		close(res)
		close(err)
	}()

	return res, err
}

func ReadMsr(addr uint32) (Msr, error) {

	var Low, High uint32
	procInit.Call()
	r1, r2, _ := proc.Call(uintptr(addr), uintptr(unsafe.Pointer(&Low)), uintptr(unsafe.Pointer(&High)))

	if r1 != 1 {
		if uint32(r2) == addr {
			return Msr{}, errors.New("program not run as administrator")
		}
		if uint32(r2) == 0 {
			return Msr{}, errors.New("invalid MSR address")
		}
	}
	return Msr{Low, High}, nil
}
