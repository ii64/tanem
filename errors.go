package emulator

import (
	"errors"
)

var (
	ErrNotImplemented        = errors.New("this is not implemented")

	ErrHeapLessEqualZero     = errors.New("heap map size was <= 0.")
	ErrMmapError             = errors.New("mmap error")
	ErrMapAddrNotMultiple    = errors.New("map addr was not multiple of page size")

	ErrUnknownBehavior       = errors.New("unknown behavior")
	
	ErrUnexpectedAsmLength   = errors.New("unexpected asm bytes length")
	ErrAsmFailed             = errors.New("asm failed")


	ErrJavaClassLoaded       = errors.New("java class already loaded")

	ErrFailConvertToInt      = errors.New("failed to parse binary")

	ErrELFReadFail           = errors.New("reader ELF fail")
	ErrELFReadNoDynamic      = errors.New("no dynamic in this ELF")
	ErrELF64NotSupported     = errors.New("64bit not supported now")
	ErrELFNHash              = errors.New("can not detect nsymbol by DT_HASH, DT_GNUHASH, not support now")
	ErrELFStSize             = errors.New("unknown handler for stsize")
	ErrELFSOFileTooLong      = errors.New("ELF SO filename is longer than 128")

	ErrELFSymbolNotFound     = errors.New("ELF Symbol not found")
)